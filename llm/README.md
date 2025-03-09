# InfraGPT LLM Module

This module provides shared LLM (Large Language Model) functionality that can be used by both the InfraGPT CLI client and server components.

## Overview

The LLM module centralizes functionality for interacting with language models, providing a clean interface for model operations, API key validation, and prompt handling.

## Components

- **client.py** - Core LLM client functionality
  - `get_llm_client()` - Creates and returns a model client
  - `generate_gcloud_command()` - Generates gcloud commands from natural language
  - `get_parameter_info()` - Extracts parameter information from commands

- **models.py** - Type definitions and model mappings
  - `MODEL_TYPE` - Type definition for supported models
  - `MODEL_MAP` - Mapping from model types to specific model identifiers
  - `DEFAULT_PARAMS` - Default parameters for each model type

- **auth.py** - API key validation
  - `validate_api_key()` - Validates API keys against the provider

- **prompts.py** - Prompt templates and handling
  - `get_prompt_template()` - Retrieves prompt templates by name
  - `format_prompt()` - Formats prompt templates with variables

- **errors.py** - Custom exception classes
  - `LLMError` - Base exception class
  - `AuthenticationError` - Authentication-related errors
  - `GenerationError` - Command generation errors
  - `ParsingError` - Response parsing errors

## Usage Examples

### Generating Commands

```python
from llm import generate_gcloud_command

command = generate_gcloud_command(
    prompt="Create a VM instance in us-central1",
    model_type="gpt4o",
    api_key="your-api-key"
)
print(command)
```

### Working with Model Clients

```python
from llm import get_llm_client, MODEL_TYPE

# Get a model client
client = get_llm_client(
    model_type="claude",
    api_key="your-api-key",
    temperature=0.2
)

# Use the client directly
response = client.invoke("Generate a bash script to backup files")
```

### Validating API Keys

```python
from llm import validate_api_key

is_valid = validate_api_key(
    model_type="gpt4o",
    api_key="your-api-key"
)

if is_valid:
    print("API key is valid")
else:
    print("API key is invalid")
```

## Error Handling

```python
from llm import generate_gcloud_command, AuthenticationError, GenerationError

try:
    command = generate_gcloud_command(
        prompt="Create a VM instance",
        model_type="gpt4o",
        api_key="invalid-key"
    )
except AuthenticationError:
    print("Authentication failed")
except GenerationError as e:
    print(f"Command generation failed: {e}")
```