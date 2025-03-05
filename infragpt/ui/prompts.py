"""User prompting utilities for InfraGPT."""

import re
from typing import Dict, Any, Optional, Tuple, Union, List, Literal

from rich.prompt import Prompt
from rich.text import Text
from rich.panel import Panel
from rich.console import Console

from .console import console
from ..core.llm import MODEL_TYPE, validate_api_key
from ..utils.config import load_config, save_config

def prompt_credentials(existing_model: Optional[MODEL_TYPE] = None):
    """Prompt user for model and API key before starting."""
    if existing_model:
        console.print("\n[bold yellow]API key required. Please enter your credentials:[/bold yellow]")
        model_type = existing_model
    else:
        console.print("\n[bold yellow]No model configured. Please set up your credentials:[/bold yellow]")
        
        # Prompt for model choice
        model_options = ["gpt4o", "claude"]
        model_type = Prompt.ask(
            "[bold cyan]Select model[/bold cyan]",
            choices=model_options,
            default="gpt4o"
        )
    
    # Prompt for API key based on model
    provider = "OpenAI" if model_type == "gpt4o" else "Anthropic"
    
    valid_key = False
    while not valid_key:
        # Keep prompting until we get a non-empty API key
        api_key = ""
        while not api_key.strip():
            from ..utils.config import CONFIG_FILE
            api_key = Prompt.ask(
                f"[bold cyan]Enter your {provider} API key[/bold cyan] [dim](will be saved in {CONFIG_FILE})[/dim]",
                password=True
            )
            
            if not api_key.strip():
                console.print("[bold red]API key cannot be empty. Please try again.[/bold red]")
        
        # Validate the API key
        with console.status(f"[bold blue]Validating {provider} API key...[/bold blue]", spinner="dots"):
            valid_key = validate_api_key(model_type, api_key)
        
        if not valid_key:
            console.print("[bold red]Invalid API key. Please try again.[/bold red]")
    
    # Save credentials for future use
    config = load_config()
    config["model"] = model_type
    config["api_key"] = api_key
    save_config(config)
    
    console.print(f"[green]Credentials validated and saved successfully for {model_type}![/green]\n")
    return model_type, api_key

def get_parameter_info(command: str, model_type: MODEL_TYPE) -> Dict[str, Dict[str, Any]]:
    """Get information about parameters from the LLM."""
    # Extract parameters that need filling in (those in square brackets)
    bracket_params = re.findall(r'\[([A-Z_]+)\]', command)
    
    if not bracket_params:
        return {}
    
    # Import here to avoid circular imports
    from ..core.llm import get_provider
    
    # Create system prompt manually rather than using the template with examples
    system_prompt = """You are InfraGPT Parameter Helper, a specialized assistant that helps users understand Google Cloud CLI command parameters.

Analyze the Google Cloud CLI command below and provide information about each parameter that needs to be filled in.
For each parameter in square brackets like [PARAMETER_NAME], provide:
1. A brief description of what this parameter is
2. Examples of valid values
3. Any constraints or requirements

Format your response as JSON with the parameter name as key."""

    # Create a user prompt manually
    user_prompt = f"Command: {command}\n\nParameter JSON:"
    
    # Get provider
    provider = get_provider(model_type, validate=False)  # Skip validation here
    
    # Execute with provider's generate method without a status display
    # Status display is shown at the higher level in cli/commands.py
    result = provider.generate(user_prompt, system_prompt=system_prompt)
    
    # Extract the JSON part
    try:
        import json
        # Find JSON part between triple backticks if present
        if "```json" in result:
            json_part = result.split("```json")[1].split("```")[0].strip()
        elif "```" in result:
            json_part = result.split("```")[1].strip()
        else:
            json_part = result.strip()
        
        parameter_info = json.loads(json_part)
        return parameter_info
    except Exception as e:
        console.print(f"[bold yellow]Warning:[/bold yellow] Could not parse parameter info: {e}")
        return {}

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

def prompt_for_parameter(
    param: str, 
    description: Optional[str] = None, 
    examples: Optional[List[str]] = None, 
    default: Optional[str] = None,
    required: bool = True
) -> str:
    """Prompt the user for a parameter value with rich formatting.
    
    Args:
        param: The parameter name
        description: Description of the parameter
        examples: Example values for the parameter
        default: Default value if user provides no input
        required: Whether this parameter is required (cannot be empty)
        
    Returns:
        The user's input value for the parameter
    """
    # Create a rich prompt with available info
    prompt_text = f"[bold cyan]{param}[/bold cyan]"
    if description:
        prompt_text += f"\n  [dim]{description}[/dim]"
    if examples:
        examples_str = ", ".join([str(ex) for ex in examples])
        prompt_text += f"\n  [dim]Examples: {examples_str}[/dim]"
    
    # Get user input with validation if required
    while True:
        value = Prompt.ask(prompt_text, default=default or "")
        
        # Validate non-empty if required
        if required and not value.strip():
            console.print("[bold red]This parameter is required. Please enter a value.[/bold red]")
            continue
        break
        
    return value


def prompt_for_parameters(command: str, model_type: MODEL_TYPE, return_params: bool = False, 
                    pre_analyzed_params: Optional[Dict[str, Dict[str, Any]]] = None) -> Union[str, Tuple[str, Dict[str, str]]]:
    """Prompt the user for each parameter in the command with AI assistance.
    
    Args:
        command: The command containing parameters to prompt for
        model_type: The LLM model to use for parameter info
        return_params: Whether to return parameters along with the command
        pre_analyzed_params: Pre-analyzed parameter info (if already fetched)
    """
    # Parse command to get base command, existing params, and placeholder params
    base_command, params, bracket_params = parse_command_parameters(command)
    
    # If we already have pre-analyzed parameters, use those
    # Otherwise get parameter info from LLM (for backwards compatibility)
    parameter_info = {}
    if pre_analyzed_params:
        parameter_info = pre_analyzed_params
    elif bracket_params:
        # Get parameter info directly (legacy path)
        raw_param_info = get_parameter_info(command, model_type)
        
        # Convert raw parameter info to our expected format if needed
        parameter_info = {}
        for param in bracket_params:
            # Try to find info about this parameter in the raw data
            param_value = {
                "description": f"Value for {param}",
                "examples": [],
                "required": True,
                "default": None
            }
            
            # Check if the param name is directly in the response
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
            
            # Store the parameter info in our new dictionary
            parameter_info[param] = param_value
    
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
            
            # Check if parameter is required
            is_required = info.get('required', True)  # Default to required if not specified
            
            # Get user input for this parameter with validation
            while True:
                value = Prompt.ask(prompt_text, default=default or "")
                
                # Validate non-empty if required
                if is_required and not value.strip():
                    console.print("[bold red]This parameter is required. Please enter a value.[/bold red]")
                    continue
                break
            
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
        
        # Assume parameters with -- flags are required
        # Get user input with validation
        while True:
            value = Prompt.ask(prompt_text, default=default_value or "")
            
            # Validate non-empty if required
            if not value.strip():
                console.print("[bold red]This parameter is required. Please enter a value.[/bold red]")
                continue
            break
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