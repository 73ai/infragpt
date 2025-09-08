"""gRPC server implementation for the Backend Agent Service."""

import asyncio
import logging
from concurrent.futures import ThreadPoolExecutor
from typing import Optional

import grpc
from grpc_reflection.v1alpha import reflection

from src.proto import agent_pb2_grpc
from src.grpc.handlers import AgentServiceHandler
from src.grpc.interceptors import LoggingInterceptor
from src.config.settings import Settings

logger = logging.getLogger(__name__)


class GRPCServer:
    """gRPC server for the agent service."""

    def __init__(self, settings: Settings):
        self.settings = settings
        self.server: Optional[grpc.aio.Server] = None
        self.handler = AgentServiceHandler(settings)

    async def start(self) -> None:
        """Start the gRPC server."""
        logger.info("Starting gRPC server...")

        # Create server with thread pool
        self.server = grpc.aio.server(
            ThreadPoolExecutor(max_workers=10), interceptors=[LoggingInterceptor()]
        )

        # Add service handler
        agent_pb2_grpc.add_AgentServiceServicer_to_server(self.handler, self.server)

        # Enable reflection for debugging
        if self.settings.debug:
            service_names = (
                "agent.AgentService",
                reflection.SERVICE_NAME,
            )
            reflection.enable_server_reflection(service_names, self.server)
            logger.info("gRPC reflection enabled")

        # Bind to port
        listen_addr = f"[::]:{self.settings.grpc_port}"
        self.server.add_insecure_port(listen_addr)

        # Start server
        await self.server.start()
        logger.info(f"gRPC server listening on {listen_addr}")

    async def stop(self, grace_period: int = 5) -> None:
        """Stop the gRPC server."""
        if self.server:
            logger.info("Stopping gRPC server...")
            await self.server.stop(grace=grace_period)
            logger.info("gRPC server stopped")

    async def wait_for_termination(self) -> None:
        """Wait for the server to terminate."""
        if self.server:
            await self.server.wait_for_termination()
