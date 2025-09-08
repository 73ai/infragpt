"""Conversation agent for handling general user interactions."""

import json
from src.agents.base import BaseAgent, AgentType
from src.models.agent import AgentResponse
from src.models.context import AgentContext
from src.models.intent import Intent, IntentType
from src.llm import LiteLLMClient, ConversationContext


class ConversationAgent(BaseAgent):
    """Agent that handles general conversational interactions."""

    def __init__(self, llm_client: LiteLLMClient = None):
        super().__init__(AgentType.CONVERSATION)
        # Configure LLM client for conversational tasks
        self.llm_client = llm_client or LiteLLMClient(
            model="gpt-3.5-turbo", temperature=0.7, max_tokens=1000
        )

    async def can_handle(self, context: AgentContext) -> bool:
        """
        Determine if this is a conversational request using LLM intent analysis.
        """
        intent = await self._analyze_intent(context.current_message)

        # Handle conversational intents
        conversational_intents = [IntentType.CONVERSATION, IntentType.UNKNOWN]
        if intent.type in conversational_intents:
            self.logger.debug(
                f"LLM detected conversational intent: {intent.type.value}"
            )
            return True

        return False

    async def _analyze_intent(self, message: str) -> Intent:
        """Analyze user intent using LLM."""
        system_prompt = """You are an intent analysis system for an infrastructure management AI agent.
        Analyze the user's message and determine their intent. Respond with ONLY a JSON object containing:
        {
            "type": "one of: conversation, infrastructure_query, deployment, troubleshooting, monitoring, configuration, security, unknown",
            "confidence": "float between 0.0 and 1.0",
            "description": "brief description of the intent"
        }"""

        llm_response = await self.llm_client.generate_response(
            prompt=f"Analyze this message: {message}", system_prompt=system_prompt
        )

        # Handle LLM failure case
        if llm_response.metadata.get("error"):
            return Intent(
                type=IntentType.CONVERSATION,
                confidence=0.5,
                entities={},
                description="Fallback to conversation due to analysis error",
            )

        # Parse JSON response
        try:
            intent_data = json.loads(llm_response.content)
            return Intent(
                type=IntentType(intent_data.get("type", "conversation")),
                confidence=intent_data.get("confidence", 0.5),
                entities={},
                description=intent_data.get("description", ""),
            )
        except (json.JSONDecodeError, ValueError, KeyError) as e:
            self.logger.warning("Failed to parse intent analysis", error=str(e))
            # Fallback to conversation intent
            return Intent(
                type=IntentType.CONVERSATION,
                confidence=0.5,
                entities={},
                description="Fallback to conversation due to parsing error",
            )

    async def process(self, context: AgentContext) -> AgentResponse:
        """Process conversational request and generate LLM-powered response."""
        self.logger.info(f"Processing conversation for: {context.conversation_id}")

        # Create conversation context for LLM
        llm_context = self._build_llm_context(context)

        # Generate response using LLM
        system_prompt = (
            "You are an intelligent infrastructure management AI agent assistant. "
            "You help users with general questions, provide information about your capabilities, "
            "and engage in helpful conversations about infrastructure topics. "
            "Be friendly, helpful, and professional. Keep responses concise but informative."
        )

        llm_response = await self.llm_client.generate_response(
            prompt=context.current_message,
            context=llm_context,
            system_prompt=system_prompt,
        )

        # Handle LLM failure with customized error message
        if llm_response.metadata.get("error"):
            response_text = "I'm having trouble processing your message right now. Could you please try asking again in a different way?"
        else:
            response_text = llm_response.content

        metadata = {
            "agent_type": self.agent_type.value,
            "llm_metadata": llm_response.metadata,
        }

        return AgentResponse(
            success=True,
            response_text=response_text,
            agent_type=self.name,
            confidence=0.9,  # Static confidence since LLM confidence is not reliable
            tools_used=[],
            metadata=metadata,
        )

    def _build_llm_context(self, context: AgentContext) -> ConversationContext:
        """Build LLM conversation context from agent context."""
        llm_context = ConversationContext(
            conversation_id=context.conversation_id,
            user_id=context.user_id,
            channel_id=context.channel_id,
            metadata=context.metadata or {},
        )

        # Convert message history to LLM format
        for msg in context.message_history:
            llm_context.add_message(
                role="user" if msg.sender == "user" else "assistant",
                content=msg.content,
                metadata={"timestamp": msg.timestamp},
            )

        return llm_context
