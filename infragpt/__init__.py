"""InfraGPT - Natural language to Google Cloud commands converter."""

__version__ = "0.1.5"
# Import directly from cli.commands to ensure we're using the correct cli implementation
from .cli.commands import cli, main

__all__ = ["cli", "main"]