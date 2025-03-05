"""Shared CLI options for InfraGPT."""

import click
from typing import Dict, Any

def common_options() -> Dict[str, Any]:
    """Return a dictionary of common CLI options."""
    return {
        "model": click.option(
            "--model", "-m", 
            type=click.Choice(["gpt4o", "claude"]), 
            help="LLM model to use (gpt4o or claude)"
        ),
        "api_key": click.option(
            "--api-key", "-k", 
            help="API key for the selected model"
        ),
        "verbose": click.option(
            "--verbose", "-v", 
            is_flag=True, 
            help="Enable verbose output"
        )
    }

def add_common_options(command):
    """Add common options to a command."""
    options = common_options()
    for option in options.values():
        command = option(command)
    return command