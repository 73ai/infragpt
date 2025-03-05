"""Display formatting for InfraGPT."""

import os
import datetime
from typing import List, Dict, Any, Optional, Literal

from rich.panel import Panel
from rich.prompt import Prompt, Confirm

from .console import console
from ..core.commands import split_commands
from ..core.llm import MODEL_TYPE
from ..utils.history import log_interaction

try:
    import pyperclip
    CLIPBOARD_AVAILABLE = True
except ImportError:
    CLIPBOARD_AVAILABLE = False

def handle_command_result(result: str, model_type: Optional[MODEL_TYPE] = None, verbose: bool = False):
    """Handle the generated command results with options to print, copy, or execute."""
    commands = split_commands(result)
    
    if not commands:
        console.print("[bold red]No valid commands generated[/bold red]")
        return
    
    # If it's an error response, just display it
    if commands[0] == "Request cannot be fulfilled.":
        console.print(f"[bold red]{commands[0]}[/bold red]")
        return
    
    # Show the number of commands if multiple
    if len(commands) > 1:
        console.print(f"\n[bold blue]Generated {len(commands)} commands:[/bold blue]")
        for i, cmd in enumerate(commands):
            console.print(f"[dim]{i+1}.[/dim] [italic]{cmd.split()[0]}...[/italic]")
        console.print()
    
    # Process each command
    processed_commands = []
    parameter_values = {}
    
    # Import here to avoid circular imports
    from .prompts import prompt_for_parameters
    
    for i, command in enumerate(commands):
        if verbose or len(commands) > 1:
            console.print(f"\n[bold cyan]Command {i+1} of {len(commands)}:[/bold cyan]")
            
        # Check if command has parameters and prompt for them
        if '[' in command or '--' in command:
            # Split this into two parts:
            # 1. First display the command template and parse parameter needs
            # 2. Then use the spinner only for parameter info fetching, not for prompting
            from infragpt.ui.prompts import parse_command_parameters, get_parameter_info, prompt_for_parameters
            
            # First display the command template
            console.print("\n[bold blue]Command template:[/bold blue]")
            console.print(Panel(command, border_style="blue"))
            
            # Parse parameters
            base_command, params, bracket_params = parse_command_parameters(command)
            
            # Get parameter info with spinner if needed
            parameter_info = {}
            if bracket_params:
                with console.status("[bold blue]Analyzing command parameters...[/bold blue]", spinner="dots"):
                    raw_param_info = get_parameter_info(command, model_type)
                    
                    # Convert raw parameter info to expected format
                    parameter_info = {}
                    for param in bracket_params:
                        # Create default info
                        param_value = {
                            "description": f"Value for {param}",
                            "examples": [],
                            "required": True,
                            "default": None
                        }
                        
                        # Check if param name is in the response
                        if param in raw_param_info:
                            info = raw_param_info[param]
                            if isinstance(info, dict):
                                if "description" in info:
                                    param_value["description"] = info["description"]
                                if "examples" in info and isinstance(info["examples"], list):
                                    param_value["examples"] = info["examples"]
                                if "required" in info:
                                    param_value["required"] = info["required"]
                                if "default" in info and info["default"] is not None:
                                    param_value["default"] = info["default"]
                        
                        # Store the parameter info
                        parameter_info[param] = param_value
            
            # Now prompt for parameters (no spinner during prompting)
            console.print("\n[bold magenta]Command requires the following parameters:[/bold magenta]")
            processed_command, params = prompt_for_parameters(command, model_type, return_params=True, 
                                                             pre_analyzed_params=parameter_info)
            processed_commands.append(processed_command)
            parameter_values[f"command_{i+1}"] = params
            console.print(Panel(processed_command, border_style="green", title=f"Final Command {i+1}"))
        else:
            processed_commands.append(command)
            parameter_values[f"command_{i+1}"] = {}
            console.print(Panel(command, border_style="green", title=f"Command {i+1}"))
    
    # Set choices to just copy and run, with copy as default
    choices = []
    if CLIPBOARD_AVAILABLE:
        choices.append("copy")
    choices.append("run")
    
    # If nothing is available, add print option
    if not choices:
        choices.append("print")
    
    # Default to copy if available, otherwise first option
    default = "copy" if CLIPBOARD_AVAILABLE else choices[0]
    
    # For each command, ask what to do
    for i, command in enumerate(processed_commands):
        if len(commands) > 1:
            console.print(f"\n[bold cyan]Action for command {i+1}:[/bold cyan]")
            console.print(Panel(command, border_style="blue"))
        
        # Use rich to display options and get choice
        choice = Prompt.ask(
            "[bold yellow]What would you like to do with this command?[/bold yellow]",
            choices=choices,
            default=default
        )
        
        # Log the user's choice and the parameters they provided
        try:
            action_data = {
                "command_index": i,
                "original_command": commands[i],
                "processed_command": command,
                "parameters": parameter_values.get(f"command_{i+1}", {}),
                "action": choice,
                "model": model_type,
                "verbose": verbose
            }
            log_interaction("command_action", action_data)
        except Exception:
            # Log failures should not interrupt the flow
            pass
        
        if choice == "copy" and CLIPBOARD_AVAILABLE:
            try:
                pyperclip.copy(command)
                console.print("[bold green]Command copied to clipboard![/bold green]")
            except Exception as e:
                console.print(f"[bold red]Failed to copy to clipboard: {e}[/bold red]")
                console.print("[dim]You can manually copy the command above.[/dim]")
        elif choice == "run":
            console.print("\n[bold yellow]Executing command...[/bold yellow]")
            start_time = datetime.datetime.now()
            try:
                exit_code = os.system(command)
                end_time = datetime.datetime.now()
                
                # Log command execution
                try:
                    execution_data = {
                        "command": command,
                        "exit_code": exit_code,
                        "duration_ms": (end_time - start_time).total_seconds() * 1000,
                        "parameters": parameter_values.get(f"command_{i+1}", {}),
                        "model": model_type,
                        "verbose": verbose
                    }
                    log_interaction("command_execution", execution_data)
                except Exception:
                    pass
                
            except Exception as e:
                console.print(f"[bold red]Error executing command: {e}[/bold red]")
            
            if i < len(processed_commands) - 1:
                # Ask if they want to continue with the next command
                if not Confirm.ask("[bold yellow]Continue with the next command?[/bold yellow]", default=True):
                    break