"""Command implementations for InfraGPT CLI."""

import sys
import os
import json
import time
import logging
import datetime
import pathlib
import click
from typing import Tuple, Optional, List

from ..ui.console import console
from ..ui.display import handle_command_result
from ..core.commands import generate_gcloud_command
from ..core.llm import MODEL_TYPE, LLMResponseCache
from ..utils.config import init_config, Config
from ..utils.history import get_interaction_history
from .options import add_common_options

@click.group(invoke_without_command=True)
@click.pass_context
@click.version_option(package_name='infragpt')
def cli(ctx):
    """InfraGPT - Convert natural language to Google Cloud commands and manage history."""
    # If no subcommand is specified, go to interactive mode
    if ctx.invoked_subcommand is None:
        ctx.invoke(main, prompt=())

    # Set up global exception handling
    import sys
    from ..utils.logging import handle_error, logger, InfraGPTError

    def global_exception_handler(exc_type, exc_value, exc_traceback):
        """Handle uncaught exceptions."""
        if issubclass(exc_type, KeyboardInterrupt):
            # Don't handle keyboard interrupt with our custom handler
            sys.__excepthook__(exc_type, exc_value, exc_traceback)
            return

        if issubclass(exc_type, InfraGPTError):
            # Handle our custom exceptions
            handle_error(
                exc_value,
                log_level=logging.ERROR,
                exit_app=True,
                show_traceback=True,
                console_msg=f"Unexpected error: {str(exc_value)}"
            )
        else:
            # Handle other exceptions
            handle_error(
                exc_value,
                log_level=logging.CRITICAL,
                exit_app=True,
                show_traceback=True,
                console_msg="An unexpected error occurred. This has been logged."
            )

    # Set the global exception handler
    sys.excepthook = global_exception_handler

@cli.command(name='history')
@click.option('--limit', '-l', type=int, default=10, help='Number of history entries to display')
@click.option('--type', '-t', help='Filter by interaction type (e.g., command_generation, command_action, command_execution)')
@click.option('--export', '-e', help='Export history to file path')
@click.option('--format', '-f', type=click.Choice(['jsonl', 'json', 'csv']), default='jsonl',
              help='Export format (only used with --export)')
@click.option('--start-date', help='Filter entries after this date (YYYY-MM-DD or ISO format)')
@click.option('--end-date', help='Filter entries before this date (YYYY-MM-DD or ISO format)')
@click.option('--search', '-s', help='Filter entries containing this term')
@click.option('--stats', is_flag=True, help='Show history statistics instead of entries')
@click.option('--clear', is_flag=True, help='Clear history entries')
@click.option('--clear-before', help='Clear entries older than this date (YYYY-MM-DD or ISO format)')
def history_command(limit, type, export, format, start_date, end_date, search, stats, clear, clear_before):
    """View or export interaction history."""
    from ..utils.history import history_manager
    import json

    # Handle clear requests first
    if clear or clear_before:
        if clear_before:
            success, message = history_manager.clear_history(older_than=clear_before)
        else:
            if not click.confirm("Are you sure you want to clear all history?", default=False):
                console.print("[yellow]Operation cancelled.[/yellow]")
                return
            success, message = history_manager.clear_history()

        if success:
            console.print(f"[green]{message}[/green]")
        else:
            console.print(f"[bold red]Error:[/bold red] {message}")
        return

    # Handle stats request
    if stats:
        statistics = history_manager.get_statistics()

        console.print("[bold]History Statistics:[/bold]\n")

        console.print(f"[bold cyan]Total entries:[/bold cyan] {statistics['total_entries']}")

        # Show date range
        first_date = statistics.get('first_entry_date')
        last_date = statistics.get('last_entry_date')
        if first_date and last_date:
            if isinstance(first_date, datetime.datetime):
                first_date = first_date.strftime("%Y-%m-%d")
            if isinstance(last_date, datetime.datetime):
                last_date = last_date.strftime("%Y-%m-%d")
            console.print(f"[bold cyan]Date range:[/bold cyan] {first_date} to {last_date}")

        # Show entries by type
        console.print("\n[bold cyan]Entries by type:[/bold cyan]")
        for entry_type, count in statistics.get('entries_by_type', {}).items():
            console.print(f"  {entry_type}: {count}")

        # Show top commands
        top_commands = statistics.get('top_commands', [])
        if top_commands:
            console.print("\n[bold cyan]Top commands:[/bold cyan]")
            for cmd in top_commands:
                console.print(f"  {cmd['command']}: {cmd['count']} uses")

        # Show top parameters
        top_params = statistics.get('top_parameters', {})
        if top_params:
            console.print("\n[bold cyan]Top parameters:[/bold cyan]")
            for param, values in top_params.items():
                console.print(f"  {param}:")
                for val in values:
                    console.print(f"    {val['value']}: {val['count']} uses")

        return

    # Read history with filters
    entries = history_manager.get_entries(
        limit=limit,
        interaction_type=type,
        start_date=start_date,
        end_date=end_date,
        search_term=search
    )

    if not entries:
        console.print("[yellow]No history entries found.[/yellow]")
        return

    # Export if requested
    if export:
        success, message = history_manager.export_entries(
            export,
            format=format,
            limit=limit,
            interaction_type=type,
            start_date=start_date,
            end_date=end_date,
            search_term=search
        )

        if success:
            console.print(f"[green]{message}[/green]")
        else:
            console.print(f"[bold red]Error:[/bold red] {message}")
        return

    # Display history
    console.print(f"[bold]Last {len(entries)} interaction(s):[/bold]")

    for i, entry in enumerate(entries):
        entry_type = entry.get('type', 'unknown')
        timestamp = entry.get('timestamp', '')
        timestamp_short = timestamp.split('T')[0] if 'T' in timestamp else timestamp
        time_part = timestamp.split('T')[1].split('.')[0] if 'T' in timestamp else ""
        display_time = f"{timestamp_short} {time_part}"

        if entry_type == 'command_generation':
            data = entry.get('data', {})
            model = data.get('model', 'unknown')
            provider = data.get('provider', 'unknown')
            cached = "cached" if data.get('cached', False) else "not cached"
            prompt = data.get('prompt', '')
            result = data.get('result', '')
            duration_ms = data.get('duration_ms', 0)

            console.print(f"\n[dim]{i+1}. {display_time}[/dim] [bold blue]Command Generation[/bold blue] [dim]({provider}/{model}, {duration_ms/1000:.2f}s, {cached})[/dim]")
            console.print(f"[bold cyan]Prompt:[/bold cyan] {prompt}")
            console.print(f"[bold green]Result:[/bold green] {result}")

        elif entry_type == 'command_action':
            data = entry.get('data', {})
            action = data.get('action', 'unknown')
            command = data.get('processed_command', '')
            params = data.get('parameters', {})

            console.print(f"\n[dim]{i+1}. {display_time}[/dim] [bold magenta]Command Action[/bold magenta] [dim]({action})[/dim]")
            console.print(f"[bold cyan]Command:[/bold cyan] {command}")
            if params:
                console.print(f"[bold yellow]Parameters:[/bold yellow] {json.dumps(params)}")

        elif entry_type == 'command_execution':
            data = entry.get('data', {})
            command = data.get('command', '')
            exit_code = data.get('exit_code', -1)
            duration = data.get('duration_ms', 0) / 1000

            console.print(f"\n[dim]{i+1}. {display_time}[/dim] [bold green]Command Execution[/bold green] [dim](exit: {exit_code}, {duration:.2f}s)[/dim]")
            console.print(f"[bold cyan]Command:[/bold cyan] {command}")

        else:
            console.print(f"\n[dim]{i+1}. {display_time}[/dim] [bold]{entry_type}[/bold]")
            console.print(json.dumps(entry.get('data', {}), indent=2))

@cli.command(name='logs')
@click.option('--limit', '-l', type=int, default=50, help='Number of log lines to display')
@click.option('--level', type=click.Choice(['DEBUG', 'INFO', 'WARNING', 'ERROR', 'CRITICAL']),
              default='INFO', help='Minimum log level to display')
@click.option('--clear', is_flag=True, help='Clear log file')
@click.option('--export', '-e', help='Export logs to file path')
def logs_command(limit, level, clear, export):
    """View or manage application logs."""
    from ..utils.logging import get_logs, LOG_FILE
    import logging

    # Map string levels to integers
    level_map = {
        'DEBUG': logging.DEBUG,
        'INFO': logging.INFO,
        'WARNING': logging.WARNING,
        'ERROR': logging.ERROR,
        'CRITICAL': logging.CRITICAL
    }
    log_level = level_map.get(level, logging.INFO)

    # Handle clearing logs
    if clear:
        if not click.confirm("Are you sure you want to clear the log file?", default=False):
            console.print("[yellow]Operation cancelled.[/yellow]")
            return

        try:
            # Truncate log file
            open(LOG_FILE, 'w').close()
            console.print(f"[green]Log file cleared: {LOG_FILE}[/green]")
            return
        except Exception as e:
            console.print(f"[bold red]Error clearing logs:[/bold red] {e}")
            return

    # Get log entries
    log_entries = get_logs(max_lines=limit, level=log_level)

    if not log_entries:
        console.print("[yellow]No log entries found.[/yellow]")
        return

    # Handle exporting logs
    if export:
        try:
            with open(export, 'w') as f:
                f.writelines(log_entries)
            console.print(f"[green]Exported {len(log_entries)} log entries to {export}[/green]")
            return
        except Exception as e:
            console.print(f"[bold red]Error exporting logs:[/bold red] {e}")
            return

    # Display logs
    console.print(f"[bold]Last {len(log_entries)} log entries ([{level}] and above):[/bold]\n")

    # Format and colorize log entries
    for entry in log_entries:
        # Simple colorization based on log level
        if "[ERROR]" in entry or "[CRITICAL]" in entry:
            console.print(f"[red]{entry.strip()}[/red]")
        elif "[WARNING]" in entry:
            console.print(f"[yellow]{entry.strip()}[/yellow]")
        elif "[INFO]" in entry:
            console.print(f"[blue]{entry.strip()}[/blue]")
        else:
            console.print(f"[dim]{entry.strip()}[/dim]")

@cli.command(name='cache')
@click.option('--clear', is_flag=True, help='Clear all cached responses')
@click.option('--clear-older-than', type=int, help='Clear responses older than N hours')
@click.option('--stats', is_flag=True, help='Show cache statistics')
def cache_command(clear, clear_older_than, stats):
    """Manage LLM response cache."""
    cache = LLMResponseCache()
    cache_dir = cache.cache_dir

    # Handle clear requests first
    if clear:
        if not click.confirm("Are you sure you want to clear the entire response cache?", default=False):
            console.print("[yellow]Operation cancelled.[/yellow]")
            return

        count = cache.clear()
        console.print(f"[green]Cleared {count} cached responses.[/green]")
        return

    # Handle clear older than
    if clear_older_than is not None:
        # Convert hours to seconds
        max_age_seconds = clear_older_than * 3600
        count = cache.clear(max_age=max_age_seconds)
        console.print(f"[green]Cleared {count} cached responses older than {clear_older_than} hours.[/green]")
        return

    # Show stats by default
    total_files = sum(1 for _ in cache_dir.glob("*.json"))
    total_size = sum(f.stat().st_size for f in cache_dir.glob("*.json") if f.is_file())

    console.print("[bold]Cache Statistics:[/bold]\n")
    console.print(f"[bold cyan]Cache directory:[/bold cyan] {cache_dir}")
    console.print(f"[bold cyan]Total entries:[/bold cyan] {total_files}")
    console.print(f"[bold cyan]Total size:[/bold cyan] {total_size / 1024:.2f} KB")

    # Get models from cache entries
    models = {}
    total_age = 0
    newest_time = 0
    oldest_time = float('inf')

    for cache_file in cache_dir.glob("*.json"):
        try:
            with open(cache_file, "r") as f:
                entry = json.load(f)

            model = entry.get('model', 'unknown')
            if model in models:
                models[model] += 1
            else:
                models[model] = 1

            timestamp = entry.get('timestamp', 0)
            age = time.time() - timestamp
            total_age += age

            if timestamp > newest_time:
                newest_time = timestamp
            if timestamp < oldest_time:
                oldest_time = timestamp

        except Exception:
            continue

    # Show models
    if models:
        console.print("\n[bold cyan]Entries by model:[/bold cyan]")
        for model, count in models.items():
            console.print(f"  {model}: {count}")

    # Show age stats
    if total_files > 0:
        avg_age = total_age / total_files
        console.print(f"\n[bold cyan]Average age:[/bold cyan] {avg_age / 3600:.2f} hours")

        if newest_time > 0:
            newest_date = datetime.datetime.fromtimestamp(newest_time).strftime("%Y-%m-%d %H:%M:%S")
            console.print(f"[bold cyan]Newest entry:[/bold cyan] {newest_date}")

        if oldest_time < float('inf'):
            oldest_date = datetime.datetime.fromtimestamp(oldest_time).strftime("%Y-%m-%d %H:%M:%S")
            console.print(f"[bold cyan]Oldest entry:[/bold cyan] {oldest_date}")

@cli.command(name='generate', help="Generate gcloud commands from natural language")
@click.argument('prompt', nargs=-1, required=False)
@add_common_options
@click.option('--no-cache', is_flag=True, help='Disable response caching')
def main(prompt, model, api_key, verbose, no_cache):
    """InfraGPT - Convert natural language to Google Cloud commands."""
    # Initialize config file if it doesn't exist
    init_config()

    if verbose:
        from importlib.metadata import version
        try:
            console.print(f"[dim]InfraGPT version: {version('infragpt')}[/dim]")
        except:
            console.print("[dim]InfraGPT: Version information not available[/dim]")

    # Determine whether to use cache
    use_cache = not no_cache

    # If no prompt was provided, enter interactive mode
    if not prompt:
        interactive_mode(model, api_key, verbose, use_cache)
    else:
        user_prompt = " ".join(prompt)
        with console.status("[bold green]Generating command...[/bold green]", spinner="dots"):
            result = generate_gcloud_command(
                user_prompt,
                model,
                api_key,
                verbose,
                use_cache=use_cache
            )

        handle_command_result(result, model, verbose)

def interactive_mode(model_type: Optional[MODEL_TYPE] = None, api_key: Optional[str] = None,
                    verbose: bool = False, use_cache: bool = True):
    """Run InfraGPT in interactive mode with enhanced prompting."""
    import pathlib
    import time
    from prompt_toolkit import PromptSession
    from prompt_toolkit.history import FileHistory
    from prompt_toolkit.styles import Style

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
    config = Config()
    credentials = config.get_credentials()
    actual_model = credentials.get('model') if not model_type else model_type

    # Welcome message
    from rich.panel import Panel
    from rich.text import Text

    console.print(Panel.fit(
        Text("InfraGPT - Convert natural language to gcloud commands", style="bold green"),
        border_style="blue"
    ))

    # If no model configured, prompt for credentials now
    if not actual_model:
        from ..ui.prompts import prompt_credentials
        model_type, api_key = prompt_credentials()
        actual_model = model_type

    console.print(f"[yellow]Using model:[/yellow] [bold]{actual_model}[/bold]")
    console.print(f"[yellow]Cache:[/yellow] {'enabled' if use_cache else 'disabled'}")
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
                result = generate_gcloud_command(
                    user_input,
                    model_type,
                    api_key,
                    verbose,
                    use_cache=use_cache
                )

            handle_command_result(result, model_type, verbose)
        except KeyboardInterrupt:
            # Clear the current line and show a new prompt
            console.print("\n[yellow]Input cleared. Enter a new prompt:[/yellow]")
            continue
        except EOFError:
            # Exit on Ctrl+D
            console.print("\n[bold]Exiting InfraGPT.[/bold]")
            sys.exit(0)
