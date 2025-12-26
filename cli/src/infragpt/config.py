import os
import yaml
import pathlib
from typing import Any, Dict

from rich.console import Console

try:
    import pyperclip  # noqa: F401

    CLIPBOARD_AVAILABLE = True
except ImportError:
    CLIPBOARD_AVAILABLE = False

console = Console()

CONFIG_DIR = pathlib.Path.home() / ".config" / "infragpt"
CONFIG_FILE = CONFIG_DIR / "config.yaml"


def load_config() -> Dict[str, Any]:
    """Load configuration from config file."""
    if not CONFIG_FILE.exists():
        return {}

    try:
        with open(CONFIG_FILE, "r") as f:
            return yaml.safe_load(f) or {}
    except (yaml.YAMLError, OSError) as e:
        console.print(f"[yellow]Warning:[/yellow] Could not load config: {e}")
        return {}


def save_config(config: Dict[str, Any]) -> None:
    """Save configuration to config file."""
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)

    try:
        with open(CONFIG_FILE, "w") as f:
            yaml.dump(config, f)
    except (yaml.YAMLError, OSError) as e:
        console.print(f"[yellow]Warning:[/yellow] Could not save config: {e}")


def init_config() -> None:
    """Initialize configuration file with environment variables if it doesn't exist."""
    if CONFIG_FILE.exists():
        return

    CONFIG_DIR.mkdir(parents=True, exist_ok=True)

    from infragpt.history import init_history_dir

    init_history_dir()

    config = {}

    from infragpt.llm import validate_env_api_keys

    openai_key = os.getenv("OPENAI_API_KEY")
    anthropic_key = os.getenv("ANTHROPIC_API_KEY")
    env_model = os.getenv("INFRAGPT_MODEL")

    model, api_key = validate_env_api_keys()

    if model and api_key:
        config["model"] = model
        config["api_key"] = api_key
    elif anthropic_key and (not env_model or env_model == "claude"):
        config["model"] = "claude"
        config["api_key"] = anthropic_key
    elif openai_key and (not env_model or env_model == "gpt4o"):
        config["model"] = "gpt4o"
        config["api_key"] = openai_key

    if config:
        save_config(config)
