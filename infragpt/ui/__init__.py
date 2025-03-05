"""User interface components for InfraGPT."""

from .console import console
from .prompts import prompt_credentials, prompt_for_parameters
from .display import handle_command_result

__all__ = [
    "console", 
    "prompt_credentials", 
    "prompt_for_parameters",
    "handle_command_result"
]