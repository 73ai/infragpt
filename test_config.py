#!/usr/bin/env python3
"""Test script to see why the config isn't being loaded correctly."""

import os
import sys
import json
import yaml
import pathlib

def main():
    """Check the config and print diagnostics."""
    print("Config diagnosis script")
    
    # Import path info from the package
    from infragpt.utils.config import CONFIG_DIR, CONFIG_FILE, Config
    
    print(f"\nConfig paths:")
    print(f"  CONFIG_DIR: {CONFIG_DIR}")
    print(f"  CONFIG_FILE: {CONFIG_FILE}")
    
    # Check if config file exists
    print(f"\nConfig file exists: {CONFIG_FILE.exists()}")
    
    # Try to load and print the config content
    if CONFIG_FILE.exists():
        try:
            with open(CONFIG_FILE, "r") as f:
                raw_config = f.read()
                print(f"\nRaw config content:")
                print(raw_config)
                
                try:
                    config_data = yaml.safe_load(raw_config)
                    print(f"\nParsed config data:")
                    print(json.dumps(config_data, indent=2))
                except Exception as e:
                    print(f"Error parsing config: {e}")
        except Exception as e:
            print(f"Error reading config file: {e}")
    
    # Try using the Config class
    config = Config()
    config.load()
    
    print("\nLoaded config from Config class:")
    print(f"  Loaded status: {config._loaded}")
    
    if config._loaded:
        print(f"  Internal config data:")
        print(json.dumps(config._config, indent=2))
    
    # Try getting credentials
    credentials = config.get_credentials()
    print(f"\nCredentials from Config class:")
    print(json.dumps(credentials, indent=2))
    
    # Get the default profile
    print(f"\nDefault profile: {config.get_default_profile()}")
    
    # Try the API lookup chain
    from infragpt.core.llm import get_credentials
    
    print("\nTesting credential lookup chain:")
    model, api_key = get_credentials(verbose=True)
    print(f"  Resolved model: {model}")
    print(f"  API key (masked): {'*' * 10 if api_key else 'None'}")
    
    # Check environment variables
    print("\nEnvironment variables:")
    print(f"  OPENAI_API_KEY set: {'Yes' if os.environ.get('OPENAI_API_KEY') else 'No'}")
    print(f"  ANTHROPIC_API_KEY set: {'Yes' if os.environ.get('ANTHROPIC_API_KEY') else 'No'}")
    print(f"  INFRAGPT_MODEL set: {'Yes' if os.environ.get('INFRAGPT_MODEL') else 'No'}")
    
    return 0

if __name__ == "__main__":
    sys.exit(main())