"""InfraGPT API client integration for the agent service."""

import logging
from typing import Optional

from infragptapi.client import InfraGPTClient as BaseInfraGPTClient
from infragptapi.exceptions import InfraGPTError, ConnectionError, RequestError

logger = logging.getLogger(__name__)


class InfraGPTClient:
    """
    Wrapper around the InfraGPT API client for agent service use.
    
    Provides a clean interface for sending replies back to Slack
    through the InfraGPT Go service.
    """
    
    def __init__(self, host: str = "localhost", port: int = 9090):
        """
        Initialize InfraGPT client wrapper.
        
        Args:
            host: InfraGPT service host
            port: InfraGPT service gRPC port
        """
        self.client = BaseInfraGPTClient(host=host, port=port)
        self.logger = logger
    
    async def send_reply(self, conversation_id: str, message: str) -> bool:
        """
        Send a reply to a Slack conversation asynchronously.
        
        Args:
            conversation_id: The conversation UUID to reply to
            message: The message text to send
            
        Returns:
            bool: True if successful, False otherwise
        """
        try:
            # Note: The base client's send_reply is synchronous
            # We wrap it to maintain async compatibility with agent service
            success = self.client.send_reply(conversation_id, message)
            
            if success:
                self.logger.info(
                    "Successfully sent reply to conversation",
                    extra={
                        "conversation_id": conversation_id,
                        "message_length": len(message)
                    }
                )
            else:
                self.logger.warning(
                    "Failed to send reply to conversation",
                    extra={"conversation_id": conversation_id}
                )
            
            return success
            
        except ConnectionError as e:
            self.logger.error(
                "Connection error sending reply",
                extra={
                    "conversation_id": conversation_id,
                    "error": str(e)
                }
            )
            return False
            
        except RequestError as e:
            self.logger.error(
                "Request error sending reply",
                extra={
                    "conversation_id": conversation_id,
                    "error": str(e)
                }
            )
            return False
            
        except InfraGPTError as e:
            self.logger.error(
                "InfraGPT service error sending reply",
                extra={
                    "conversation_id": conversation_id,
                    "error": str(e)
                }
            )
            return False
            
        except Exception as e:
            self.logger.error(
                "Unexpected error sending reply",
                extra={
                    "conversation_id": conversation_id,
                    "error": str(e)
                }
            )
            return False
    
    def close(self):
        """Close the connection to InfraGPT service."""
        if hasattr(self.client, 'close'):
            self.client.close()