"""
OpenAI provider implementation using direct SDK.
"""

import json
from typing import List, Dict, Any, Iterator, Optional
from openai import OpenAI
import openai

from ..base import BaseLLMProvider
from ..models import StreamChunk, ToolCall
from ..exceptions import (
    AuthenticationError,
    RateLimitError,
    APIError,
    ToolCallError,
    ContextWindowError,
    ValidationError,
)


class OpenAIProvider(BaseLLMProvider):
    """OpenAI provider using official Python SDK."""
    
    def _initialize_client(self, **kwargs):
        """Initialize OpenAI client."""
        return OpenAI(api_key=self.api_key)
    
    def validate_api_key(self) -> bool:
        """Validate API key with a simple test call."""
        try:
            # Use appropriate parameter based on model
            params = {
                "model": self.model,
                "messages": [{"role": "user", "content": "hi"}]
            }
            
            # Newer models like o4-mini use max_completion_tokens
            if self.model.startswith("o4") or self.model.startswith("o1"):
                params["max_completion_tokens"] = 10
            else:
                params["max_tokens"] = 10
                
            response = self._client.chat.completions.create(**params)
            return True
        except Exception as e:
            raise self._map_error(e)
    
    def stream(self, messages: List[Dict[str, Any]], tools: Optional[List[Dict]] = None, **kwargs) -> Iterator[StreamChunk]:
        """Stream response with unified tool calling support."""
        try:
            # Convert to OpenAI format
            request_params = self._build_request(messages, tools, **kwargs)
            
            # Stream response
            response = self._client.chat.completions.create(**request_params)
            
            # Buffer for tool calls - persistent across chunks
            tool_call_buffer = {}
            accumulated_tool_calls = []
            last_call_id = None  # Track the last call ID for continuations
            
            for chunk in response:
                if chunk.choices and len(chunk.choices) > 0:
                    choice = chunk.choices[0]
                    
                    # Handle content
                    content = None
                    if choice.delta.content:
                        content = choice.delta.content
                    
                    # Handle tool calls
                    tool_calls = None
                    if choice.delta.tool_calls:
                        # Update last_call_id if we see a new ID
                        for tc in choice.delta.tool_calls:
                            if tc.id:
                                last_call_id = tc.id
                                
                        tool_calls = self._process_tool_calls(choice.delta.tool_calls, tool_call_buffer, last_call_id)
                        if tool_calls:
                            accumulated_tool_calls.extend(tool_calls)
                    
                    # Handle finish reason
                    finish_reason = choice.finish_reason
                    
                    # Yield content chunks without tool calls
                    if content or (finish_reason and finish_reason != "tool_calls"):
                        yield StreamChunk(
                            content=content,
                            tool_calls=None,  # Don't yield tool calls during streaming
                            finish_reason=finish_reason if finish_reason != "tool_calls" else None
                        )
                    
                    # Only emit tool calls once when we get the finish reason
                    if finish_reason == "tool_calls" and accumulated_tool_calls:
                        yield StreamChunk(
                            content=None,
                            tool_calls=accumulated_tool_calls,
                            finish_reason=finish_reason
                        )
                        accumulated_tool_calls = []  # Reset after emitting
                    
        except Exception as e:
            raise self._map_error(e)
    
    def _build_request(self, messages: List[Dict[str, Any]], tools: Optional[List[Dict]], **kwargs) -> Dict[str, Any]:
        """Build OpenAI API request."""
        request = {
            "model": self.model,
            "messages": messages,
            "stream": True,
        }
        
        # Handle temperature - o4 models only support default temperature of 1.0
        if self.model.startswith("o4") or self.model.startswith("o1"):
            # Don't set temperature for o4/o1 models (uses default 1.0)
            pass
        else:
            request["temperature"] = kwargs.get("temperature", 0.0)
        
        if kwargs.get("max_tokens"):
            # Use appropriate parameter based on model
            if self.model.startswith("o4") or self.model.startswith("o1"):
                request["max_completion_tokens"] = kwargs["max_tokens"]
            else:
                request["max_tokens"] = kwargs["max_tokens"]
        
        if tools:
            request["tools"] = self._convert_tools(tools)
            request["tool_choice"] = "auto"
        
        return request
    
    def _convert_messages(self, messages: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """Convert unified message format to OpenAI format."""
        # OpenAI format is already our unified format
        return messages
    
    def _convert_tools(self, tools: List['Tool']) -> List[Dict]:
        """Convert Tool objects to OpenAI format."""
        from ..models import Tool
        
        openai_tools = []
        for tool in tools:
            # Convert InputSchema to dict format (same as Anthropic but wrapped differently)
            parameters_dict = {
                "type": tool.input_schema.type,
                "properties": {
                    name: {
                        "type": param.type,
                        "description": param.description,
                        **({"enum": param.enum} if param.enum else {}),
                        **({"default": param.default} if param.default is not None else {})
                    }
                    for name, param in tool.input_schema.properties.items()
                },
                "required": tool.input_schema.required,
                "additionalProperties": tool.input_schema.additionalProperties
            }
            
            openai_tools.append({
                "type": "function",
                "function": {
                    "name": tool.name,
                    "description": tool.description,
                    "parameters": parameters_dict
                }
            })
        
        return openai_tools
    
    def _process_tool_calls(self, delta_tool_calls, buffer, last_known_id=None) -> Optional[List[ToolCall]]:
        """Process streaming tool calls."""
        completed_tools = []
        current_call_id = None
        
        for i, delta_call in enumerate(delta_tool_calls):
            # Get call ID - if None, use the last known call ID (continuation of previous)
            call_id = delta_call.id
            if call_id:
                current_call_id = call_id
                if call_id not in buffer:
                    buffer[call_id] = {
                        "id": call_id,
                        "name": "",
                        "arguments": "",
                        "complete": False
                    }
            else:
                # No ID in this delta - use last known ID from parameter or current
                call_id = last_known_id if last_known_id else current_call_id
                if not call_id:
                    continue
                
            if delta_call.function:
                if delta_call.function.name:
                    buffer[call_id]["name"] = delta_call.function.name
                if delta_call.function.arguments:
                    buffer[call_id]["arguments"] += delta_call.function.arguments
            
            # Check if this tool call is complete
            # For o4 models, we need to check if the arguments form valid JSON
            tool_data = buffer[call_id]
            if not buffer[call_id]["complete"] and tool_data["name"] and tool_data["arguments"]:
                # Try to parse the arguments to see if they're complete JSON
                try:
                    arguments = json.loads(tool_data["arguments"])
                    # If parsing succeeds, the tool call is complete
                    buffer[call_id]["complete"] = True
                    completed_tools.append(ToolCall(
                        id=tool_data["id"],
                        name=tool_data["name"],
                        arguments=arguments
                    ))
                except json.JSONDecodeError:
                    # Not complete yet, keep accumulating
                    pass
        
        return completed_tools if completed_tools else None
    
    def _normalize_chunk(self, raw_chunk) -> StreamChunk:
        """Convert OpenAI chunk to unified format."""
        # This method is not used in the current implementation
        # as we handle normalization in the stream method
        pass
    
    def _map_error(self, error: Exception) -> Exception:
        """Map OpenAI errors to unified exceptions."""
        if isinstance(error, openai.AuthenticationError):
            return AuthenticationError(str(error), provider="openai", model=self.model)
        elif isinstance(error, openai.RateLimitError):
            return RateLimitError(str(error), provider="openai", model=self.model)
        elif isinstance(error, openai.BadRequestError):
            if "context window" in str(error).lower():
                return ContextWindowError(str(error), provider="openai", model=self.model)
            return ValidationError(str(error), provider="openai", model=self.model)
        elif isinstance(error, openai.APIStatusError):
            return APIError(str(error), status_code=error.status_code, provider="openai", model=self.model)
        else:
            return APIError(f"OpenAI API error: {error}", provider="openai", model=self.model)