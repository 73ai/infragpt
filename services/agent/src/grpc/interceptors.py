"""gRPC interceptors for logging and monitoring."""

import logging
import time
from typing import Callable, Any

import grpc


class LoggingInterceptor(grpc.aio.ServerInterceptor):
    """Interceptor for logging gRPC requests and responses."""

    def __init__(self):
        self.logger = logging.getLogger(__name__)

    async def intercept_service(
        self,
        continuation: Callable,
        handler_call_details: grpc.HandlerCallDetails,
    ) -> Any:
        """Intercept and log gRPC service calls."""
        start_time = time.time()
        method = handler_call_details.method

        self.logger.info(f"gRPC call started: {method}")

        try:
            response = await continuation(handler_call_details)
            duration = time.time() - start_time
            self.logger.info(f"gRPC call completed: {method} ({duration:.3f}s)")
            return response
        except Exception as e:
            duration = time.time() - start_time
            self.logger.error(f"gRPC call failed: {method} ({duration:.3f}s) - {e}")
            raise
