"""Tool registry for managing and executing tools."""

import asyncio
from typing import Dict, List, Optional, Any
import structlog

from .base import BaseTool, ToolExecutionResult, ToolCapability

logger = structlog.get_logger(__name__)


class ToolRegistry:
    """Registry for managing and executing tools."""
    
    def __init__(self):
        """Initialize the tool registry."""
        self._tools: Dict[str, BaseTool] = {}
        self._initialized = False
        self.logger = structlog.get_logger("tool_registry")
    
    async def initialize(self) -> bool:
        """Initialize all registered tools.
        
        Returns:
            True if all tools initialized successfully, False otherwise
        """
        if self._initialized:
            return True
        
        self.logger.info("Initializing tool registry", tool_count=len(self._tools))
        
        failed_tools = []
        for tool_name, tool in self._tools.items():
            try:
                success = await tool.initialize()
                if not success:
                    failed_tools.append(tool_name)
                    self.logger.warning("Tool initialization failed", tool=tool_name)
                else:
                    self.logger.info("Tool initialized successfully", tool=tool_name)
            except Exception as e:
                failed_tools.append(tool_name)
                self.logger.error("Tool initialization error", tool=tool_name, error=str(e))
        
        if failed_tools:
            self.logger.warning("Some tools failed to initialize", failed_tools=failed_tools)
            return False
        
        self._initialized = True
        self.logger.info("Tool registry initialized successfully")
        return True
    
    async def register_tool(self, tool: BaseTool) -> bool:
        """Register a tool in the registry.
        
        Args:
            tool: The tool to register
            
        Returns:
            True if registration was successful, False otherwise
        """
        try:
            if tool.name in self._tools:
                self.logger.warning("Tool already registered, replacing", tool=tool.name)
            
            self._tools[tool.name] = tool
            self.logger.info("Tool registered", tool=tool.name, description=tool.description)
            
            # If registry is already initialized, initialize this tool
            if self._initialized:
                success = await tool.initialize()
                if not success:
                    self.logger.error("Failed to initialize newly registered tool", tool=tool.name)
                    return False
            
            return True
            
        except Exception as e:
            self.logger.error("Error registering tool", tool=tool.name, error=str(e))
            return False
    
    async def unregister_tool(self, tool_name: str) -> bool:
        """Unregister a tool from the registry.
        
        Args:
            tool_name: Name of the tool to unregister
            
        Returns:
            True if unregistration was successful, False otherwise
        """
        try:
            if tool_name not in self._tools:
                self.logger.warning("Tool not found for unregistration", tool=tool_name)
                return False
            
            tool = self._tools[tool_name]
            await tool.cleanup()
            del self._tools[tool_name]
            
            self.logger.info("Tool unregistered", tool=tool_name)
            return True
            
        except Exception as e:
            self.logger.error("Error unregistering tool", tool=tool_name, error=str(e))
            return False
    
    async def execute_tool(
        self,
        tool_name: str,
        parameters: Dict[str, Any],
        timeout: Optional[int] = None
    ) -> ToolExecutionResult:
        """Execute a tool with given parameters.
        
        Args:
            tool_name: Name of the tool to execute
            parameters: Parameters for the tool execution
            timeout: Optional timeout in seconds
            
        Returns:
            ToolExecutionResult containing the execution result
        """
        try:
            # Check if tool exists
            if tool_name not in self._tools:
                return ToolExecutionResult(
                    success=False,
                    error=f"Tool '{tool_name}' not found in registry",
                    execution_time=0.0,
                    metadata={"available_tools": list(self._tools.keys())}
                )
            
            tool = self._tools[tool_name]
            
            # Check if tool is healthy
            if not await tool.health_check():
                return ToolExecutionResult(
                    success=False,
                    error=f"Tool '{tool_name}' is not healthy",
                    execution_time=0.0,
                    metadata={"tool": tool_name}
                )
            
            # Validate parameters
            if not await tool.validate_parameters(parameters):
                return ToolExecutionResult(
                    success=False,
                    error=f"Invalid parameters for tool '{tool_name}'",
                    execution_time=0.0,
                    metadata={"tool": tool_name, "parameters": parameters}
                )
            
            self.logger.info("Executing tool", tool=tool_name, parameters=parameters)
            
            # Execute with optional timeout
            if timeout:
                try:
                    result = await asyncio.wait_for(
                        tool.execute(parameters),
                        timeout=timeout
                    )
                except asyncio.TimeoutError:
                    return ToolExecutionResult(
                        success=False,
                        error=f"Tool execution timed out after {timeout} seconds",
                        execution_time=timeout,
                        metadata={"tool": tool_name, "timeout": timeout}
                    )
            else:
                result = await tool.execute(parameters)
            
            self.logger.info(
                "Tool execution completed",
                tool=tool_name,
                success=result.success,
                execution_time=result.execution_time
            )
            
            return result
            
        except Exception as e:
            self.logger.error("Error executing tool", tool=tool_name, error=str(e))
            return ToolExecutionResult(
                success=False,
                error=f"Unexpected error executing tool '{tool_name}': {str(e)}",
                execution_time=0.0,
                metadata={"tool": tool_name, "exception": str(e)}
            )
    
    def list_tools(self) -> List[str]:
        """Get list of registered tool names.
        
        Returns:
            List of tool names
        """
        return list(self._tools.keys())
    
    def get_tool(self, tool_name: str) -> Optional[BaseTool]:
        """Get a tool by name.
        
        Args:
            tool_name: Name of the tool
            
        Returns:
            The tool instance or None if not found
        """
        return self._tools.get(tool_name)
    
    def get_tool_capabilities(self, tool_name: Optional[str] = None) -> Dict[str, List[ToolCapability]]:
        """Get capabilities of tools.
        
        Args:
            tool_name: Specific tool name, or None for all tools
            
        Returns:
            Dictionary mapping tool names to their capabilities
        """
        if tool_name:
            if tool_name in self._tools:
                return {tool_name: self._tools[tool_name].get_capabilities()}
            else:
                return {}
        
        capabilities = {}
        for name, tool in self._tools.items():
            try:
                capabilities[name] = tool.get_capabilities()
            except Exception as e:
                self.logger.error("Error getting capabilities", tool=name, error=str(e))
                capabilities[name] = []
        
        return capabilities
    
    async def health_check(self) -> Dict[str, bool]:
        """Check health of all registered tools.
        
        Returns:
            Dictionary mapping tool names to their health status
        """
        health_status = {}
        
        for tool_name, tool in self._tools.items():
            try:
                health_status[tool_name] = await tool.health_check()
            except Exception as e:
                self.logger.error("Error checking tool health", tool=tool_name, error=str(e))
                health_status[tool_name] = False
        
        return health_status
    
    async def cleanup(self) -> None:
        """Clean up all tools and the registry."""
        self.logger.info("Cleaning up tool registry")
        
        for tool_name, tool in self._tools.items():
            try:
                await tool.cleanup()
                self.logger.info("Tool cleaned up", tool=tool_name)
            except Exception as e:
                self.logger.error("Error cleaning up tool", tool=tool_name, error=str(e))
        
        self._tools.clear()
        self._initialized = False
        self.logger.info("Tool registry cleanup completed")