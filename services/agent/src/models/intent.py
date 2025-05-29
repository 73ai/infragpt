"""Intent analysis models for agent decision making."""

from enum import Enum
from typing import Any, Dict
from pydantic import BaseModel, Field


class IntentType(str, Enum):
    """Types of user intents."""
    CONVERSATION = "conversation"
    INFRASTRUCTURE_QUERY = "infrastructure_query"
    DEPLOYMENT = "deployment"
    TROUBLESHOOTING = "troubleshooting"
    MONITORING = "monitoring"
    CONFIGURATION = "configuration"
    SECURITY = "security"
    UNKNOWN = "unknown"


class Intent(BaseModel):
    """Represents a user's intent parsed from their message."""
    type: IntentType = Field(description="Type of intent")
    confidence: float = Field(ge=0.0, le=1.0, description="Confidence score for the intent")
    entities: Dict[str, Any] = Field(default_factory=dict, description="Extracted entities")
    description: str = Field(description="Human-readable description of the intent")