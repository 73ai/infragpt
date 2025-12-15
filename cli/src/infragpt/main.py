#!/usr/bin/env python3

import os
import sys
from typing import Optional

import click
from rich.panel import Panel

from infragpt.config import init_config, console
from infragpt.llm.router import LLMRouter
from infragpt.llm.exceptions import ValidationError, AuthenticationError
from infragpt.history import history_command
from infragpt.agent import run_shell_agent
from infragpt.container import is_sandbox_mode, get_executor, cleanup_old_containers, DockerNotAvailableError
from infragpt.tools import cleanup_executor


@click.group(invoke_without_command=True)
@click.pass_context
@click.version_option(package_name="infragpt")
@click.option(
    "--model",
    "-m",
    help="Model in provider:model format (e.g., openai:gpt-4o, anthropic:claude-3-5-sonnet-20241022)",
)
@click.option("--api-key", "-k", help="API key for the selected provider")
@click.option("--verbose", "-v", is_flag=True, help="Enable verbose output")
def cli(ctx, model, api_key, verbose):
    """InfraGPT V2 - Interactive shell operations with direct SDK integration."""
    # If no subcommand is specified, go to interactive mode
    if ctx.invoked_subcommand is None:
        main(model=model, api_key=api_key, verbose=verbose)


@cli.command(name="history")
@click.option(
    "--limit", "-l", type=int, default=10, help="Number of history entries to display"
)
@click.option("--type", "-t", help="Filter by interaction type")
@click.option("--export", "-e", help="Export history to file path")
def history_cli(limit, type, export):
    """View or export interaction history."""
    history_command(limit, type, export)


@cli.command(name="providers")
def providers_cli():
    """Show supported providers and example model strings."""
    console.print(
        Panel.fit(
            "Supported Providers and Model Examples",
            border_style="blue",
            title="[bold green]Providers[/bold green]",
        )
    )

    providers = LLMRouter.get_supported_providers()
    examples = LLMRouter.get_provider_examples()

    for provider, config in providers.items():
        console.print(f"\n[bold cyan]{provider.upper()}[/bold cyan]")
        console.print(f"  Example: [yellow]{examples[provider]}[/yellow]")
        console.print(f"  Default params: {config['default_params']}")


def get_credentials_v2(
    model_string: Optional[str] = None,
    api_key: Optional[str] = None,
    verbose: bool = False,
):
    """Get credentials for the new system."""
    # If model is provided, validate it
    if model_string:
        if not LLMRouter.validate_model_string(model_string):
            raise ValidationError("Invalid model format. Use 'provider:model' format.")

        provider_name, model_name = LLMRouter.parse_model_string(model_string)
    else:
        provider_name = None

    # Try to get API key from various sources
    if not api_key:
        # Try environment variables
        if provider_name == "openai":
            api_key = os.getenv("OPENAI_API_KEY")
        elif provider_name == "anthropic":
            api_key = os.getenv("ANTHROPIC_API_KEY")
        else:
            # Check both if provider not specified
            openai_key = os.getenv("OPENAI_API_KEY")
            anthropic_key = os.getenv("ANTHROPIC_API_KEY")

            if openai_key and not model_string:
                model_string = "openai:gpt-5"
                api_key = openai_key
                provider_name = "openai"
            elif anthropic_key and not model_string:
                model_string = "anthropic:claude-sonnet-4-20250514"
                api_key = anthropic_key
                provider_name = "anthropic"

    # If still no credentials, prompt user
    if not model_string or not api_key:
        console.print(
            "\n[yellow]No valid credentials found. Please provide model and API key.[/yellow]"
        )

        if not model_string:
            console.print("\nSupported formats:")
            examples = LLMRouter.get_provider_examples()
            for provider, example in examples.items():
                console.print(f"  {provider}: [cyan]{example}[/cyan]")

            while True:
                model_input = console.input("\nEnter model (provider:model): ").strip()
                if LLMRouter.validate_model_string(model_input):
                    model_string = model_input
                    provider_name, _ = LLMRouter.parse_model_string(model_string)
                    break
                else:
                    console.print(
                        "[red]Invalid format. Please use 'provider:model' format.[/red]"
                    )

        if not api_key:
            api_key = console.input(f"Enter API key for {provider_name}: ").strip()

    return model_string, api_key


def main(model, api_key, verbose):
    """InfraGPT V2 - Interactive shell operations with direct SDK integration."""
    # Initialize config file if it doesn't exist
    init_config()

    if verbose:
        from importlib.metadata import version

        try:
            console.print(f"[dim]InfraGPT V2 version: {version('infragpt')}[/dim]")
        except Exception:
            console.print("[dim]InfraGPT V2: Version information not available[/dim]")

    # Sandbox mode check - at startup
    sandbox_enabled = is_sandbox_mode()
    if sandbox_enabled:
        try:
            removed = cleanup_old_containers()
            if removed > 0:
                console.print(f"[dim]Cleaned up {removed} old sandbox container(s)[/dim]")
            console.print(
                "[yellow]Sandbox mode enabled - starting Docker container...[/yellow]"
            )
            executor = get_executor()
            executor.start()  # This checks Docker availability
            console.print("[green]Sandbox container ready.[/green]\n")
        except DockerNotAvailableError as e:
            console.print(f"[red]Error: {e}[/red]")
            console.print(
                "Please fix the issue above or disable sandbox mode with INFRAGPT_ISOLATED=false"
            )
            sys.exit(1)

    try:
        model_string, resolved_api_key = get_credentials_v2(model, api_key, verbose)

        if verbose:
            console.print(f"[dim]Using model: {model_string}[/dim]")

        run_shell_agent(model_string, resolved_api_key, verbose)

    except ValidationError as e:
        console.print(f"[red]Validation Error: {e}[/red]")
        console.print(
            "\nUse --help to see usage information or run 'infragpt providers' to see supported providers."
        )
        sys.exit(1)
    except AuthenticationError as e:
        console.print(f"[red]Authentication Error: {e}[/red]")
        sys.exit(1)
    except Exception as e:
        console.print(f"[red]Error: {e}[/red]")
        if verbose:
            import traceback

            console.print(traceback.format_exc())
        sys.exit(1)
    finally:
        if sandbox_enabled:
            cleanup_executor()


if __name__ == "__main__":
    cli()
