"""Configuration handling for InfraGPT."""

import os
import yaml
import pathlib
from typing import Dict, Any, Optional, Literal, List

from ..ui.console import console
from .logging import logger, ConfigError, handle_error

# Define type for model selection
MODEL_TYPE = Literal["gpt4o", "claude"]

# Path to config directory
CONFIG_DIR = pathlib.Path.home() / ".config" / "infragpt"
CONFIG_FILE = CONFIG_DIR / "config.yaml"

class Config:
    """Configuration manager for InfraGPT."""
    
    def __init__(self):
        """Initialize configuration with defaults."""
        self._config = {
            "version": 1,
            "credentials": {
                "default": None,
                "profiles": {}
            },
            "settings": {
                "default_model": "gpt4o",
                "history_limit": 100,
                "verbose": False
            }
        }
        self._loaded = False
    
    def load(self) -> None:
        """Load configuration from disk."""
        if not CONFIG_FILE.exists():
            self._loaded = False
            logger.debug(f"Config file does not exist at {CONFIG_FILE}")
            return
        
        try:
            with open(CONFIG_FILE, "r") as f:
                loaded_config = yaml.safe_load(f) or {}
                
            # Handle version migration
            file_version = loaded_config.get('version', 0)
            if file_version < 1:
                logger.info(f"Migrating config from version {file_version} to version 1")
                # Migrate old format to new format
                if 'model' in loaded_config and 'api_key' in loaded_config:
                    # Create a default profile from old settings
                    self._config['credentials']['profiles']['default'] = {
                        'model': loaded_config['model'],
                        'api_key': loaded_config['api_key']
                    }
                    self._config['credentials']['default'] = 'default'
                # Keep version 1
            else:
                # Use the loaded config
                self._config = loaded_config
                
            self._loaded = True
            logger.debug("Configuration loaded successfully")
            
        except Exception as e:
            error_msg = f"Could not load config: {e}"
            logger.error(error_msg, exc_info=True)
            console.print(f"[yellow]Warning:[/yellow] {error_msg}")
            self._loaded = False
            # Don't raise the error to maintain backward compatibility
    
    def save(self) -> None:
        """Save configuration to disk."""
        # Ensure directory exists
        CONFIG_DIR.mkdir(parents=True, exist_ok=True)
        
        try:
            with open(CONFIG_FILE, "w") as f:
                yaml.dump(self._config, f)
            logger.debug("Configuration saved successfully")
        except Exception as e:
            error_msg = f"Could not save config: {e}"
            logger.error(error_msg, exc_info=True)
            console.print(f"[yellow]Warning:[/yellow] {error_msg}")
            # For errors during save, we could raise a ConfigError
            # but for backward compatibility, we'll just log it
    
    def get_profile_names(self) -> List[str]:
        """Get list of profile names."""
        return list(self._config.get('credentials', {}).get('profiles', {}).keys())
    
    def get_default_profile(self) -> Optional[str]:
        """Get the default profile name."""
        return self._config.get('credentials', {}).get('default')
    
    def set_default_profile(self, profile_name: str) -> None:
        """Set the default profile name."""
        if profile_name not in self.get_profile_names():
            error_msg = f"Profile '{profile_name}' does not exist"
            logger.error(error_msg)
            raise ConfigError(error_msg, details={"profile_name": profile_name, "available_profiles": self.get_profile_names()})
        
        logger.info(f"Setting default profile to '{profile_name}'")
        self._config.setdefault('credentials', {})['default'] = profile_name
        self.save()
    
    def get_credentials(self, profile: Optional[str] = None) -> Dict[str, Any]:
        """Get credentials for a specific profile or default."""
        if not self._loaded:
            self.load()
        
        if not profile:
            profile = self.get_default_profile()
        
        if not profile or profile not in self.get_profile_names():
            return {}
        
        return self._config.get('credentials', {}).get('profiles', {}).get(profile, {})
    
    def set_credentials(self, model: MODEL_TYPE, api_key: str, profile: str = 'default') -> None:
        """Set credentials for a specific profile."""
        if not self._loaded:
            self.load()
        
        # Ensure the credentials section exists
        self._config.setdefault('credentials', {}).setdefault('profiles', {})
        
        # Set the profile
        self._config['credentials']['profiles'][profile] = {
            'model': model,
            'api_key': api_key
        }
        
        # If this is the first profile, set it as default
        if not self.get_default_profile():
            self._config['credentials']['default'] = profile
        
        self.save()
    
    def get_setting(self, key: str, default: Any = None) -> Any:
        """Get a setting value."""
        if not self._loaded:
            self.load()
        
        return self._config.get('settings', {}).get(key, default)
    
    def set_setting(self, key: str, value: Any) -> None:
        """Set a setting value."""
        if not self._loaded:
            self.load()
        
        self._config.setdefault('settings', {})[key] = value
        self.save()
    
    def update_credentials(self, model: MODEL_TYPE, api_key: str, profile: str = 'default') -> None:
        """Update credentials for a specific profile.
        
        This is an alias for set_credentials, for backward compatibility.
        """
        return self.set_credentials(model, api_key, profile)

# Global config instance
config = Config()

def load_config() -> Dict[str, Any]:
    """Legacy function to load configuration from config file.
    
    This is kept for backward compatibility.
    """
    if not CONFIG_FILE.exists():
        return {}
    
    try:
        with open(CONFIG_FILE, "r") as f:
            return yaml.safe_load(f) or {}
    except Exception as e:
        console.print(f"[yellow]Warning:[/yellow] Could not load config: {e}")
        return {}

def save_config(config_data: Dict[str, Any]) -> None:
    """Legacy function to save configuration to config file.
    
    This is kept for backward compatibility.
    """
    # Ensure directory exists
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    
    try:
        with open(CONFIG_FILE, "w") as f:
            yaml.dump(config_data, f)
    except Exception as e:
        console.print(f"[yellow]Warning:[/yellow] Could not save config: {e}")

def init_config() -> None:
    """Initialize configuration file with environment variables if it doesn't exist."""
    # Load the global config object to ensure it's initialized
    if not config._loaded:
        config.load()
    
    # If credentials already exist, we're done
    if config.get_default_profile() and config.get_credentials():
        return
    
    # Check for environment variables to populate initial config
    openai_key = os.getenv("OPENAI_API_KEY")
    anthropic_key = os.getenv("ANTHROPIC_API_KEY")
    env_model = os.getenv("INFRAGPT_MODEL")
    
    # Validate environment variable API keys
    from .validation import validate_env_api_keys
    model, api_key = validate_env_api_keys()
    
    # If we got valid credentials from validation, save those
    if model and api_key:
        config.set_credentials(model, api_key)
    # Otherwise use the original environment variables
    elif anthropic_key and (not env_model or env_model == "claude"):
        config.set_credentials("claude", anthropic_key)
    elif openai_key and (not env_model or env_model == "gpt4o"):
        config.set_credentials("gpt4o", openai_key)