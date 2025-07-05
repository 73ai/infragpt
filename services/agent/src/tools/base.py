"""Base tool implementation for Backend Agent Service."""

import asyncio
import time
from abc import ABC, abstractmethod
from datetime import datetime
from typing import Any, Dict, List, Optional
from pydantic import BaseModel, Field
import structlog

logger = structlog.get_logger(__name__)


class ToolExecutionResult(BaseModel):
    """Result of a tool execution."""
    success: bool = Field(description="Whether the tool execution was successful")
    output: Optional[str] = Field(default=None, description="Output from the tool")
    error: Optional[str] = Field(default=None, description="Error message if failed")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Additional metadata")
    execution_time: float = Field(description="Execution time in seconds")
    timestamp: datetime = Field(default_factory=lambda: datetime.now())


class ToolCapability(BaseModel):
    """Describes a capability of a tool."""
    name: str = Field(description="Name of the capability")
    description: str = Field(description="Description of what this capability does")
    parameters: Dict[str, Any] = Field(default_factory=dict, description="Required parameters")
    examples: List[str] = Field(default_factory=list, description="Usage examples")


class BaseTool(ABC):
    """Base class for all tools in the InfraGPT agent system."""
    
    def __init__(self, name: str, description: str):
        """Initialize the tool.
        
        Args:
            name: Name of the tool
            description: Description of what the tool does
        """
        self.name = name
        self.description = description
        self.logger = structlog.get_logger(f"tool.{name}")
        self._is_initialized = False
    
    @abstractmethod
    async def initialize(self) -> bool:
        """Initialize the tool (e.g., check dependencies, authenticate).
        
        Returns:
            True if initialization was successful, False otherwise
        """
        pass
    
    @abstractmethod
    async def execute(self, parameters: Dict[str, Any]) -> ToolExecutionResult:
        """Execute the tool with given parameters.
        
        Args:
            parameters: Parameters for the tool execution
            
        Returns:
            ToolExecutionResult containing the execution result
        """
        pass
    
    @abstractmethod
    def get_capabilities(self) -> List[ToolCapability]:
        """Get list of capabilities this tool provides.
        
        Returns:
            List of ToolCapability objects describing what the tool can do
        """
        pass
    
    async def validate_parameters(self, parameters: Dict[str, Any]) -> bool:
        """Validate parameters before execution.
        
        Args:
            parameters: Parameters to validate
            
        Returns:
            True if parameters are valid, False otherwise
        """
        # Default implementation - can be overridden by specific tools
        return True
    
    async def health_check(self) -> bool:
        """Check if the tool is healthy and ready to use.
        
        Returns:
            True if tool is healthy, False otherwise
        """
        return self._is_initialized
    
    async def cleanup(self) -> None:
        """Clean up resources used by the tool."""
        self._is_initialized = False
    
    def _create_error_result(self, error_message: str, execution_time: float = 0.0) -> ToolExecutionResult:
        """Create an error result.
        
        Args:
            error_message: The error message
            execution_time: Time taken before the error occurred
            
        Returns:
            ToolExecutionResult with error information
        """
        return ToolExecutionResult(
            success=False,
            error=error_message,
            execution_time=execution_time,
            metadata={"tool": self.name}
        )
    
    def _create_success_result(
        self,
        output: str,
        execution_time: float,
        metadata: Optional[Dict[str, Any]] = None
    ) -> ToolExecutionResult:
        """Create a success result.
        
        Args:
            output: The output from the tool
            execution_time: Time taken for execution
            metadata: Additional metadata
            
        Returns:
            ToolExecutionResult with success information
        """
        result_metadata = {"tool": self.name}
        if metadata:
            result_metadata.update(metadata)
            
        return ToolExecutionResult(
            success=True,
            output=output,
            execution_time=execution_time,
            metadata=result_metadata
        )


class ConversationTool(BaseTool):
    """Tool for handling conversational responses."""
    
    def __init__(self):
        super().__init__(
            name="conversation",
            description="Tool for generating conversational responses"
        )
    
    async def initialize(self) -> bool:
        """Initialize the conversation tool."""
        self._is_initialized = True
        return True
    
    async def execute(self, parameters: Dict[str, Any]) -> ToolExecutionResult:
        """Execute a conversation response.
        
        Args:
            parameters: Should contain 'response' key with the response text
            
        Returns:
            ToolExecutionResult with the response
        """
        start_time = time.time()
        
        try:
            response = parameters.get("response", "I'm here to help!")
            execution_time = time.time() - start_time
            
            return self._create_success_result(
                output=response,
                execution_time=execution_time,
                metadata={"type": "conversation"}
            )
            
        except Exception as e:
            execution_time = time.time() - start_time
            self.logger.error("Error in conversation tool", error=str(e))
            return self._create_error_result(str(e), execution_time)
    
    def get_capabilities(self) -> List[ToolCapability]:
        """Get conversation tool capabilities."""
        return [
            ToolCapability(
                name="respond",
                description="Generate a conversational response",
                parameters={"response": "string"},
                examples=["Provide helpful information", "Answer questions"]
            )
        ]


class MCPTool(BaseTool):
    """Base class for tools that communicate with MCP servers."""
    
    def __init__(self, name: str, description: str, server_url: Optional[str] = None):
        """Initialize the MCP tool.
        
        Args:
            name: Name of the tool
            description: Description of the tool
            server_url: URL of the MCP server (if applicable)
        """
        super().__init__(name, description)
        self.server_url = server_url
        self._client = None
    
    async def _call_mcp_function(
        self,
        function_name: str,
        parameters: Dict[str, Any],
        timeout: int = 30
    ) -> ToolExecutionResult:
        """Call an MCP server function.
        
        Args:
            function_name: Name of the function to call
            parameters: Parameters for the function
            timeout: Timeout in seconds
            
        Returns:
            ToolExecutionResult with function output
        """
        start_time = time.time()
        
        try:
            self.logger.info("Calling MCP function", function=function_name, parameters=parameters)
            
            # This would be implemented to call the actual MCP server
            # For now, we'll simulate the behavior
            execution_time = time.time() - start_time
            
            # Placeholder implementation - would integrate with actual MCP client
            result_output = f"MCP function '{function_name}' called with parameters: {parameters}"
            
            return self._create_success_result(
                output=result_output,
                execution_time=execution_time,
                metadata={
                    "function": function_name,
                    "server_url": self.server_url,
                    "mcp_version": "1.0"
                }
            )
            
        except Exception as e:
            execution_time = time.time() - start_time
            self.logger.error("Error calling MCP function", function=function_name, error=str(e))
            return self._create_error_result(str(e), execution_time)
    
    async def validate_parameters(self, parameters: Dict[str, Any]) -> bool:
        """Validate MCP function parameters."""
        # Check that required parameters exist
        if "function" not in parameters:
            return False
        
        function_name = parameters["function"]
        if not isinstance(function_name, str) or not function_name.strip():
            return False
        
        return True


class FunctionTool(BaseTool):
    """Base class for tools that use direct function calls."""
    
    def __init__(self, name: str, description: str):
        """Initialize the function tool.
        
        Args:
            name: Name of the tool
            description: Description of the tool
        """
        super().__init__(name, description)
        self._functions: Dict[str, callable] = {}
    
    def register_function(self, name: str, func: callable, description: str = "") -> None:
        """Register a function with this tool.
        
        Args:
            name: Name of the function
            func: The callable function
            description: Description of what the function does
        """
        self._functions[name] = {
            "func": func,
            "description": description
        }
        self.logger.info("Function registered", function=name, description=description)
    
    async def _call_function(
        self,
        function_name: str,
        parameters: Dict[str, Any],
        timeout: int = 30
    ) -> ToolExecutionResult:
        """Call a registered function.
        
        Args:
            function_name: Name of the function to call
            parameters: Parameters for the function
            timeout: Timeout in seconds
            
        Returns:
            ToolExecutionResult with function output
        """
        start_time = time.time()
        
        try:
            if function_name not in self._functions:
                return self._create_error_result(
                    f"Function '{function_name}' not found in tool '{self.name}'"
                )
            
            func_info = self._functions[function_name]
            func = func_info["func"]
            
            self.logger.info("Calling function", function=function_name, parameters=parameters)
            
            # Call the function (handle both sync and async)
            if asyncio.iscoroutinefunction(func):
                result = await asyncio.wait_for(func(**parameters), timeout=timeout)
            else:
                result = func(**parameters)
            
            execution_time = time.time() - start_time
            
            # Convert result to string if it's not already
            output = str(result) if result is not None else "Function executed successfully"
            
            return self._create_success_result(
                output=output,
                execution_time=execution_time,
                metadata={
                    "function": function_name,
                    "description": func_info["description"]
                }
            )
            
        except asyncio.TimeoutError:
            execution_time = time.time() - start_time
            return self._create_error_result(
                f"Function call timed out after {timeout} seconds",
                execution_time
            )
        except Exception as e:
            execution_time = time.time() - start_time
            self.logger.error("Error calling function", function=function_name, error=str(e))
            return self._create_error_result(str(e), execution_time)
    
    def list_functions(self) -> List[str]:
        """Get list of registered function names."""
        return list(self._functions.keys())
    
    async def validate_parameters(self, parameters: Dict[str, Any]) -> bool:
        """Validate function parameters."""
        if "function" not in parameters:
            return False
        
        function_name = parameters["function"]
        if not isinstance(function_name, str) or not function_name.strip():
            return False
        
        if function_name not in self._functions:
            return False
        
        return True