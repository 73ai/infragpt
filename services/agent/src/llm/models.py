"""Data models for LLM integration."""

from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional
from pydantic import BaseModel, Field


class Message(BaseModel):
    """Represents a message in a conversation."""

    role: str = Field(
        description="Role of the message sender (user, assistant, system)"
    )
    content: str = Field(description="Content of the message")
    timestamp: datetime = Field(default_factory=lambda: datetime.now())
    metadata: Dict[str, Any] = Field(default_factory=dict)


class ConversationContext(BaseModel):
    """Context for a conversation including history and metadata."""

    conversation_id: str = Field(description="Unique identifier for the conversation")
    messages: List[Message] = Field(default_factory=list, description="Message history")
    user_id: Optional[str] = Field(default=None, description="User identifier")
    channel_id: Optional[str] = Field(default=None, description="Channel identifier")
    metadata: Dict[str, Any] = Field(
        default_factory=dict, description="Additional context"
    )

    def add_message(
        self, role: str, content: str, metadata: Optional[Dict[str, Any]] = None
    ) -> None:
        """Add a message to the conversation context."""
        message = Message(role=role, content=content, metadata=metadata or {})
        self.messages.append(message)

    def get_recent_messages(self, limit: int = 10) -> List[Message]:
        """Get the most recent messages from the conversation."""
        return self.messages[-limit:] if self.messages else []


class LLMResponse(BaseModel):
    """Response from the LLM."""

    content: str = Field(description="Generated response content")
    metadata: Dict[str, Any] = Field(
        default_factory=dict, description="Model and performance metadata"
    )
