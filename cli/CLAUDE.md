# InfraGPT CLI Service

The InfraGPT CLI is an interactive command-line tool that converts natural language descriptions into Google Cloud Platform (gcloud) commands using large language models (LLMs). It provides a seamless interface for DevOps engineers and cloud administrators to generate, validate, and execute infrastructure commands.

## Purpose and Overview

The CLI service serves as the primary user interface for InfraGPT, offering:

- **Natural Language Processing**: Convert plain English descriptions into executable gcloud commands
- **Interactive Command Generation**: Real-time command generation with user feedback
- **Multi-Model Support**: Compatible with OpenAI GPT-4o and Anthropic Claude models
- **Parameter Management**: Intelligent parameter prompting and validation
- **Command History**: Persistent logging of interactions for analysis and reuse
- **Execution Options**: Copy to clipboard or direct command execution

## Architecture

### Core Components

- **`cli.py`**: Main CLI interface with Click framework integration
- **`llm.py`**: LLM adapter layer interfacing with shared `/llm/` module
- **`config.py`**: Configuration management and credential handling
- **`prompts.py`**: Command processing and parameter handling logic
- **`history.py`**: Interaction logging and history management
- **`__main__.py`**: Entry point for CLI execution

### Key Design Patterns

1. **Adapter Pattern**: `cli/llm.py` acts as an adapter to the shared LLM module
2. **Command Pattern**: Interactive prompting with parameter collection
3. **Strategy Pattern**: Multiple credential sources with priority hierarchy
4. **Observer Pattern**: History logging for all user interactions

## How to Run the CLI

### Prerequisites

- Python 3.9 or higher
- Virtual environment (recommended)
- API key for OpenAI (GPT-4o) or Anthropic (Claude)

### Installation Methods

#### Development Installation
```bash
# Clone repository
git clone https://github.com/priyanshujain/infragpt.git
cd infragpt

# Install in development mode
pip install -e .
```

#### Production Installation
```bash
# Using pipx (recommended)
pipx install infragpt

# Or using pip
pip install infragpt
```

### Execution

#### Interactive Mode (Default)
```bash
# Basic usage
infragpt

# With specific model
infragpt --model claude
infragpt --model gpt4o

# With API key
infragpt --model gpt4o --api-key "your-openai-key"

# With verbose output
infragpt --verbose
```

#### History Commands
```bash
# View recent history
infragpt history

# Limit number of entries
infragpt history --limit 20

# Filter by interaction type
infragpt history --type command_execution

# Export history to file
infragpt history --export history.jsonl
```

### Configuration

The CLI uses a cascading configuration system:

1. **Command-line parameters** (highest priority)
2. **Configuration file** (`~/.config/infragpt/config.yaml`)
3. **Environment variables**
4. **Interactive prompts** (lowest priority)

#### Environment Variables
```bash
export OPENAI_API_KEY="your-openai-key"
export ANTHROPIC_API_KEY="your-anthropic-key"
export INFRAGPT_MODEL="gpt4o"  # or "claude"
```

#### Configuration File Location
- **Config**: `~/.config/infragpt/config.yaml`
- **History**: `~/.config/infragpt/history/history.jsonl`
- **Prompt History**: `~/.infragpt/history`

## How to Test the CLI

### Unit Testing Framework

The CLI service uses pytest for testing with the following structure:

```bash
# Run all CLI tests
python -m pytest cli/tests/ -v

# Run specific test categories
python -m pytest cli/tests/test_config.py -v
python -m pytest cli/tests/test_llm.py -v
python -m pytest cli/tests/test_prompts.py -v

# Run with coverage
python -m pytest cli/tests/ --cov=cli --cov-report=html
```

### Test Categories

1. **Configuration Tests** (`test_config.py`)
   - Config file loading/saving
   - Credential hierarchy validation
   - Environment variable handling

2. **LLM Integration Tests** (`test_llm.py`)
   - API key validation
   - Model switching logic
   - Error handling for invalid credentials

3. **Prompt Processing Tests** (`test_prompts.py`)
   - Parameter extraction and prompting
   - Command parsing and reconstruction
   - Multi-command handling

4. **History Tests** (`test_history.py`)
   - Interaction logging
   - History retrieval and filtering
   - Export functionality

5. **CLI Interface Tests** (`test_cli.py`)
   - Click command parsing
   - Interactive mode simulation
   - Command-line argument validation

### Manual Testing

#### Test Scenarios
```bash
# Test basic command generation
echo "create a VM called test-vm in us-central1" | infragpt

# Test parameter handling
echo "create a GKE cluster with 3 nodes" | infragpt

# Test history functionality
infragpt history --limit 5

# Test configuration handling
infragpt --model claude --verbose
```

#### Integration Testing with Mock API
```bash
# Set up test environment
export OPENAI_API_KEY="test-key"
export INFRAGPT_TEST_MODE="true"

# Run integration tests
python -m pytest cli/tests/integration/ -v
```

## Development Standards and Conventions

### Code Style and Standards

#### Python Standards
- **PEP 8**: Follow Python style guidelines
- **Type Hints**: Use type annotations for all function signatures
- **Docstrings**: Google-style docstrings for all public functions
- **Error Handling**: Explicit exception handling with user-friendly messages

#### Import Organization
```python
# Standard library imports
import os
import sys
from typing import Optional, Dict, Any

# Third-party imports
import click
from rich.console import Console

# Local imports
from cli.config import load_config
from cli.llm import generate_gcloud_command
```

#### Rich Console Usage
- Use Rich library for all user-facing output
- Consistent color scheme:
  - **Blue**: Information and prompts
  - **Green**: Success messages
  - **Yellow**: Warnings
  - **Red**: Errors
  - **Cyan**: User input prompts
  - **Dim**: Secondary information

### Architecture Patterns

#### Error Handling Strategy
```python
def function_with_error_handling():
    """Example of proper error handling pattern."""
    try:
        # Main logic
        result = risky_operation()
        return result
    except SpecificException as e:
        console.print(f"[bold red]Error:[/bold red] {e}")
        return None
    except Exception as e:
        console.print(f"[bold yellow]Warning:[/bold yellow] Unexpected error: {e}")
        return None
```

#### Configuration Pattern
```python
def get_config_value(key: str, default: Any = None) -> Any:
    """Standard pattern for configuration access."""
    config = load_config()
    return config.get(key, default)
```

#### Logging Pattern
```python
def operation_with_logging(data: Dict[str, Any]):
    """Example of logging pattern."""
    try:
        # Perform operation
        result = perform_operation(data)
        
        # Log success
        log_interaction("operation_success", {
            "input": data,
            "result": result,
            "timestamp": datetime.now().isoformat()
        })
        
        return result
    except Exception as e:
        # Log failure (but don't interrupt flow)
        log_interaction("operation_failure", {
            "input": data,
            "error": str(e),
            "timestamp": datetime.now().isoformat()
        })
        raise
```

### Dependencies and Requirements

#### Core Dependencies
- **click**: Command-line interface framework
- **rich**: Terminal formatting and user interface
- **prompt-toolkit**: Interactive prompting with history
- **pyperclip**: Clipboard integration (optional)
- **pyyaml**: Configuration file handling

#### Shared Dependencies
- **LLM Module**: Uses `/llm/` for model interactions
- **LangChain**: Through shared LLM module for prompt templates

### File Structure Standards

```
cli/
├── __init__.py          # Package initialization
├── __main__.py          # CLI entry point
├── cli.py               # Main CLI interface
├── config.py            # Configuration management
├── llm.py               # LLM adapter layer
├── prompts.py           # Command processing logic
├── history.py           # Interaction logging
├── tests/               # Test suite
│   ├── test_cli.py
│   ├── test_config.py
│   ├── test_llm.py
│   ├── test_prompts.py
│   └── test_history.py
└── CLAUDE.md           # This documentation
```

### CLI-Specific Patterns

#### Interactive Flow Pattern
1. **Initialization**: Load configuration and validate credentials
2. **Prompt Loop**: Accept user input with rich formatting
3. **Processing**: Generate commands using LLM integration
4. **Parameter Collection**: Prompt for required parameters
5. **Action Selection**: Offer copy/execute options
6. **Logging**: Record all interactions for history

#### Command Generation Workflow
```python
def command_generation_workflow(user_input: str, model: str) -> str:
    """Standard workflow for command generation."""
    # 1. Validate inputs
    if not user_input.strip():
        return ""
    
    # 2. Generate command
    with console.status("Generating command..."):
        command = generate_gcloud_command(user_input, model)
    
    # 3. Process parameters
    if has_parameters(command):
        command = prompt_for_parameters(command, model)
    
    # 4. Log interaction
    log_interaction("command_generation", {
        "input": user_input,
        "output": command,
        "model": model
    })
    
    return command
```

## CLI Service Implementation Details

### Entry Point Configuration
The CLI is configured as a console script in `pyproject.toml`:
```toml
[project.scripts]
infragpt = "cli.cli:cli"
```

### Credential Management Priority
1. Command-line `--api-key` parameter
2. Saved configuration file
3. Environment variables (`OPENAI_API_KEY`, `ANTHROPIC_API_KEY`)
4. Interactive prompt with validation

### History and Analytics
- All interactions are logged to JSONL format
- Includes timestamps, model used, input/output, and execution results
- Supports filtering and export for analysis
- Privacy-conscious: logs are stored locally only

This CLI service provides a robust, user-friendly interface for infrastructure command generation while maintaining high code quality standards and comprehensive testing coverage.