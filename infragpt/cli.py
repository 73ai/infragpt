#\!/usr/bin/env python3

import os
import sys
import json
import datetime
from typing import Optional, List, Dict, Any

import click
from prompt_toolkit import PromptSession
from prompt_toolkit.history import FileHistory
from prompt_toolkit.styles import Style
from rich.panel import Panel
from rich.text import Text
import pathlib

try:
    import pyperclip
except ImportError:
    pass

from infragpt.config import (
    CONFIG_FILE, load_config, init_config, console
)
from infragpt.llm import (
    MODEL_TYPE, generate_gcloud_command, prompt_credentials, validate_env_api_keys
)
from infragpt.prompts import handle_command_result
from infragpt.history import history_command

def interactive_mode(model_type: Optional[MODEL_TYPE] = None, api_key: Optional[str] = None, verbose: bool = False):
    """Run InfraGPT in interactive mode with natural language prompting."""
    # Ensure history directory exists
    history_dir = pathlib.Path.home() / ".infragpt"
    history_dir.mkdir(exist_ok=True)
    history_file = history_dir / "history"
    
    # Setup prompt toolkit session with history
    session = PromptSession(history=FileHistory(str(history_file)))
    
    # Style for prompt
    style = Style.from_dict({
        'prompt': '#00FFFF bold',
    })
    
    # Get actual model to display, either from params or config
    actual_model = model_type
    if not actual_model:
        config = load_config()
        actual_model = config.get("model")
    
    # Welcome message
    console.print(Panel.fit(
        Text("InfraGPT - Interactive natural language to gcloud commands", style="bold green"),
        border_style="blue"
    ))
    
    # If no model configured or empty API key, prompt for credentials now
    config = load_config()
    has_model = actual_model is not None
    has_api_key = api_key is not None and api_key.strip()
    
    if not has_model and not has_api_key:
        # Check config as well for empty API key
        config_api_key = config.get("api_key", "")
        if actual_model and (not config_api_key or not config_api_key.strip()):
            model_type, api_key = prompt_credentials(actual_model)
        else:
            model_type, api_key = prompt_credentials(actual_model)
        actual_model = model_type
    
    console.print(f"[yellow]Using model:[/yellow] [bold]{actual_model}[/bold]")
    console.print("[dim]Press Ctrl+D to exit, Ctrl+C to clear input[/dim]\n")
    
    while True:
        try:
            # Get user input with prompt toolkit
            user_input = session.prompt(
                [('class:prompt', '> ')], 
                style=style,
                multiline=False
            )
            
            if not user_input.strip():
                continue
                
            with console.status("[bold green]Generating command...[/bold green]", spinner="dots"):
                result = generate_gcloud_command(user_input, model_type, verbose)
            
            handle_command_result(result, model_type, verbose)
        except KeyboardInterrupt:
            # Clear the current line and show a new prompt
            console.print("\n[yellow]Input cleared. Enter a new prompt:[/yellow]")
            continue
        except EOFError:
            # Exit on Ctrl+D
            console.print("\n[bold]Exiting InfraGPT.[/bold]")
            sys.exit(0)

@click.group(invoke_without_command=True)
@click.pass_context
@click.version_option(package_name='infragpt')
@click.option('--model', '-m', type=click.Choice(['gpt4o', 'claude']), 
              help='LLM model to use (gpt4o or claude)')
@click.option('--api-key', '-k', help='API key for the selected model')
@click.option('--verbose', '-v', is_flag=True, help='Enable verbose output')
def cli(ctx, model, api_key, verbose):
    """InfraGPT - Convert natural language to Google Cloud commands in interactive mode."""
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
    """InfraGPT - Convert natural language to Google Cloud commands in interactive mode."""
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
    
    # Enter interactive mode
    interactive_mode(model, api_key, verbose)

if __name__ == "__main__":
    cli()
