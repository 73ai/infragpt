"""Tests for health check endpoints."""

import pytest
from fastapi.testclient import TestClient

# Import the app without the lifespan context for testing
from src.config.settings import Settings
from src.config.logging import setup_logging
from src.api.health import router as health_router
from fastapi import FastAPI


def create_test_app() -> FastAPI:
    """Create a test FastAPI app without the gRPC server."""
    settings = Settings()
    setup_logging(settings.log_level)

    app = FastAPI(
        title="Agent Service (Test)",
        description="AI-powered infrastructure management agent",
        version="0.1.0",
    )

    app.include_router(health_router, prefix="/health", tags=["health"])
    return app


@pytest.fixture
def client():
    """Create a test client."""
    app = create_test_app()
    return TestClient(app)


def test_health_check(client):
    """Test the basic health check endpoint."""
    response = client.get("/health/")

    assert response.status_code == 200
    data = response.json()

    assert data["status"] == "healthy"
    assert "timestamp" in data
    assert "version" in data
    assert "components" in data
    assert data["components"]["fastapi"] == "running"


def test_liveness_check(client):
    """Test the liveness check endpoint."""
    response = client.get("/health/live")

    assert response.status_code == 200
    data = response.json()

    assert data["status"] == "alive"
    assert "timestamp" in data


def test_readiness_check(client):
    """Test the readiness check endpoint."""
    response = client.get("/health/ready")

    # Should be ready since all basic checks should pass
    assert response.status_code == 200
    data = response.json()

    assert data["ready"] is True
    assert "checks" in data
    assert "timestamp" in data
