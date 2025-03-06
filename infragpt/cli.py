#!/usr/bin/env python3

import os
import sys
import re
import yaml
import json
import uuid
import datetime
from typing import Literal, Optional, List, Dict, Tuple, Any, Union

import click
from prompt_toolkit import PromptSession
from prompt_toolkit.history import FileHistory
from prompt_toolkit.styles import Style
from rich.console import Console
from rich.panel import Panel
from rich.text import Text
from rich.prompt import Prompt, Confirm
import pathlib

try:
    import pyperclip
    CLIPBOARD_AVAILABLE = True
except ImportError:
    CLIPBOARD_AVAILABLE = False

from langchain_openai import ChatOpenAI
from langchain_anthropic import ChatAnthropic
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.output_parsers import StrOutputParser

# Initialize console for rich output
console = Console()

# Define type for model selection
MODEL_TYPE = Literal["gpt4o", "claude"]

# Path to config directory
CONFIG_DIR = pathlib.Path.home() / ".config" / "infragpt"
CONFIG_FILE = CONFIG_DIR / "config.yaml"
HISTORY_DIR = CONFIG_DIR / "history"
HISTORY_DB_FILE = HISTORY_DIR / "history.jsonl"

def load_config():
    """Load configuration from config file."""
    if not CONFIG_FILE.exists():
        return {}
    
    try:
        with open(CONFIG_FILE, "r") as f:
            return yaml.safe_load(f) or {}
    except Exception as e:
        console.print(f"[yellow]Warning:[/yellow] Could not load config: {e}")
        return {}

def save_config(config):
    """Save configuration to config file."""
    # Ensure directory exists
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    
    try:
        with open(CONFIG_FILE, "w") as f:
            yaml.dump(config, f)
    except Exception as e:
        console.print(f"[yellow]Warning:[/yellow] Could not save config: {e}")

def get_credentials(model_type: Optional[MODEL_TYPE] = None, api_key: Optional[str] = None, verbose: bool = False):
    """
    Get API credentials based on priority:
    1. Command line parameters
    2. Stored config
    3. Environment variables
    4. Interactive prompt
    """
    config = load_config()
    
    # Priority 1: Command line parameters
    if model_type and api_key and api_key.strip():  # Ensure API key is not empty
        # Update config for future use
        config["model"] = model_type
        config["api_key"] = api_key
        save_config(config)
        return model_type, api_key
    
    # Priority 2: Check stored config
    if config.get("model") and config.get("api_key") and config.get("api_key").strip():  # Ensure API key is not empty
        if verbose:
            console.print(f"[dim]Using credentials from config file[/dim]")
        return config["model"], config["api_key"]
    
    # Priority 3: Check environment variables
    openai_key = os.getenv("OPENAI_API_KEY")
    anthropic_key = os.getenv("ANTHROPIC_API_KEY")
    env_model = os.getenv("INFRAGPT_MODEL")
    
    # Command line model takes precedence over env var model
    resolved_model = model_type or env_model
    
    # Validate environment credentials
    if anthropic_key and openai_key:
        # If both keys are provided, use the model to decide
        if resolved_model == "claude":
            if verbose:
                console.print(f"[dim]Using Anthropic API key from environment[/dim]")
            # Save to config for future use
            config["model"] = "claude"
            config["api_key"] = anthropic_key
            save_config(config)
            return "claude", anthropic_key
        elif resolved_model == "gpt4o":
            if verbose:
                console.print(f"[dim]Using OpenAI API key from environment[/dim]")
            # Save to config for future use
            config["model"] = "gpt4o"
            config["api_key"] = openai_key
            save_config(config)
            return "gpt4o", openai_key
        elif not resolved_model:
            # Default to OpenAI if model not specified
            if verbose:
                console.print(f"[dim]Multiple API keys found, defaulting to OpenAI[/dim]")
            # Save to config for future use
            config["model"] = "gpt4o"
            config["api_key"] = openai_key
            save_config(config)
            return "gpt4o", openai_key
    elif anthropic_key:
        if resolved_model and resolved_model != "claude":
            console.print("[bold red]Error:[/bold red] Anthropic API key is set but model is not claude.")
            sys.exit(1)
        if verbose:
            console.print(f"[dim]Using Anthropic API key from environment[/dim]")
        # Save to config for future use
        config["model"] = "claude"
        config["api_key"] = anthropic_key
        save_config(config)
        return "claude", anthropic_key
    elif openai_key:
        if resolved_model and resolved_model != "gpt4o":
            console.print("[bold red]Error:[/bold red] OpenAI API key is set but model is not gpt4o.")
            sys.exit(1)
        if verbose:
            console.print(f"[dim]Using OpenAI API key from environment[/dim]")
        # Save to config for future use
        config["model"] = "gpt4o"
        config["api_key"] = openai_key
        save_config(config)
        return "gpt4o", openai_key
    
    # Priority 4: Prompt user interactively
    console.print("\n[bold yellow]API credentials required[/bold yellow]")
    
    # If model is provided, use that, otherwise prompt for model choice
    if not model_type:
        model_options = ["gpt4o", "claude"]
        model_type = Prompt.ask(
            "[bold cyan]Select model[/bold cyan]",
            choices=model_options,
            default="gpt4o"
        )
    
    # Prompt for API key based on model
    provider = "OpenAI" if model_type == "gpt4o" else "Anthropic"
    api_key = Prompt.ask(
        f"[bold cyan]Enter your {provider} API key[/bold cyan] [dim](will be saved in {CONFIG_FILE})[/dim]",
        password=True
    )
    
    # Save credentials for future use
    config["model"] = model_type
    config["api_key"] = api_key
    save_config(config)
    
    return model_type, api_key

def validate_api_key(model_type: MODEL_TYPE, api_key: str) -> bool:
    """Validate if the API key is correct by making a minimal API call."""
    try:
        if model_type == "gpt4o":
            # Create a minimal OpenAI client to validate the key
            llm = ChatOpenAI(
                model="gpt-4o", 
                temperature=0, 
                api_key=api_key,
                max_tokens=5  # Minimal response to reduce token usage
            )
            # Make a minimal request
            response = llm.invoke("Say OK")
            return True
        elif model_type == "claude":
            # Create a minimal Anthropic client to validate the key
            llm = ChatAnthropic(
                model="claude-3-sonnet-20240229", 
                temperature=0, 
                api_key=api_key,
                max_tokens=5  # Minimal response to reduce token usage
            )
            # Make a minimal request
            response = llm.invoke("Say OK")
            return True
        else:
            raise ValueError(f"Unsupported model type: {model_type}")
    except Exception as e:
        if "API key" in str(e) or "auth" in str(e).lower() or "key" in str(e).lower() or "token" in str(e).lower():
            console.print(f"[bold red]Invalid API key:[/bold red] {e}")
            return False
        else:
            # If the error is not related to authentication, re-raise it
            console.print(f"[bold yellow]Warning:[/bold yellow] API connection error: {e}")
            # For other errors, we still allow the key - it might be a temporary issue
            return True

def get_llm(model_type: Optional[MODEL_TYPE] = None, api_key: Optional[str] = None, verbose: bool = False, validate: bool = True):
    """Initialize the appropriate LLM based on user selection."""
    # Get credentials and actual model type
    resolved_model, resolved_api_key = get_credentials(model_type, api_key, verbose)
    
    # Validate API key if requested
    if validate:
        # If key is invalid, prompt for a new one
        while not validate_api_key(resolved_model, resolved_api_key):
            console.print("[bold red]API key validation failed.[/bold red]")
            resolved_model, resolved_api_key = prompt_credentials(resolved_model)
    
    if resolved_model == "gpt4o":
        return ChatOpenAI(model="gpt-4o", temperature=0, api_key=resolved_api_key)
    elif resolved_model == "claude":
        return ChatAnthropic(model="claude-3-sonnet-20240229", temperature=0, api_key=resolved_api_key)
    else:
        raise ValueError(f"Unsupported model type: {resolved_model}")

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

def get_parameter_info(command: str, model_type: MODEL_TYPE) -> Dict[str, Dict[str, Any]]:
    """Get information about parameters from the LLM."""
    # Extract parameters that need filling in (those in square brackets)
    bracket_params = re.findall(r'\[([A-Z_]+)\]', command)
    
    if not bracket_params:
        return {}
    
    # Create a prompt to get parameter info
    llm = get_llm(model_type)
    prompt_template = create_parameter_prompt()
    
    # Create and execute the chain
    chain = prompt_template | llm | StrOutputParser()
    
    with console.status("[bold blue]Analyzing command parameters...[/bold blue]", spinner="dots"):
        result = chain.invoke({"command": command})
    
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

def generate_gcloud_command(prompt: str, model_type: Optional[MODEL_TYPE] = None, api_key: Optional[str] = None, verbose: bool = False) -> str:
    """Generate a gcloud command based on the user's natural language prompt."""
    # Initialize the LLM and get the actual model type used
    # Always validate the API key on the first real command generation
    llm = get_llm(model_type, api_key, verbose, validate=True)
    
    # Get the actual model being used from llm configuration or config file
    actual_model = model_type
    if not actual_model:
        config = load_config()
        actual_model = config.get("model", "unknown")
    
    if verbose and actual_model:
        console.print(f"[dim]Generating command using {actual_model}...[/dim]")
    
    # Create the prompt
    prompt_template = create_prompt()
    
    # Create and execute the chain
    chain = prompt_template | llm | StrOutputParser()
    start_time = datetime.datetime.now()
    result = chain.invoke({"prompt": prompt})
    end_time = datetime.datetime.now()
    
    # Log the interaction for future intelligence
    try:
        interaction_data = {
            "model": actual_model,
            "prompt": prompt,
            "result": result.strip(),
            "duration_ms": (end_time - start_time).total_seconds() * 1000,
            "verbose": verbose
        }
        log_interaction("command_generation", interaction_data)
    except Exception:
        # Log failures should not interrupt the flow
        pass
    
    return result.strip()

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

def interactive_mode(model_type: Optional[MODEL_TYPE] = None, api_key: Optional[str] = None, verbose: bool = False):
    """Run InfraGPT in interactive mode with enhanced prompting."""
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
        Text("InfraGPT - Convert natural language to gcloud commands", style="bold green"),
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

def validate_env_api_keys():
    """Validate API keys from environment variables and prompt if invalid."""
    openai_key = os.getenv("OPENAI_API_KEY")
    anthropic_key = os.getenv("ANTHROPIC_API_KEY")
    env_model = os.getenv("INFRAGPT_MODEL")
    
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

def init_config():
    """Initialize configuration file with environment variables if it doesn't exist."""
    if CONFIG_FILE.exists():
        return
    
    # Create config directory if it doesn't exist
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    
    # Create history directory too
    HISTORY_DIR.mkdir(parents=True, exist_ok=True)
    
    config = {}
    
    # Check for environment variables to populate initial config
    openai_key = os.getenv("OPENAI_API_KEY")
    anthropic_key = os.getenv("ANTHROPIC_API_KEY")
    env_model = os.getenv("INFRAGPT_MODEL")
    
    # Validate environment variable API keys
    model, api_key = validate_env_api_keys()
    
    # If we got valid credentials from validation, save those
    if model and api_key:
        config["model"] = model
        config["api_key"] = api_key
    # Otherwise use the original environment variables
    elif anthropic_key and (not env_model or env_model == "claude"):
        config["model"] = "claude"
        config["api_key"] = anthropic_key
    elif openai_key and (not env_model or env_model == "gpt4o"):
        config["model"] = "gpt4o"
        config["api_key"] = openai_key
    
    # Save config if we have anything to save
    if config:
        save_config(config)

@click.group(invoke_without_command=True)
@click.pass_context
@click.version_option(package_name='infragpt')
def cli(ctx):
    """InfraGPT - Convert natural language to Google Cloud commands and manage history."""
    # If no subcommand is specified, go to interactive mode
    if ctx.invoked_subcommand is None:
        ctx.invoke(main, prompt=())

@cli.command(name='history')
@click.option('--limit', '-l', type=int, default=10, help='Number of history entries to display')
@click.option('--type', '-t', help='Filter by interaction type (e.g., command_generation, command_action, command_execution)')
@click.option('--export', '-e', help='Export history to file path')
def history_command(limit, type, export):
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

@cli.command(name='generate', help="Generate gcloud commands from natural language")
@click.argument('prompt', nargs=-1, required=False)
@click.option('--model', '-m', type=click.Choice(['gpt4o', 'claude']), 
              help='LLM model to use (gpt4o or claude)')
@click.option('--api-key', '-k', help='API key for the selected model')
@click.option('--verbose', '-v', is_flag=True, help='Enable verbose output')
def main(prompt, model, api_key, verbose):
    """InfraGPT - Convert natural language to Google Cloud commands."""
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
    
    # If no prompt was provided, enter interactive mode
    if not prompt:
        interactive_mode(model, api_key, verbose)
    else:
        user_prompt = " ".join(prompt)
        with console.status("[bold green]Generating command...[/bold green]", spinner="dots"):
            result = generate_gcloud_command(user_prompt, model, api_key, verbose)
        
        handle_command_result(result, model, verbose)

if __name__ == "__main__":
    cli()