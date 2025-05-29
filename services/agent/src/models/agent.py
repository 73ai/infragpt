"""Data models for agent requests and responses."""

from typing import List, Optional
from pydantic import BaseModel, Field


class AgentRequest(BaseModel):
    """Request model for agent processing."""
    
    conversation_id: str = Field(..., description="Unique identifier for the conversation thread")
    current_message: str = Field(..., description="The current message to process")
    past_messages: List[str] = Field(default_factory=list, description="Previous messages in the conversation")
    context: str = Field(default="", description="Additional context information")
    user_id: Optional[str] = Field(default=None, description="User ID for personalization")
    channel_id: Optional[str] = Field(default=None, description="Channel/workspace information")


class AgentResponse(BaseModel):
    """Response model from agent processing."""
    
    success: bool = Field(..., description="Whether the processing was successful")
    response_text: str = Field(..., description="The generated response text")
    error_message: str = Field(default="", description="Error message if processing failed")
    agent_type: Optional[str] = Field(default=None, description="Type of agent that processed the request")
    confidence: Optional[float] = Field(default=None, description="Confidence score (0.0 to 1.0)")
    tools_used: List[str] = Field(default_factory=list, description="List of tools used in processing")