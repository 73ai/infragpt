"""
Simple Python client for InfraGPT service

This client provides a clean interface for the Python agent service
to communicate with the Go InfraGPT service, hiding all gRPC complexity.
"""

import logging
from typing import Optional

import grpc

from .generated import infragpt_pb2, infragpt_pb2_grpc
from .exceptions import InfraGPTError, ConnectionError, RequestError


class InfraGPTClient:
    """
    Simple client for sending replies through InfraGPT service.
    
    Handles all gRPC connection details and provides a clean interface
    for the agent service.
    """
    
    def __init__(self, host: str = "localhost", port: int = 9090):
        """
        Initialize InfraGPT client.
        
        Args:
            host: InfraGPT service host
            port: InfraGPT service gRPC port
        """
        self.host = host
        self.port = port
        self._channel = None
        self._client = None
        self.logger = logging.getLogger(__name__)
    
    def send_reply(self, conversation_id: str, message: str) -> bool:
        """
        Send a reply to a Slack conversation.
        
        Args:
            conversation_id: The conversation UUID to reply to
            message: The message text to send
            
        Returns:
            bool: True if successful
            
        Raises:
            InfraGPTError: If the request fails
            ConnectionError: If unable to connect to service
            RequestError: If the service returns an error
        """
        try:
            self._ensure_connected()
            
            # Create request
            request = infragpt_pb2.SendReplyCommand(
                conversation_id=conversation_id,
                message=message
            )
            
            # Send request
            response = self._client.SendReply(request)
            
            if not response.success:
                raise RequestError(f"Service error: {response.error}")
            
            self.logger.info(f"Successfully sent reply to conversation {conversation_id}")
            return True
            
        except grpc.RpcError as e:
            error_msg = f"gRPC error: {e.details()}"
            self.logger.error(error_msg)
            raise ConnectionError(error_msg)
        except (RequestError, ConnectionError):
            # Re-raise our custom exceptions
            raise
        except Exception as e:
            error_msg = f"Unexpected error: {e}"
            self.logger.error(error_msg)
            raise InfraGPTError(error_msg)
    
    def _ensure_connected(self):
        """Ensure gRPC connection is established."""
        if self._client is None:
            try:
                self._channel = grpc.insecure_channel(f'{self.host}:{self.port}')
                self._client = infragpt_pb2_grpc.InfraGPTServiceStub(self._channel)
                
                self.logger.info(f"Connected to InfraGPT service at {self.host}:{self.port}")
                
            except Exception as e:
                raise ConnectionError(f"Failed to connect to InfraGPT service: {e}")
    
    def close(self):
        """Close the connection to InfraGPT service."""
        if self._channel:
            self._channel.close()
            self._client = None
            self._channel = None
            self.logger.info("Disconnected from InfraGPT service")
    
    def __enter__(self):
        """Context manager entry."""
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.close()