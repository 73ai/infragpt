"""
Main entry point for the InfraGPT Agent Service.

This service provides AI-powered infrastructure management capabilities
through both FastAPI (health checks) and gRPC (agent processing) interfaces.
"""

import asyncio
import signal
import sys
from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI

from src.config.settings import Settings
from src.config.logging import setup_logging, get_logger
from src.api.health import router as health_router
from src.grpc.server import GRPCServer

# Global variables for service lifecycle
grpc_server = None
logger = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Manage the service lifecycle."""
    global grpc_server, logger
    
    # Startup
    logger.info("Starting InfraGPT Agent Service...")
    
    # Start gRPC server
    settings = Settings()
    grpc_server = GRPCServer(settings)
    await grpc_server.start()
    
    logger.info("Service startup completed")
    
    yield
    
    # Shutdown
    logger.info("Shutting down InfraGPT Agent Service...")
    if grpc_server:
        await grpc_server.stop()
    logger.info("Service shutdown completed")


def create_app() -> FastAPI:
    """Create and configure the FastAPI application."""
    # Load settings and setup logging
    settings = Settings()
    setup_logging(settings.log_level)
    
    global logger
    logger = get_logger(__name__)
    
    # Create FastAPI app
    app = FastAPI(
        title="InfraGPT Agent Service",
        description="AI-powered infrastructure management agent",
        version="0.1.0",
        lifespan=lifespan
    )
    
    # Add routers
    app.include_router(health_router, prefix="/health", tags=["health"])
    
    return app


# Create the app instance
app = create_app()


def main():
    """Main service entry point."""
    try:
        settings = Settings()
        
        # Run FastAPI with uvicorn
        uvicorn.run(
            "main:app",
            host=settings.host,
            port=settings.http_port,
            log_level=settings.log_level.lower(),
            reload=settings.debug,
            access_log=True
        )
    except KeyboardInterrupt:
        logger.info("Service interrupted by user")
        sys.exit(0)
    except Exception as e:
        logger.error(f"Service failed to start: {e}", exc_info=True)
        sys.exit(1)


if __name__ == "__main__":
    main()
