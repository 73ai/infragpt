"""Conversation agent for handling general user interactions."""

from src.agents.base import BaseAgent, AgentType
from src.models.agent import AgentResponse
from src.models.context import AgentContext


class ConversationAgent(BaseAgent):
    """Agent that handles general conversational interactions."""
    
    def __init__(self):
        super().__init__(AgentType.CONVERSATION)
    
    async def can_handle(self, context: AgentContext) -> bool:
        """
        Determine if this is a conversational request.
        
        Returns True for general questions, greetings, status requests,
        and any message that doesn't indicate a technical issue.
        """
        message = context.current_message.lower()
        
        # Conversational indicators
        conversational_patterns = [
            "hello", "hi", "hey", "thanks", "thank you",
            "how are you", "what can you do", "help",
            "status", "how is", "tell me about"
        ]
        
        # Check for conversational patterns
        for pattern in conversational_patterns:
            if pattern in message:
                self.logger.debug(f"Detected conversational pattern: {pattern}")
                return True
        
        # If message is a question without error indicators
        if message.endswith("?") and not self._has_error_indicators(message):
            self.logger.debug("Detected general question")
            return True
        
        # Default to conversation for simple, short messages
        if len(message.split()) <= 5 and not self._has_error_indicators(message):
            self.logger.debug("Detected simple message")
            return True
        
        return False
    
    async def process(self, context: AgentContext) -> AgentResponse:
        """Process conversational request and generate response."""
        self.logger.info(f"Processing conversation for: {context.conversation_id}")
        
        try:
            # For now, use a simple rule-based response
            # This will be replaced with LLM integration
            response_text = await self._generate_conversational_response(context)
            
            return AgentResponse(
                success=True,
                response_text=response_text,
                agent_type=self.name,
                confidence=0.8,
                tools_used=[]
            )
        
        except Exception as e:
            self.logger.error(f"Error in conversation processing: {e}")
            return AgentResponse(
                success=False,
                response_text="I'm having trouble processing your request right now.",
                error_message=str(e),
                agent_type=self.name,
                confidence=0.0,
                tools_used=[]
            )
    
    async def _generate_conversational_response(self, context: AgentContext) -> str:
        """Generate a conversational response."""
        message = context.current_message.lower()
        
        # Simple rule-based responses (will be replaced with LLM)
        if any(greeting in message for greeting in ["hello", "hi", "hey"]):
            return "Hello! I'm the InfraGPT Agent. I can help you with infrastructure questions and troubleshoot issues. What can I assist you with today?"
        
        elif any(thanks in message for thanks in ["thanks", "thank you"]):
            return "You're welcome! Is there anything else I can help you with?"
        
        elif "what can you do" in message or "help" in message:
            return ("I can help you with:\n"
                   "• General infrastructure questions\n"
                   "• Troubleshooting and root cause analysis\n"
                   "• System status and monitoring\n"
                   "\nJust describe what you need help with!")
        
        elif "status" in message:
            return "I'm running and ready to help! What system or service status would you like me to check?"
        
        elif message.endswith("?"):
            return f"That's a great question about '{context.current_message}'. Let me help you with that. Could you provide a bit more context about what specifically you'd like to know?"
        
        else:
            return f"I understand you mentioned: '{context.current_message}'. How can I help you with this? Feel free to ask me any infrastructure-related questions."
    
    def _has_error_indicators(self, message: str) -> bool:
        """Check if message contains error indicators that suggest RCA is needed."""
        error_indicators = [
            "error", "failed", "failure", "broken", "not working",
            "down", "crashed", "exception", "timeout", "500", "404"
        ]
        
        return any(indicator in message for indicator in error_indicators)