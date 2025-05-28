"""RCA agent for root cause analysis and troubleshooting."""

from src.agents.base import BaseAgent, AgentType
from src.models.agent import AgentResponse
from src.models.context import AgentContext


class RCAAgent(BaseAgent):
    """Agent that handles root cause analysis and troubleshooting."""
    
    def __init__(self):
        super().__init__(AgentType.RCA)
    
    async def can_handle(self, context: AgentContext) -> bool:
        """
        Determine if this requires root cause analysis.
        
        Returns True for error reports, failure descriptions,
        performance issues, and troubleshooting requests.
        """
        message = context.current_message.lower()
        
        # Error and failure indicators
        error_indicators = [
            "error", "failed", "failure", "broken", "not working",
            "down", "crashed", "exception", "timeout", "slow",
            "issue", "problem", "bug", "troubleshoot", "debug"
        ]
        
        # HTTP status codes
        status_codes = ["500", "502", "503", "504", "404", "403", "401"]
        
        # Check for error indicators
        for indicator in error_indicators:
            if indicator in message:
                self.logger.debug(f"Detected error indicator: {indicator}")
                return True
        
        # Check for status codes
        for code in status_codes:
            if code in message:
                self.logger.debug(f"Detected status code: {code}")
                return True
        
        # Check for stack traces or log patterns
        if any(pattern in message for pattern in ["stack trace", "traceback", "at line", "exception:"]):
            self.logger.debug("Detected stack trace pattern")
            return True
        
        # Check for operational issues
        operational_patterns = [
            "can't connect", "connection refused", "unable to", "not responding",
            "performance", "latency", "memory", "cpu", "disk"
        ]
        
        for pattern in operational_patterns:
            if pattern in message:
                self.logger.debug(f"Detected operational issue: {pattern}")
                return True
        
        return False
    
    async def process(self, context: AgentContext) -> AgentResponse:
        """Process RCA request and generate analysis."""
        self.logger.info(f"Processing RCA for: {context.conversation_id}")
        
        try:
            # Analyze the issue and generate response
            response_text = await self._perform_root_cause_analysis(context)
            
            return AgentResponse(
                success=True,
                response_text=response_text,
                agent_type=self.name,
                confidence=0.85,
                tools_used=[]  # Will be populated when we add tool integration
            )
        
        except Exception as e:
            self.logger.error(f"Error in RCA processing: {e}")
            return AgentResponse(
                success=False,
                response_text="I encountered an issue while analyzing your problem. Let me try a different approach.",
                error_message=str(e),
                agent_type=self.name,
                confidence=0.0,
                tools_used=[]
            )
    
    async def _perform_root_cause_analysis(self, context: AgentContext) -> str:
        """Perform root cause analysis on the reported issue."""
        message = context.current_message.lower()
        
        # Categorize the issue type
        issue_category = self._categorize_issue(message)
        
        # Generate structured analysis based on category
        if issue_category == "connectivity":
            return self._analyze_connectivity_issue(context)
        elif issue_category == "performance":
            return self._analyze_performance_issue(context)
        elif issue_category == "application_error":
            return self._analyze_application_error(context)
        elif issue_category == "http_error":
            return self._analyze_http_error(context)
        else:
            return self._analyze_generic_issue(context)
    
    def _categorize_issue(self, message: str) -> str:
        """Categorize the type of issue reported."""
        if any(pattern in message for pattern in ["can't connect", "connection refused", "timeout", "not responding"]):
            return "connectivity"
        elif any(pattern in message for pattern in ["slow", "performance", "latency", "memory", "cpu"]):
            return "performance"
        elif any(pattern in message for pattern in ["500", "502", "503", "504", "404", "403"]):
            return "http_error"
        elif any(pattern in message for pattern in ["exception", "error", "failed", "traceback"]):
            return "application_error"
        else:
            return "generic"
    
    def _analyze_connectivity_issue(self, context: AgentContext) -> str:
        """Analyze connectivity-related issues."""
        return (
            f"🔍 **Connectivity Issue Analysis**\n\n"
            f"I've identified a potential connectivity issue: `{context.current_message}`\n\n"
            f"**Common causes to investigate:**\n"
            f"• Network connectivity between services\n"
            f"• Firewall rules or security groups\n"
            f"• Service discovery configuration\n"
            f"• DNS resolution issues\n"
            f"• Load balancer health checks\n\n"
            f"**Recommended next steps:**\n"
            f"1. Check if the target service is running and healthy\n"
            f"2. Verify network connectivity (ping, telnet, curl)\n"
            f"3. Review security group and firewall rules\n"
            f"4. Check service logs for any startup issues\n\n"
            f"Would you like me to help you check any specific component?"
        )
    
    def _analyze_performance_issue(self, context: AgentContext) -> str:
        """Analyze performance-related issues."""
        return (
            f"⚡ **Performance Issue Analysis**\n\n"
            f"I've detected a performance concern: `{context.current_message}`\n\n"
            f"**Areas to investigate:**\n"
            f"• Resource utilization (CPU, memory, disk)\n"
            f"• Database query performance\n"
            f"• Network latency and bandwidth\n"
            f"• Application bottlenecks\n"
            f"• Cache efficiency\n\n"
            f"**Recommended monitoring:**\n"
            f"1. Check system resource metrics\n"
            f"2. Review application performance metrics\n"
            f"3. Analyze slow query logs if database-related\n"
            f"4. Monitor network latency patterns\n\n"
            f"What specific performance metrics would you like me to help you examine?"
        )
    
    def _analyze_application_error(self, context: AgentContext) -> str:
        """Analyze application errors and exceptions."""
        return (
            f"🐛 **Application Error Analysis**\n\n"
            f"I've identified an application error: `{context.current_message}`\n\n"
            f"**Analysis approach:**\n"
            f"• Review complete error message and stack trace\n"
            f"• Check application logs around the time of failure\n"
            f"• Verify configuration and environment variables\n"
            f"• Check for recent deployments or changes\n"
            f"• Validate input data and dependencies\n\n"
            f"**Next steps:**\n"
            f"1. Collect full stack trace if not already available\n"
            f"2. Check application logs for context\n"
            f"3. Verify all required services are accessible\n"
            f"4. Review recent changes that might be related\n\n"
            f"Can you share the full error message or stack trace?"
        )
    
    def _analyze_http_error(self, context: AgentContext) -> str:
        """Analyze HTTP status code errors."""
        message = context.current_message
        
        if "500" in message:
            error_type = "Internal Server Error (500)"
            description = "The server encountered an unexpected condition."
        elif "502" in message:
            error_type = "Bad Gateway (502)"
            description = "The server received an invalid response from an upstream server."
        elif "503" in message:
            error_type = "Service Unavailable (503)"
            description = "The server is temporarily unable to handle the request."
        elif "404" in message:
            error_type = "Not Found (404)"
            description = "The requested resource could not be found."
        else:
            error_type = "HTTP Error"
            description = "An HTTP error occurred."
        
        return (
            f"🌐 **HTTP Error Analysis**\n\n"
            f"**Error Type:** {error_type}\n"
            f"**Description:** {description}\n\n"
            f"**Investigation steps:**\n"
            f"1. Check server logs for detailed error information\n"
            f"2. Verify the service is running and healthy\n"
            f"3. Review load balancer and proxy configurations\n"
            f"4. Check for resource constraints or dependencies\n\n"
            f"**Context:** `{context.current_message}`\n\n"
            f"Would you like me to help you check specific logs or service status?"
        )
    
    def _analyze_generic_issue(self, context: AgentContext) -> str:
        """Analyze generic issues that don't fit specific categories."""
        return (
            f"🔧 **Issue Analysis**\n\n"
            f"I'm analyzing the reported issue: `{context.current_message}`\n\n"
            f"**General troubleshooting approach:**\n"
            f"• Gather more specific error details\n"
            f"• Check system and application logs\n"
            f"• Verify service status and dependencies\n"
            f"• Review recent changes or deployments\n"
            f"• Check resource availability and limits\n\n"
            f"**Next steps:**\n"
            f"1. Can you provide more specific error messages?\n"
            f"2. What was happening when the issue occurred?\n"
            f"3. Have there been any recent changes?\n"
            f"4. Which services or components are affected?\n\n"
            f"The more details you can provide, the better I can help diagnose the root cause."
        )