"""Health check endpoints for the Backend Agent Service."""

from datetime import datetime, UTC
from typing import Dict

from fastapi import APIRouter, HTTPException, status
from pydantic import BaseModel

router = APIRouter()


class HealthResponse(BaseModel):
    """Health check response model."""

    status: str
    timestamp: datetime
    version: str = "0.1.0"
    components: Dict[str, str]


class ReadinessResponse(BaseModel):
    """Readiness check response model."""

    ready: bool
    checks: Dict[str, bool]
    timestamp: datetime


@router.get("/", response_model=HealthResponse)
async def health_check():
    """
    Basic health check endpoint.

    Returns the overall health status of the service.
    """
    return HealthResponse(
        status="healthy",
        timestamp=datetime.now(UTC),
        components={
            "grpc_server": "running",
            "fastapi": "running",
            "agent_system": "placeholder",
        },
    )


@router.get("/ready", response_model=ReadinessResponse)
async def readiness_check():
    """
    Readiness check for Kubernetes and deployment systems.

    Validates that all required components are ready to serve traffic.
    """
    checks = {
        "grpc_server": await _check_grpc_server(),
        "config_loaded": _check_config(),
        "dependencies": await _check_dependencies(),
    }

    ready = all(checks.values())

    if not ready:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE, detail="Service not ready"
        )

    return ReadinessResponse(ready=ready, checks=checks, timestamp=datetime.now(UTC))


@router.get("/live")
async def liveness_check():
    """
    Liveness check for Kubernetes.

    Simple endpoint that returns 200 if the service is alive.
    """
    return {"status": "alive", "timestamp": datetime.now(UTC)}


async def _check_grpc_server() -> bool:
    """Check if gRPC server is responding."""
    # TODO: Implement actual gRPC server health check
    # For now, assume it's healthy if we can import the modules
    try:
        from src.grpc.server import GRPCServer  # noqa: F401

        return True
    except ImportError:
        return False


def _check_config() -> bool:
    """Check if configuration is valid."""
    try:
        from src.config.settings import Settings

        settings = Settings()
        # Basic validation that required fields are present
        return bool(settings.host and settings.grpc_port and settings.http_port)
    except Exception:
        return False


async def _check_dependencies() -> bool:
    """Check if critical dependencies are available."""
    try:
        # Check if we can import critical dependencies
        import grpc  # noqa: F401
        import fastapi  # noqa: F401
        import pydantic  # noqa: F401

        return True
    except ImportError:
        return False
