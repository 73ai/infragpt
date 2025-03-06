#!/usr/bin/env python3

import re
import json
from typing import List, Dict, Tuple, Any, Union, Optional

from rich.console import Console
from rich.panel import Panel
from rich.prompt import Prompt, Confirm
from langchain_core.prompts import ChatPromptTemplate

from infragpt.config import CLIPBOARD_AVAILABLE
from infragpt.history import log_interaction
from infragpt.llm import MODEL_TYPE, get_parameter_info

# Initialize console for rich output
console = Console()

def create_prompt():
    """Create the prompt template for generating cloud commands."""
    template = """You are InfraGPT, a specialized assistant that helps users convert their natural language requests into
appropriate Google Cloud (gcloud) CLI commands.

INSTRUCTIONS:
1. Analyze the user's input to understand the intended cloud operation.
2. If the request is valid and related to Google Cloud operations, respond with ONLY the appropriate gcloud command(s).
3. If the operation requires multiple commands, separate them with a newline.
4. Include parameter placeholders in square brackets like [PROJECT_ID], [TOPIC_NAME], [SUBSCRIPTION_NAME], etc.
5. Do not include any explanations, markdown formatting, or additional text in your response.

Examples:
- Request: "Create a new VM instance called test-instance with 2 CPUs in us-central1-a"
  Response: gcloud compute instances create test-instance --machine-type=e2-medium --zone=us-central1-a

- Request: "Give viewer permissions to user@example.com for a pubsub topic"
  Response: gcloud pubsub topics add-iam-policy-binding [TOPIC_NAME] --member=user:user@example.com --role=roles/pubsub.viewer

- Request: "Create a VM instance and attach a new disk to it"
  Response: gcloud compute instances create [INSTANCE_NAME] --zone=[ZONE] --machine-type=e2-medium
gcloud compute disks create [DISK_NAME] --size=200GB --zone=[ZONE]
gcloud compute instances attach-disk [INSTANCE_NAME] --disk=[DISK_NAME] --zone=[ZONE]

- Request: "What's the weather like today?"
  Response: Request cannot be fulfilled.

User request: {prompt}

Your gcloud command(s):"""
    
    return ChatPromptTemplate.from_template(template)

def create_parameter_prompt():
    """Create prompt template for extracting parameter info from a command."""
    template = """You are InfraGPT Parameter Helper, a specialized assistant that helps users understand Google Cloud CLI command parameters.

TASK:
Analyze the Google Cloud CLI command below and provide information about each parameter that needs to be filled in.
For each parameter in square brackets like [PARAMETER_NAME], provide:
1. A brief description of what this parameter is
2. Examples of valid values
3. Any constraints or requirements

Format your response as JSON with the parameter name as key, like this:
```json
{{
  "PARAMETER_NAME": {{
    "description": "Brief description of the parameter",
    "examples": ["example1", "example2"], 
    "required": true,
    "default": "default value if any, otherwise null"
  }}
}}
```

Command: {command}

Parameter JSON:"""
    
    return ChatPromptTemplate.from_template(template)

def parse_command_parameters(command: str) -> Tuple[str, Dict[str, str], List[str]]:
    """Parse a command to extract its parameters and bracket placeholders."""
    # Extract base command and arguments
    parts = command.split()
    base_command = []
    
    params = {}
    current_param = None
    bracket_params = []
    
    for part in parts:
        # Extract parameters in square brackets (could be in any part of the command)
        bracket_matches = re.findall(r'\[([A-Z_]+)\]', part)
        if bracket_matches:
            for match in bracket_matches:
                bracket_params.append(match)
            
        if part.startswith('--'):
            # Handle --param=value format
            if '=' in part:
                param_name, param_value = part.split('=', 1)
                params[param_name[2:]] = param_value
            else:
                current_param = part[2:]
                params[current_param] = None
        elif current_param is not None:
            # This is a value for the previous parameter
            params[current_param] = part
            current_param = None
        else:
            # This is part of the base command
            base_command.append(part)
    
    return ' '.join(base_command), params, bracket_params

def prompt_for_parameters(command: str, model_type: MODEL_TYPE, return_params: bool = False) -> Union[str, Tuple[str, Dict[str, str]]]:
    """Prompt the user for each parameter in the command with AI assistance."""
    # Show the original command template first
    console.print("\n[bold blue]Command template:[/bold blue]")
    console.print(Panel(command, border_style="blue"))

    # Parse command to get base command, existing params, and placeholder params
    base_command, params, bracket_params = parse_command_parameters(command)
    
    # If command contains bracket params, get parameter info from LLM
    parameter_info = {}
    if bracket_params:
        parameter_info = get_parameter_info(command, model_type)
    
    # If no parameters of any kind, just return the command as is
    if not params and not bracket_params:
        if return_params:
            return command, {}
        return command
    
    # First handle bracket parameters with a separate section
    collected_params = {}
    
    if bracket_params:
        console.print("\n[bold magenta]Command requires the following parameters:[/bold magenta]")
        
        # Replace bracket parameters in base command and all params
        command_with_replacements = command
        
        for param in bracket_params:
            info = parameter_info.get(param, {})
            description = info.get('description', f"Value for {param}")
            examples = info.get('examples', [])
            default = info.get('default', None)
            
            # Create a rich prompt with available info
            prompt_text = f"[bold cyan]{param}[/bold cyan]"
            if description:
                prompt_text += f"\n  [dim]{description}[/dim]"
            if examples:
                examples_str = ", ".join([str(ex) for ex in examples])
                prompt_text += f"\n  [dim]Examples: {examples_str}[/dim]"
                
            # Get user input for this parameter
            value = Prompt.ask(prompt_text, default=default or "")
            
            # Store parameter value
            collected_params[param] = value
            
            # Replace all occurrences of [PARAM] with the value
            command_with_replacements = command_with_replacements.replace(f"[{param}]", value)
        
        # Now we have a command with all bracket params replaced
        if return_params:
            return command_with_replacements, collected_params
        return command_with_replacements
    
    # If we just have regular parameters (no brackets), handle them normally
    console.print("\n[bold yellow]Command parameters:[/bold yellow]")
    
    # Prompt for each parameter
    updated_params = {}
    for param, default_value in params.items():
        prompt_text = f"[bold cyan]{param}[/bold cyan]"
        if default_value:
            prompt_text += f" [default: {default_value}]"
        
        value = Prompt.ask(prompt_text, default=default_value or "")
        updated_params[param] = value
        collected_params[param] = value
    
    # Reconstruct command
    reconstructed_command = base_command
    for param, value in updated_params.items():
        if value:  # Only add non-empty parameters
            reconstructed_command += f" --{param}={value}"
    
    if return_params:
        return reconstructed_command, collected_params
    return reconstructed_command

def split_commands(result: str) -> List[str]:
    """Split multiple commands from the response."""
    if "Request cannot be fulfilled." in result:
        return [result]
    
    # Split by newlines and filter out empty lines
    commands = [cmd.strip() for cmd in result.splitlines() if cmd.strip()]
    return commands

def handle_command_result(result: str, model_type: Optional[MODEL_TYPE] = None, verbose: bool = False):
    """Handle the generated command results with options to print, copy, or execute."""
    try:
        import pyperclip
    except ImportError:
        pass

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
    
    for i, command in enumerate(commands):
        if verbose or len(commands) > 1:
            console.print(f"\n[bold cyan]Command {i+1} of {len(commands)}:[/bold cyan]")
            
        # Check if command has parameters and prompt for them
        if '[' in command or '--' in command:
            processed_command, params = prompt_for_parameters(command, model_type, return_params=True)
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
                import pyperclip
                pyperclip.copy(command)
                console.print("[bold green]Command copied to clipboard![/bold green]")
            except Exception as e:
                console.print(f"[bold red]Failed to copy to clipboard: {e}[/bold red]")
                console.print("[dim]You can manually copy the command above.[/dim]")
        elif choice == "run":
            console.print("\n[bold yellow]Executing command...[/bold yellow]")
            import os
            import datetime
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