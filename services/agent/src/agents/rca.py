"""RCA agent for root cause analysis and troubleshooting."""

import json
from src.agents.base import BaseAgent, AgentType
from src.models.agent import AgentResponse
from src.models.context import AgentContext
from src.models.intent import Intent, IntentType
from src.llm import LiteLLMClient, ConversationContext, Message


class RCAAgent(BaseAgent):
    """Agent that handles root cause analysis and troubleshooting."""

    def __init__(self, llm_client: LiteLLMClient = None):
        super().__init__(AgentType.RCA)
        # Configure LLM client for reasoning tasks - higher quality model
        self.llm_client = llm_client or LiteLLMClient(
            model="gpt-4o-mini",  # Use reasoning model for RCA
            temperature=0.3,  # Lower temperature for more focused analysis
            max_tokens=1500,  # Longer responses for detailed analysis
        )

    async def can_handle(self, context: AgentContext) -> bool:
        """
        Determine if this requires root cause analysis using LLM intent analysis.
        """
        intent = await self._analyze_intent(context.current_message)

        # Handle troubleshooting intents - include deployment failures
        rca_intents = [
            IntentType.TROUBLESHOOTING,
            IntentType.INFRASTRUCTURE_QUERY,
            IntentType.MONITORING,
            IntentType.DEPLOYMENT,  # Include deployment issues for RCA
        ]
        if intent.type in rca_intents:
            self.logger.debug(f"LLM detected RCA intent: {intent.type.value}")
            return True

        return False

    async def _analyze_intent(self, message: str) -> Intent:
        """Analyze user intent using LLM for RCA classification."""
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
                type=IntentType.UNKNOWN,
                confidence=0.5,
                entities={},
                description="Fallback due to analysis error",
            )

        # Parse JSON response
        try:
            intent_data = json.loads(llm_response.content)
            return Intent(
                type=IntentType(intent_data.get("type", "unknown")),
                confidence=intent_data.get("confidence", 0.5),
                entities={},
                description=intent_data.get("description", ""),
            )
        except (json.JSONDecodeError, ValueError, KeyError) as e:
            self.logger.warning("Failed to parse intent analysis", error=str(e))
            # Conservative fallback - don't assume it's RCA
            return Intent(
                type=IntentType.UNKNOWN,
                confidence=0.5,
                entities={},
                description="Fallback due to parsing error",
            )

    async def process(self, context: AgentContext) -> AgentResponse:
        """Process RCA request and generate LLM-powered analysis."""
        self.logger.info(f"Processing RCA for: {context.conversation_id}")

        # Create conversation context for LLM
        llm_context = self._build_llm_context(context)

        # Generate analysis using LLM
        system_prompt = (
            "You are an expert infrastructure troubleshooting and root cause analysis agent. "
            "Analyze the reported issue and provide structured, actionable analysis. "
            "Follow this format:\n\n"
            "## Issue Analysis\n"
            "[Brief summary of the reported issue]\n\n"
            "## Potential Root Causes\n"
            "[List 2-3 most likely causes]\n\n"
            "## Recommended Next Steps\n"
            "[Specific actions to diagnose or resolve]\n\n"
            "## Additional Information Needed\n"
            "[What else would help with diagnosis]\n\n"
            "Be technical but clear, and focus on actionable recommendations."
        )

        llm_response = await self.llm_client.generate_response(
            prompt=context.current_message,
            context=llm_context,
            system_prompt=system_prompt,
        )

        # Handle LLM failure with customized error message for RCA
        if llm_response.metadata.get("error"):
            response_text = "I'm unable to perform a detailed analysis right now. Please try describing the issue again, or consider checking basic troubleshooting steps like logs, resource usage, and recent changes."
        else:
            response_text = llm_response.content

        metadata = {
            "agent_type": self.agent_type.value,
            "analysis_type": "root_cause_analysis",
            "llm_metadata": llm_response.metadata,
        }

        return AgentResponse(
            success=True,
            response_text=response_text,
            agent_type=self.name,
            confidence=0.8,  # High confidence for RCA analysis when successful
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
