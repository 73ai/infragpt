"""gRPC service handlers for the InfraGPT Agent Service."""

import logging
from typing import Dict, Any

import grpc

from src.proto import agent_pb2, agent_pb2_grpc
from src.models.agent import AgentRequest, AgentResponse
from src.config.settings import Settings

logger = logging.getLogger(__name__)


class AgentServiceHandler(agent_pb2_grpc.AgentServiceServicer):
    """Handler for AgentService gRPC methods."""
    
    def __init__(self, settings: Settings):
        self.settings = settings
        # Agent system will be initialized here later
        self.agent_system = None
        logger.info("AgentServiceHandler initialized")
    
    async def ProcessMessage(
        self, 
        request: agent_pb2.AgentRequest, 
        context: grpc.aio.ServicerContext
    ) -> agent_pb2.AgentResponse:
        """
        Process incoming agent requests from InfraGPT service.
        
        Args:
            request: The gRPC request containing message and context
            context: The gRPC service context
            
        Returns:
            AgentResponse with the processed result
        """
        try:
            logger.info(f"Processing message for conversation: {request.conversation_id}")
            
            # Convert protobuf to internal model
            agent_request = self._convert_request(request)
            
            # Process with agent system (placeholder for now)
            response_text = await self._process_message_placeholder(agent_request)
            
            # Convert response back to protobuf
            return agent_pb2.AgentResponse(
                success=True,
                response_text=response_text,
                error_message="",
                agent_type="placeholder",
                confidence=0.9,
                tools_used=[]
            )
            
        except Exception as e:
            logger.error(f"Error processing agent request: {e}", exc_info=True)
            return agent_pb2.AgentResponse(
                success=False,
                response_text="",
                error_message=str(e),
                agent_type="error",
                confidence=0.0,
                tools_used=[]
            )
    
    def _convert_request(self, pb_request: agent_pb2.AgentRequest) -> AgentRequest:
        """Convert protobuf request to internal model."""
        return AgentRequest(
            conversation_id=pb_request.conversation_id,
            current_message=pb_request.current_message,
            past_messages=list(pb_request.past_messages),
            context=pb_request.context,
            user_id=pb_request.user_id if pb_request.user_id else None,
            channel_id=pb_request.channel_id if pb_request.channel_id else None
        )
    
    async def _process_message_placeholder(self, request: AgentRequest) -> str:
        """
        Placeholder message processing - will be replaced with agent system.
        
        Args:
            request: The converted agent request
            
        Returns:
            A placeholder response string
        """
        message_count = len(request.past_messages)
        
        return (
            f"Hello! I'm the InfraGPT Agent. I received your message: '{request.current_message}'. "
            f"I see we have {message_count} previous messages in our conversation. "
            f"This is a placeholder response that will be replaced with intelligent agent processing."
        )