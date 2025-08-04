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
            response = self._client.chat.completions.create(
                model=self.model,
                messages=[{"role": "user", "content": "hi"}],
                max_tokens=1
            )
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
            
            # Buffer for tool calls
            tool_call_buffer = {}
            
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
                        tool_calls = self._process_tool_calls(choice.delta.tool_calls, tool_call_buffer)
                    
                    # Handle finish reason
                    finish_reason = choice.finish_reason
                    
                    yield StreamChunk(
                        content=content,
                        tool_calls=tool_calls,
                        finish_reason=finish_reason
                    )
                    
        except Exception as e:
            raise self._map_error(e)
    
    def _build_request(self, messages: List[Dict[str, Any]], tools: Optional[List[Dict]], **kwargs) -> Dict[str, Any]:
        """Build OpenAI API request."""
        request = {
            "model": self.model,
            "messages": messages,
            "stream": True,
            "temperature": kwargs.get("temperature", 0.0),
        }
        
        if kwargs.get("max_tokens"):
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
    
    def _process_tool_calls(self, delta_tool_calls, buffer) -> Optional[List[ToolCall]]:
        """Process streaming tool calls."""
        completed_tools = []
        
        for delta_call in delta_tool_calls:
            call_id = delta_call.id
            if call_id:
                if call_id not in buffer:
                    buffer[call_id] = {
                        "id": call_id,
                        "name": "",
                        "arguments": "",
                        "complete": False
                    }
                
                if delta_call.function:
                    if delta_call.function.name:
                        buffer[call_id]["name"] = delta_call.function.name
                    if delta_call.function.arguments:
                        buffer[call_id]["arguments"] += delta_call.function.arguments
                
                # Check if this tool call is complete (no more deltas expected)
                tool_data = buffer[call_id]
                if (not buffer[call_id]["complete"] and 
                    tool_data["name"] and 
                    tool_data["arguments"] and
                    not (delta_call.function and delta_call.function.arguments)):
                    
                    # Mark as complete and try to parse
                    buffer[call_id]["complete"] = True
                    try:
                        arguments = json.loads(tool_data["arguments"]) if tool_data["arguments"].strip() else {}
                        completed_tools.append(ToolCall(
                            id=tool_data["id"],
                            name=tool_data["name"],
                            arguments=arguments
                        ))
                    except json.JSONDecodeError as e:
                        print(f"Warning: Failed to parse JSON for {tool_data['name']}: {tool_data['arguments']} - Error: {e}")
        
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