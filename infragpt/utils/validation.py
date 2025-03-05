"""Validation functions for InfraGPT."""

import os
import sys
from typing import Tuple, Optional, Literal

from ..ui.console import console

MODEL_TYPE = Literal["gpt4o", "claude"]

def validate_env_api_keys() -> Tuple[Optional[MODEL_TYPE], Optional[str]]:
    """Validate API keys from environment variables and prompt if invalid."""
    openai_key = os.getenv("OPENAI_API_KEY")
    anthropic_key = os.getenv("ANTHROPIC_API_KEY")
    env_model = os.getenv("INFRAGPT_MODEL")
    
    # Import here to avoid circular imports
    from ..core.llm import validate_api_key
    from ..ui.prompts import prompt_credentials
    
    # If we have specific model set in env but invalid key, prompt for it
    if env_model == "gpt4o" and openai_key:
        if not validate_api_key("gpt4o", openai_key):
            console.print("[bold red]Invalid OpenAI API key in environment variable.[/bold red]")
            model, api_key = prompt_credentials("gpt4o")
            # Update environment for this session
            os.environ["OPENAI_API_KEY"] = api_key
            return "gpt4o", api_key
    elif env_model == "claude" and anthropic_key:
        if not validate_api_key("claude", anthropic_key):
            console.print("[bold red]Invalid Anthropic API key in environment variable.[/bold red]")
            model, api_key = prompt_credentials("claude")
            # Update environment for this session
            os.environ["ANTHROPIC_API_KEY"] = api_key
            return "claude", api_key
    
    # For default case or when no specific model set
    if openai_key and (not env_model or env_model == "gpt4o"):
        if not validate_api_key("gpt4o", openai_key):
            console.print("[bold red]Invalid OpenAI API key in environment variable.[/bold red]")
            model, api_key = prompt_credentials("gpt4o")
            # Update environment for this session
            os.environ["OPENAI_API_KEY"] = api_key
            return "gpt4o", api_key
    elif anthropic_key and (not env_model or env_model == "claude"):
        if not validate_api_key("claude", anthropic_key):
            console.print("[bold red]Invalid Anthropic API key in environment variable.[/bold red]")
            model, api_key = prompt_credentials("claude")
            # Update environment for this session
            os.environ["ANTHROPIC_API_KEY"] = api_key
            return "claude", api_key
            
    return None, None