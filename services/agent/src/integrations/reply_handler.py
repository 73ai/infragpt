"""Reply handling integration for sending responses back to Slack."""

import json
import logging
from typing import Dict, Any, Optional

from .backend_client import BackendClient

logger = logging.getLogger(__name__)


class ReplyHandler:
    """
    Handles sending agent responses back to Slack through Backend service.
    
    Manages the integration between agent responses and Slack messaging,
    including conversation context and error handling.
    """
    
    def __init__(self, backend_host: str = "localhost", backend_port: int = 9090):
        """
        Initialize reply handler.
        
        Args:
            backend_host: Backend service host
            backend_port: Backend service gRPC port
        """
        self.backend_client = BackendClient(host=backend_host, port=backend_port)
        self.logger = logger
    
    async def send_agent_response(
        self,
        conversation_id: str,
        response_text: str,
        context: Optional[Dict[str, Any]] = None
    ) -> bool:
        """
        Send an agent response back to the appropriate Slack conversation.
        
        Args:
            conversation_id: The conversation ID to reply to
            response_text: The agent's response text
            context: Optional context information from the original request
            
        Returns:
            bool: True if the response was sent successfully
        """
        if not response_text or not response_text.strip():
            self.logger.warning(
                "Attempted to send empty response",
                extra={"conversation_id": conversation_id}
            )
            return False
        
        try:
            # Extract additional context if available
            if context:
                self.logger.debug(
                    "Sending agent response with context",
                    extra={
                        "conversation_id": conversation_id,
                        "response_length": len(response_text),
                        "context_keys": list(context.keys()) if isinstance(context, dict) else None
                    }
                )
            
            # Send the reply through Backend service
            success = await self.backend_client.send_reply(
                conversation_id=conversation_id,
                message=response_text
            )
            
            if success:
                self.logger.info(
                    "Agent response sent successfully",
                    extra={
                        "conversation_id": conversation_id,
                        "response_length": len(response_text)
                    }
                )
            else:
                self.logger.error(
                    "Failed to send agent response",
                    extra={"conversation_id": conversation_id}
                )
            
            return success
            
        except Exception as e:
            self.logger.error(
                "Error sending agent response",
                extra={
                    "conversation_id": conversation_id,
                    "error": str(e),
                    "response_length": len(response_text)
                }
            )
            return False
    
    def extract_conversation_id_from_context(self, context: str) -> Optional[str]:
        """
        Extract conversation ID from context string.
        
        The Go service sends context as JSON string containing conversation info.
        
        Args:
            context: JSON string containing conversation context
            
        Returns:
            Optional conversation ID if found
        """
        try:
            if not context:
                return None
            
            context_data = json.loads(context)
            
            # Try to extract conversation ID from various possible locations
            if isinstance(context_data, dict):
                # Direct conversation ID
                if "conversation_id" in context_data:
                    return context_data["conversation_id"]
                
                # Build from conversation components
                conversation = context_data.get("conversation", {})
                if isinstance(conversation, dict):
                    team_id = conversation.get("team_id")
                    channel_id = conversation.get("channel_id") 
                    thread_ts = conversation.get("thread_ts")
                    
                    if all([team_id, channel_id, thread_ts]):
                        return f"{team_id}_{channel_id}_{thread_ts}"
            
            return None
            
        except (json.JSONDecodeError, AttributeError) as e:
            self.logger.warning(
                "Failed to parse conversation context",
                extra={"context": context, "error": str(e)}
            )
            return None
    
    def close(self):
        """Close the Backend client connection."""
        self.backend_client.close()