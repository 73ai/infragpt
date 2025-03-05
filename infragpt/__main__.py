#!/usr/bin/env python3
"""Command-line entry point for InfraGPT."""

# Import directly from cli.commands to ensure we're using the correct cli implementation
from .cli.commands import cli

if __name__ == "__main__":
    cli()