#\!/usr/bin/env python3

import os
import sys
from typing import Optional

import click
from rich.panel import Panel

from infragpt.config import (
    CONFIG_FILE, load_config, init_config, console
)
from infragpt.llm.models import MODEL_TYPE
from infragpt.llm_adapter import (
    validate_env_api_keys, prompt_credentials
)
from infragpt.history import history_command
from infragpt.agent import run_shell_agent


@click.group(invoke_without_command=True)
@click.pass_context
@click.version_option(package_name='infragpt')
@click.option('--model', '-m', type=click.Choice(['gpt4o', 'claude']), 
              help='LLM model to use (gpt4o or claude)')
@click.option('--api-key', '-k', help='API key for the selected model')
@click.option('--verbose', '-v', is_flag=True, help='Enable verbose output')
def cli(ctx, model, api_key, verbose):
    """InfraGPT - Interactive shell operations and debugging agent."""
    # If no subcommand is specified, go to interactive mode
    if ctx.invoked_subcommand is None:
        main(model=model, api_key=api_key, verbose=verbose)

@cli.command(name='history')
@click.option('--limit', '-l', type=int, default=10, help='Number of history entries to display')
@click.option('--type', '-t', help='Filter by interaction type (e.g., command_generation, command_action, command_execution)')
@click.option('--export', '-e', help='Export history to file path')
def history_cli(limit, type, export):
    """View or export interaction history."""
    history_command(limit, type, export)

def main(model, api_key, verbose):
    """InfraGPT - Interactive shell operations and debugging agent."""
    # Initialize config file if it doesn't exist
    init_config()
    
    if verbose:
        from importlib.metadata import version
        try:
            console.print(f"[dim]InfraGPT version: {version('infragpt')}[/dim]")
        except:
            console.print("[dim]InfraGPT: Version information not available[/dim]")
    
    # Check if we need to prompt for credentials before starting
    config = load_config()
    
    # Case 1: Command-line provided model but empty API key
    if model and (not api_key or not api_key.strip()):
        model, api_key = prompt_credentials(model)
    # Case 2: No command-line credentials
    elif not model and not api_key:
        has_model = config.get("model") is not None
        has_api_key = config.get("api_key") is not None and config.get("api_key").strip()
        
        # Case 2a: Config has model but empty API key
        if has_model and not has_api_key:
            model, api_key = prompt_credentials(config.get("model"))
        # Case 2b: No valid credentials in config or empty API key
        elif not (has_model and has_api_key):
            # Check if we have environment variables
            openai_key = os.getenv("OPENAI_API_KEY")
            anthropic_key = os.getenv("ANTHROPIC_API_KEY")
            
            if not (openai_key or anthropic_key):
                # No credentials anywhere, prompt before continuing
                model, api_key = prompt_credentials()
    
    # Enter shell agent mode
    # Get credentials if not provided
    from infragpt.llm_adapter import get_credentials
    resolved_model, resolved_api_key = get_credentials(model, api_key, verbose)
    
    run_shell_agent(resolved_model, resolved_api_key, verbose)

if __name__ == "__main__":
    cli()
