"""LiteLLM client for AI agent interactions."""

import os
from typing import AsyncGenerator, Dict, List, Optional
import litellm
from litellm import acompletion
import structlog
from dotenv import load_dotenv

from .models import ConversationContext, LLMResponse, Message

# Load environment variables
load_dotenv()

logger = structlog.get_logger(__name__)


class LiteLLMClient:
    """Clean LiteLLM client focused only on LLM API calls."""

    def __init__(
        self,
        model: str = "gpt-3.5-turbo",
        api_key: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: int = 1000,
    ):
        """Initialize the LiteLLM client.

        Args:
            model: The model to use (e.g., "gpt-3.5-turbo", "o1-preview")
            api_key: API key for the model provider
            temperature: Sampling temperature (0.0 to 1.0)
            max_tokens: Maximum tokens to generate
        """
        self.model = model
        self.temperature = temperature
        self.max_tokens = max_tokens

        # Try to get API key from environment if not provided
        self.api_key = (
            api_key or os.getenv("AGENT_LITELLM_API_KEY") or os.getenv("OPENAI_API_KEY")
        )

        # Configure litellm
        if self.api_key:
            litellm.api_key = self.api_key

        logger.info("LiteLLM client initialized", model=model)

    async def generate_response(
        self,
        prompt: str,
        context: Optional[ConversationContext] = None,
        system_prompt: Optional[str] = None,
    ) -> LLMResponse:
        """Generate a response using the LLM.

        Args:
            prompt: The user prompt
            context: Conversation context for multi-turn conversations
            system_prompt: System prompt to guide the model

        Returns:
            LLMResponse containing the generated content and metadata
        """
        try:
            messages = self._build_messages(prompt, context, system_prompt)

            response = await acompletion(
                model=self.model,
                messages=messages,
                temperature=self.temperature,
                max_tokens=self.max_tokens,
            )

            content = response.choices[0].message.content

            # Extract basic metadata
            metadata = {
                "model": self.model,
                "temperature": self.temperature,
                "max_tokens": self.max_tokens,
            }

            # Add token usage if available (future token counting)
            if response.usage:
                try:
                    metadata["usage"] = {
                        "prompt_tokens": getattr(response.usage, "prompt_tokens", 0),
                        "completion_tokens": getattr(
                            response.usage, "completion_tokens", 0
                        ),
                        "total_tokens": getattr(response.usage, "total_tokens", 0),
                    }
                except Exception as usage_error:
                    logger.warning(
                        "Failed to extract usage data", error=str(usage_error)
                    )

            return LLMResponse(content=content, metadata=metadata)

        except Exception as e:
            logger.error("LLM API call failed", error=str(e), model=self.model)
            # Return generic error response - agents will customize the message
            return LLMResponse(
                content="I encountered an error processing your request. Please try again.",
                metadata={"error": True, "model": self.model},
            )

    async def stream_response(
        self,
        prompt: str,
        context: Optional[ConversationContext] = None,
        system_prompt: Optional[str] = None,
    ) -> AsyncGenerator[str, None]:
        """Stream a response from the LLM.

        Args:
            prompt: The user prompt
            context: Conversation context for multi-turn conversations
            system_prompt: System prompt to guide the model

        Yields:
            Chunks of the generated response
        """
        try:
            messages = self._build_messages(prompt, context, system_prompt)

            response = await acompletion(
                model=self.model,
                messages=messages,
                temperature=self.temperature,
                max_tokens=self.max_tokens,
                stream=True,
            )

            async for chunk in response:
                if chunk.choices and chunk.choices[0].delta.content:
                    yield chunk.choices[0].delta.content

        except Exception as e:
            logger.error("LLM streaming failed", error=str(e), model=self.model)
            yield "I encountered an error processing your request. Please try again."

    def _build_messages(
        self,
        prompt: str,
        context: Optional[ConversationContext] = None,
        system_prompt: Optional[str] = None,
    ) -> List[Dict[str, str]]:
        """Build message list for the LLM API call."""
        messages = []

        # Add system prompt
        if system_prompt:
            messages.append({"role": "system", "content": system_prompt})
        else:
            messages.append(
                {
                    "role": "system",
                    "content": "You are an intelligent infrastructure management AI agent. "
                    "Help users manage, deploy, and troubleshoot their cloud infrastructure. "
                    "Be helpful, accurate, and provide actionable advice.",
                }
            )

        # Add conversation history
        if context and context.messages:
            for msg in context.get_recent_messages(limit=10):
                messages.append(
                    {
                        "role": msg.role,
                        "content": msg.content,
                    }
                )

        # Add current prompt
        messages.append({"role": "user", "content": prompt})

        return messages
