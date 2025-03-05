"""
Helper functions and simplified versions of actual code for testing.
These are isolated versions of functions to avoid dependency issues.
"""
import json
import pathlib
import yaml
from typing import Dict, Any, List


# --- Config Helpers ---

def load_config_test(config_file: pathlib.Path) -> Dict[str, Any]:
    """Simplified load_config for testing."""
    if not config_file.exists():
        return {}
    
    try:
        with open(config_file, "r") as f:
            return yaml.safe_load(f) or {}
    except Exception:
        return {}


def save_config_test(config: Dict[str, Any], config_file: pathlib.Path, config_dir: pathlib.Path) -> None:
    """Simplified save_config for testing."""
    # Ensure directory exists
    config_dir.mkdir(parents=True, exist_ok=True)
    
    try:
        with open(config_file, "w") as f:
            yaml.dump(config, f)
    except Exception:
        pass  # Silently fail


# --- History Helpers ---

def log_interaction_test(
    interaction_type: str, 
    data: Dict[str, Any], 
    history_file: pathlib.Path, 
    history_dir: pathlib.Path
) -> None:
    """Simplified log_interaction for testing."""
    try:
        # Ensure history directory exists
        history_dir.mkdir(parents=True, exist_ok=True)
        
        # Prepare the history entry
        entry = {
            "id": "test-id",
            "timestamp": "2023-01-01T00:00:00",
            "type": interaction_type,
            "data": data
        }
        
        # Append to history file
        with open(history_file, "a") as f:
            f.write(json.dumps(entry) + "\n")
            
    except Exception:
        # Silently fail - history logging should not interrupt user flow
        pass


def get_interaction_history_test(history_file: pathlib.Path, limit: int = 100) -> List[Dict[str, Any]]:
    """Simplified get_interaction_history for testing."""
    if not history_file.exists():
        return []
        
    try:
        entries = []
        with open(history_file, "r") as f:
            for line in f:
                if line.strip():
                    entries.append(json.loads(line))
        
        # Return most recent entries first
        return list(reversed(entries[-limit:]))
    except Exception:
        return []


# --- Command Helpers ---

def split_commands_test(result: str) -> List[str]:
    """Simplified split_commands for testing."""
    if "Request cannot be fulfilled." in result:
        return [result]
    
    # Split by newlines and filter out empty lines
    commands = [cmd.strip() for cmd in result.splitlines() if cmd.strip()]
    return commands


def parse_command_parameters_test(command: str):
    """Simplified parse_command_parameters for testing."""
    # Extract base command and arguments
    parts = command.split()
    base_command = []
    
    params = {}
    current_param = None
    bracket_params = []
    
    for part in parts:
        # Extract parameters in square brackets
        if '[' in part and ']' in part:
            # Extract the parameter name within brackets
            start_idx = part.find('[')
            end_idx = part.find(']')
            if start_idx != -1 and end_idx != -1:
                param_name = part[start_idx+1:end_idx]
                bracket_params.append(param_name)
            
        if part.startswith('--'):
            # Handle --param=value format
            if '=' in part:
                param_name, param_value = part.split('=', 1)
                params[param_name[2:]] = param_value
                
                # Check if the value contains bracketed parameters
                if '[' in param_value and ']' in param_value:
                    start_idx = param_value.find('[')
                    end_idx = param_value.find(']')
                    if start_idx != -1 and end_idx != -1:
                        nested_param = param_value[start_idx+1:end_idx]
                        bracket_params.append(nested_param)
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