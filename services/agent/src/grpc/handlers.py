"""gRPC service handlers for the InfraGPT Agent Service."""

import logging
from typing import Optional

import grpc

from src.proto import agent_pb2, agent_pb2_grpc
from src.models.agent import AgentRequest, AgentResponse
from src.config.settings import Settings
from src.agents import AgentSystem

logger = logging.getLogger(__name__)


class AgentServiceHandler(agent_pb2_grpc.AgentServiceServicer):
    """Handler for AgentService gRPC methods."""
    
    def __init__(self, settings: Settings):
        self.settings = settings
        self.agent_system: Optional[AgentSystem] = None
        logger.info("AgentServiceHandler initialized")
    
    async def initialize_agent_system(self) -> None:
        """Initialize the agent system."""
        if self.agent_system is None:
            self.agent_system = AgentSystem()
            await self.agent_system.initialize()
            logger.info("Agent system initialized in gRPC handler")
    
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
            
            # Ensure agent system is initialized
            await self.initialize_agent_system()
            
            if not self.agent_system or not self.agent_system.is_ready():
                raise RuntimeError("Agent system not ready")
            
            # Convert protobuf to internal model
            agent_request = self._convert_request(request)
            
            # Process with agent system
            agent_response = await self.agent_system.process_request(agent_request)
            
            # Convert response back to protobuf
            return self._convert_response(agent_response)
            
        except Exception as e:
            logger.error(f"Error processing agent request: {e}", exc_info=True)
            return agent_pb2.AgentResponse(
                success=False,
                response_text="I encountered an error while processing your request.",
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
    
    def _convert_response(self, agent_response: AgentResponse) -> agent_pb2.AgentResponse:
        """Convert internal response to protobuf."""
        return agent_pb2.AgentResponse(
            success=agent_response.success,
            response_text=agent_response.response_text,
            error_message=agent_response.error_message,
            agent_type=agent_response.agent_type or "",
            confidence=agent_response.confidence or 0.0,
            tools_used=agent_response.tools_used
        )