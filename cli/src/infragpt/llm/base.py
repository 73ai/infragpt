"""
Abstract base class for LLM providers.
"""

from abc import ABC, abstractmethod
from typing import List, Dict, Any, Iterator, Optional
from .models import StreamChunk, Message


class BaseLLMProvider(ABC):
    """Abstract base class for LLM providers with unified tool calling interface."""
    
    def __init__(self, api_key: str, model: str, **kwargs):
        self.api_key = api_key
        self.model = model
        self.provider_name = self.__class__.__name__.replace("Provider", "").lower()
        self._client = self._initialize_client(**kwargs)
    
    @abstractmethod
    def _initialize_client(self, **kwargs):
        """Initialize provider-specific client."""
        pass
    
    @abstractmethod
    def stream(self, messages: List[Dict[str, Any]], tools: Optional[List[Dict]] = None, **kwargs) -> Iterator[StreamChunk]:
        """Stream response with unified tool calling support."""
        pass
    
    @abstractmethod
    def validate_api_key(self) -> bool:
        """Validate API key with a simple test call."""
        pass
    
    @abstractmethod
    def _convert_messages(self, messages: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Convert unified message format to provider-specific format."""
        pass
    
    @abstractmethod
    def _convert_tools(self, tools: List['Tool']) -> List[Dict]:
        """Convert Tool objects to provider-specific format."""
        pass
    
    @abstractmethod
    def _normalize_chunk(self, raw_chunk) -> StreamChunk:
        """Convert provider-specific chunk to unified format."""
        pass
    
    @abstractmethod
    def _map_error(self, error: Exception) -> Exception:
        """Map provider-specific error to unified exception."""
        pass