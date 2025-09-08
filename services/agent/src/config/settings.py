"""Configuration settings for the Backend Agent Service."""

from typing import Optional
from pydantic import Field
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""

    # Service configuration
    host: str = Field(default="0.0.0.0", description="Host to bind the service to")
    http_port: int = Field(default=8000, description="HTTP port for FastAPI")
    grpc_port: int = Field(default=50051, description="gRPC port for agent service")
    debug: bool = Field(default=False, description="Enable debug mode")
    log_level: str = Field(default="INFO", description="Logging level")

    # External service configuration
    backend_service_host: str = Field(
        default="localhost", description="Backend main service host"
    )
    backend_service_port: int = Field(
        default=9090, description="Backend main service gRPC port"
    )

    # LLM configuration
    litellm_api_key: Optional[str] = Field(
        default=None, description="API key for LiteLLM (OpenAI, etc.)"
    )
    default_model: str = Field(default="gpt-4o", description="Default LLM model to use")

    # Tool configuration
    enable_kubectl: bool = Field(default=True, description="Enable kubectl tools")
    enable_gcloud: bool = Field(default=True, description="Enable gcloud tools")
    enable_github: bool = Field(default=False, description="Enable GitHub tools")

    # MCP configuration
    mcp_servers_config: str = Field(
        default="config/mcp_servers.yaml",
        description="Path to MCP servers configuration file",
    )

    model_config = {"env_file": ".env", "env_prefix": "AGENT_", "case_sensitive": False}
