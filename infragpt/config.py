#!/usr/bin/env python3

import os
import yaml
import pathlib
from typing import Dict, Any

from rich.console import Console

try:
    import pyperclip
    CLIPBOARD_AVAILABLE = True
except ImportError:
    CLIPBOARD_AVAILABLE = False

# Initialize console for rich output
console = Console()

# Path to config directory
CONFIG_DIR = pathlib.Path.home() / ".config" / "infragpt"
CONFIG_FILE = CONFIG_DIR / "config.yaml"

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

def init_config():
    """Initialize configuration file with environment variables if it doesn't exist."""
    if CONFIG_FILE.exists():
        return
    
    # Create config directory if it doesn't exist
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    
    # Initialize history directory
    from infragpt.history import init_history_dir
    init_history_dir()
    
    config = {}
    
    # Importing here to avoid circular imports
    from infragpt.llm import validate_env_api_keys
    
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