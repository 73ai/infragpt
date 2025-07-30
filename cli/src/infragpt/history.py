#!/usr/bin/env python3

import os
import json
import uuid
import datetime
import pathlib
from typing import List, Dict, Any, Optional

import click
from rich.console import Console

# Initialize console for rich output
console = Console()

# Path to history directory
HISTORY_DIR = pathlib.Path.home() / ".config" / "infragpt" / "history"
HISTORY_DB_FILE = HISTORY_DIR / "history.jsonl"

def log_interaction(interaction_type: str, data: Dict[str, Any]):
    """Log user interaction to the history database file."""
    try:
        # Ensure history directory exists
        HISTORY_DIR.mkdir(parents=True, exist_ok=True)
        
        # Prepare the history entry
        entry = {
            "id": str(uuid.uuid4()),
            "timestamp": datetime.datetime.now().isoformat(),
            "type": interaction_type,
            "data": data
        }
        
        # Append to history file
        with open(HISTORY_DB_FILE, "a") as f:
            f.write(json.dumps(entry) + "\n")
            
    except Exception as e:
        # Silently fail - history logging should not interrupt user flow
        if 'verbose' in data and data.get('verbose'):
            console.print(f"[dim]Warning: Could not log interaction: {e}[/dim]")

def get_interaction_history(limit: int = 100) -> List[Dict[str, Any]]:
    """Retrieve the most recent interaction history entries."""
    if not HISTORY_DB_FILE.exists():
        return []
        
    try:
        entries = []
        with open(HISTORY_DB_FILE, "r") as f:
            for line in f:
                if line.strip():
                    entries.append(json.loads(line))
        
        # Return most recent entries first
        return list(reversed(entries[-limit:]))
    except Exception as e:
        console.print(f"[yellow]Warning:[/yellow] Could not read history: {e}")
        return []

def init_history_dir():
    """Initialize history directory if it doesn't exist."""
    HISTORY_DIR.mkdir(parents=True, exist_ok=True)

def display_history_entry(i: int, entry: Dict[str, Any]):
    """Format and display a history entry."""
    entry_type = entry.get('type', 'unknown')
    timestamp = entry.get('timestamp', '')
    timestamp_short = timestamp.split('T')[0] if 'T' in timestamp else timestamp
    
    if entry_type == 'command_generation':
        data = entry.get('data', {})
        model = data.get('model', 'unknown')
        prompt = data.get('prompt', '')
        result = data.get('result', '')
        
        console.print(f"\n[dim]{i+1}. {timestamp_short}[/dim] [bold blue]Command Generation[/bold blue] [dim]({model})[/dim]")
        console.print(f"[bold cyan]Prompt:[/bold cyan] {prompt}")
        console.print(f"[bold green]Result:[/bold green] {result}")
        
    elif entry_type == 'command_action':
        data = entry.get('data', {})
        action = data.get('action', 'unknown')
        command = data.get('processed_command', '')
        params = data.get('parameters', {})
        
        console.print(f"\n[dim]{i+1}. {timestamp_short}[/dim] [bold magenta]Command Action[/bold magenta] [dim]({action})[/dim]")
        console.print(f"[bold cyan]Command:[/bold cyan] {command}")
        if params:
            console.print(f"[bold yellow]Parameters:[/bold yellow] {json.dumps(params)}")
            
    elif entry_type == 'command_execution':
        data = entry.get('data', {})
        command = data.get('command', '')
        exit_code = data.get('exit_code', -1)
        duration = data.get('duration_ms', 0) / 1000
        
        console.print(f"\n[dim]{i+1}. {timestamp_short}[/dim] [bold green]Command Execution[/bold green] [dim](exit: {exit_code}, {duration:.2f}s)[/dim]")
        console.print(f"[bold cyan]Command:[/bold cyan] {command}")
    
    else:
        console.print(f"\n[dim]{i+1}. {timestamp_short}[/dim] [bold]{entry_type}[/bold]")
        console.print(json.dumps(entry.get('data', {}), indent=2))

def history_command(limit: int = 10, type: Optional[str] = None, export: Optional[str] = None):
    """View or export interaction history."""
    # Ensure history directory exists
    if not HISTORY_DB_FILE.exists():
        console.print("[yellow]No history found.[/yellow]")
        return
        
    # Read history
    entries = get_interaction_history(limit=limit)
    
    if not entries:
        console.print("[yellow]No history entries found.[/yellow]")
        return
    
    # Filter by type if specified
    if type:
        entries = [entry for entry in entries if entry.get('type') == type]
        if not entries:
            console.print(f"[yellow]No history entries found with type '{type}'.[/yellow]")
            return
    
    # Export if requested
    if export:
        try:
            with open(export, 'w') as f:
                for entry in entries:
                    f.write(json.dumps(entry) + '\n')
            console.print(f"[green]Exported {len(entries)} history entries to {export}[/green]")
            return
        except Exception as e:
            console.print(f"[bold red]Error exporting history:[/bold red] {e}")
            return
    
    # Display history
    console.print(f"[bold]Last {len(entries)} interaction(s):[/bold]")
    
    for i, entry in enumerate(entries):
        display_history_entry(i, entry)