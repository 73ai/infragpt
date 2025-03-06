#!/usr/bin/env python3

import os
import sys
import datetime
from typing import Literal, Optional, Dict, Any

from rich.console import Console
from rich.prompt import Prompt
from langchain_openai import ChatOpenAI
from langchain_anthropic import ChatAnthropic
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.output_parsers import StrOutputParser

from infragpt.config import (
    CONFIG_FILE, load_config, save_config
)
from infragpt.history import log_interaction

# Initialize console for rich output
console = Console()

# Define type for model selection
MODEL_TYPE = Literal["gpt4o", "claude"]

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

def generate_gcloud_command(prompt: str, model_type: Optional[MODEL_TYPE] = None, api_key: Optional[str] = None, verbose: bool = False) -> str:
    """Generate a gcloud command based on the user's natural language prompt."""
    from infragpt.prompts import create_prompt
    
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

def get_parameter_info(command: str, model_type: MODEL_TYPE) -> Dict[str, Dict[str, Any]]:
    """Get information about parameters from the LLM."""
    import re
    from infragpt.prompts import create_parameter_prompt
    
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