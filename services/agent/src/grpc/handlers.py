"""gRPC service handlers for the Backend Agent Service."""

import logging
from typing import Optional

import grpc

from src.proto import agent_pb2, agent_pb2_grpc
from src.models.agent import AgentRequest, AgentResponse
from src.models.context import Message
from src.config.settings import Settings
from src.agents import AgentSystem
from src.integrations import ReplyHandler

logger = logging.getLogger(__name__)


class AgentServiceHandler(agent_pb2_grpc.AgentServiceServicer):
    """Handler for AgentService gRPC methods."""

    def __init__(self, settings: Settings):
        self.settings = settings
        self.agent_system: Optional[AgentSystem] = None
        self.reply_handler = ReplyHandler(
            backend_host=getattr(settings, "backend_service_host", "localhost"),
            backend_port=getattr(settings, "backend_service_port", 9090),
        )
        logger.info("AgentServiceHandler initialized")

    async def initialize_agent_system(self) -> None:
        """Initialize the agent system."""
        if self.agent_system is None:
            self.agent_system = AgentSystem()
            await self.agent_system.initialize()
            logger.info("Agent system initialized in gRPC handler")

    async def ProcessMessage(
        self, request: agent_pb2.AgentRequest, context: grpc.aio.ServicerContext
    ) -> agent_pb2.AgentResponse:
        """
        Process incoming agent requests from Backend service.

        Args:
            request: The gRPC request containing message and context
            context: The gRPC service context

        Returns:
            AgentResponse with the processed result
        """
        try:
            logger.info(
                f"Processing message for conversation: {request.conversation_id}"
            )

            # Ensure agent system is initialized
            await self.initialize_agent_system()

            if not self.agent_system or not self.agent_system.is_ready():
                raise RuntimeError("Agent system not ready")

            # Convert protobuf to internal model
            agent_request = self._convert_request(request)

            # Process with agent system
            agent_response = await self.agent_system.process_request(agent_request)

            # Send reply back to Slack if successful
            if agent_response.success and agent_response.response_text:
                await self._send_reply_to_slack(agent_request, agent_response)

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
                tools_used=[],
            )

    def _convert_request(self, pb_request: agent_pb2.AgentRequest) -> AgentRequest:
        """Convert protobuf request to internal model."""
        # Convert protobuf messages to internal Message objects
        past_messages = []
        for pb_msg in pb_request.past_messages:
            past_messages.append(
                Message(
                    message_id=pb_msg.message_id,
                    content=pb_msg.content,
                    sender=pb_msg.sender,
                    timestamp=pb_msg.timestamp,
                )
            )

        return AgentRequest(
            conversation_id=pb_request.conversation_id,
            current_message=pb_request.current_message,
            past_messages=past_messages,
            context=pb_request.context,
            user_id=pb_request.user_id if pb_request.user_id else None,
            channel_id=pb_request.channel_id if pb_request.channel_id else None,
        )

    def _convert_response(
        self, agent_response: AgentResponse
    ) -> agent_pb2.AgentResponse:
        """Convert internal response to protobuf."""
        return agent_pb2.AgentResponse(
            success=agent_response.success,
            response_text=agent_response.response_text,
            error_message=agent_response.error_message,
            agent_type=agent_response.agent_type or "",
            confidence=agent_response.confidence or 0.0,
            tools_used=agent_response.tools_used,
        )

    async def _send_reply_to_slack(
        self, request: AgentRequest, response: AgentResponse
    ) -> None:
        """Send agent response back to Slack through Backend service."""
        try:
            # Extract conversation context
            context_dict = None
            if request.context:
                import json

                try:
                    context_dict = json.loads(request.context)
                except json.JSONDecodeError:
                    logger.warning(f"Failed to parse context: {request.context}")

            # Send the reply
            success = await self.reply_handler.send_agent_response(
                conversation_id=request.conversation_id,
                response_text=response.response_text,
                context=context_dict,
            )

            if success:
                logger.info(
                    f"Successfully sent reply for conversation {request.conversation_id}"
                )
            else:
                logger.error(
                    f"Failed to send reply for conversation {request.conversation_id}"
                )

        except Exception as e:
            logger.error(f"Error sending reply to Slack: {e}", exc_info=True)
