"""
Global pytest fixtures and configuration.
"""
import json
import os
import pathlib
import tempfile
import yaml
from typing import Dict, Any, List, Tuple

import pytest


# --- Path Fixtures ---

@pytest.fixture
def temp_dir():
    """Create a temporary directory for test files."""
    with tempfile.TemporaryDirectory() as temp_dir:
        yield pathlib.Path(temp_dir)


# --- Config Fixtures ---

@pytest.fixture
def config_file(temp_dir):
    """Create a test config file with standard test configuration."""
    config_file = temp_dir / "config.yaml"
    test_config = {
        "model": "gpt4o",
        "api_key": "mock-api-key"
    }
    
    # Create directory if it doesn't exist
    config_file.parent.mkdir(parents=True, exist_ok=True)
    
    # Write config
    with open(config_file, "w") as f:
        yaml.dump(test_config, f)
    
    return config_file, test_config


@pytest.fixture
def mock_environment():
    """Setup mock environment variables for testing."""
    original_environ = os.environ.copy()
    
    # Set test environment variables
    os.environ.update({
        "OPENAI_API_KEY": "mock-openai-key",
        "ANTHROPIC_API_KEY": "mock-anthropic-key",
        "INFRAGPT_MODEL": "gpt4o"
    })
    
    yield
    
    # Restore original environment
    os.environ.clear()
    os.environ.update(original_environ)


# --- History Fixtures ---

@pytest.fixture
def history_entries() -> List[Dict[str, Any]]:
    """Return standard test history entries as parsed objects."""
    return [
        {
            "id": "test-id-1", 
            "timestamp": "2023-01-01T00:00:00", 
            "type": "command_generation", 
            "data": {"model": "gpt4o", "prompt": "test prompt 1", "result": "test result 1"}
        },
        {
            "id": "test-id-2", 
            "timestamp": "2023-01-02T00:00:00", 
            "type": "command_action", 
            "data": {"action": "copy", "command": "test command 2"}
        },
        {
            "id": "test-id-3", 
            "timestamp": "2023-01-03T00:00:00", 
            "type": "command_execution", 
            "data": {"command": "test command 3", "exit_code": 0}
        }
    ]


@pytest.fixture
def history_file(temp_dir, history_entries) -> Tuple[pathlib.Path, List[Dict[str, Any]]]:
    """Create a test history file with standard test entries."""
    history_dir = temp_dir / "history"
    history_dir.mkdir(parents=True, exist_ok=True)
    
    history_file = history_dir / "history.jsonl"
    
    # Convert entries to JSON strings
    entries_json = [json.dumps(entry) for entry in history_entries]
    
    with open(history_file, "w") as f:
        for entry in entries_json:
            f.write(entry + "\n")
    
    return history_file, history_entries


# --- LLM Response Fixtures ---

@pytest.fixture
def llm_parameter_response():
    """Return a mock LLM response for parameter information."""
    return """```json
{
  "INSTANCE_NAME": {
    "description": "Name of the virtual machine instance",
    "examples": ["my-instance", "test-vm"],
    "required": true,
    "default": null
  },
  "ZONE": {
    "description": "Compute zone to deploy the instance in",
    "examples": ["us-central1-a", "europe-west1-b"],
    "required": true,
    "default": null
  }
}
```"""