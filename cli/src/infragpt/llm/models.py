"""
Data models for unified LLM interface.
"""

from dataclasses import dataclass, field
from typing import Dict, Any, Optional, List, Union
from datetime import datetime, timezone
import uuid


@dataclass
class Parameter:
    """Schema for a single parameter."""

    type: str  # "string", "integer", "number", "boolean", "array", "object"
    description: Optional[str] = None
    enum: Optional[List[Any]] = None
    default: Optional[Any] = None


@dataclass
class InputSchema:
    """Schema for tool input parameters."""

    type: str = "object"  # Always "object" for tools
    properties: Dict[str, Parameter] = field(default_factory=dict)
    required: List[str] = field(default_factory=list)
    additionalProperties: bool = False


@dataclass
class Tool:
    """Tool definition for LLM function calling."""

    name: str
    description: str
    input_schema: InputSchema


@dataclass
class ToolCall:
    """Standardized tool call representation."""

    id: str
    name: str
    arguments: Dict[str, Any]


@dataclass
class StreamChunk:
    """Standardized streaming chunk."""

    content: Optional[str] = None
    tool_calls: Optional[List[ToolCall]] = None
    finish_reason: Optional[str] = None


@dataclass
class Message:
    """Provider-agnostic message format."""

    role: str  # 'user', 'assistant', 'system', 'tool'
    content: Union[str, List[Dict[str, Any]]]
    timestamp: datetime = field(default_factory=lambda: datetime.now(timezone.utc))

    # Tool calling support
    tool_calls: Optional[List[ToolCall]] = None
    tool_call_id: Optional[str] = None

    # Provider-native storage for efficiency
    _provider_native: Dict[str, Any] = field(default_factory=dict)

    # Metadata
    metadata: Dict[str, Any] = field(default_factory=dict)

    # Message ID for referencing
    message_id: str = field(default_factory=lambda: str(uuid.uuid4()))
