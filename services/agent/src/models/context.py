"""Context models for agent processing."""

from typing import Dict, Any, Optional
from pydantic import BaseModel, Field


class AgentContext(BaseModel):
    """Context information for agent processing."""
    
    conversation_id: str = Field(..., description="Conversation identifier")
    user_id: Optional[str] = Field(default=None, description="User identifier")
    channel_id: Optional[str] = Field(default=None, description="Channel identifier")
    
    # Message context
    current_message: str = Field(..., description="Current message being processed")
    message_history: list[str] = Field(default_factory=list, description="Previous messages")
    
    # Additional context
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Additional metadata")
    
    # Processing context
    selected_agent: Optional[str] = Field(default=None, description="Selected agent type")
    available_tools: list[str] = Field(default_factory=list, description="Available tools for processing")