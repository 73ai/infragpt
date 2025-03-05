# InfraGPT Project Notes

## Project Structure
- Main modules:
  - `infragpt/cli/`: Command-line interface components
  - `infragpt/core/`: Core functionality (LLM integration, prompts)
  - `infragpt/utils/`: Utility functions (config, history)
  - `infragpt/ui/`: User interface components

## Refactoring Status
- Completed phases:
  - Module restructuring
  - Configuration management
  - History storage enhancement
  - LLM integration refinement (provider system, caching)
  - Command processing pipeline
  - Error handling and logging
  - Testing framework setup with pytest
- Pending phases:
  - Performance optimizations
  - Expanded test coverage

## Key Improvements
- **LLM Provider System**:
  - Abstract base class for providers
  - Concrete implementations for OpenAI and Anthropic
  - Response caching with TTL and management features
  - Better credential handling via Config class

- **Command Pipeline Architecture**:
  - Component-based pipeline with standardized interfaces
  - Modular processing steps (input processing, command generation, parsing, parameter handling)
  - Extensible with pre-processors and post-processors
  - Context-based data sharing between pipeline components
  - Support for custom component implementations

- **Error Handling and Logging System**:
  - Custom exception hierarchy with domain-specific error types
  - Standardized error handling with structured error details
  - Comprehensive logging with multiple levels (DEBUG, INFO, WARNING, ERROR)
  - Log file management with rotation
  - Error boundary pattern for isolating and gracefully handling failures
  - Global exception handler for improved application stability

- **CLI Enhancements**:
  - New `cache` command for managing LLM response cache
  - New `logs` command for viewing and managing application logs
  - CLI flags for disabling cache (`--no-cache`)

## Commands
- Testing: `python -m pytest tests`
- Testing with coverage: `python -m pytest --cov=infragpt --cov-report=term-missing`
- Linting: `flake8 infragpt tests`
- Type checking: `mypy infragpt`
- Building: `python -m build`
- Running: `python -m infragpt`

## Code Style Preferences
- Black with line length of 100 characters
- isort for import sorting with black profile
- Use type hints for all function parameters and return values
- Docstrings in Google style format
- Comprehensive error handling with custom exception types
- Clear separation of concerns between modules
- Explicit over implicit
- Favor composition over inheritance
- Use meaningful variable and function names

## Rich Console UI Lessons
- **Status Display Management**:
  - Rich library only allows one live display (status spinner) to be active at once
  - Attempting to use nested status displays causes the error: `rich.errors.LiveError: Only one live display may be active at once`
  - Status spinners should only be used at a single level in the call hierarchy
  - Never have a status spinner active during user input prompts

- **User Experience Flow**:
  - Command generation: Follow a consistent user flow for better experience
  - Proper sequence: display command template → analyze parameters with spinner → prompt for values
  - Separate parameter analysis (which uses a spinner) from parameter prompting (which requires input)

- **Parameter Handling**:
  - LLM may not always return parameter info in the expected format
  - Add conversion logic to normalize parameter info for consistent display
  - Required parameters should be validated to ensure non-empty values
  - Clearly display parameter descriptions and examples in the UI

- **CLI Interface Design**:
  - Use status spinners only for background processing operations
  - Keep user interaction separate from spinner displays
  - Follow consistent patterns for command generation and parameter handling

## Known Issues and Future Improvements
- **Parameter extraction**: Improve parameter information extraction from LLM responses 
- **Error handling**: Add graceful recovery for parameter analysis failures
- **Parameter validation**: Implement more sophisticated validation for different parameter types
- **UI responsiveness**: Ensure status displays never interfere with input prompts
- **Enhanced feedback**: Add better progress indication during long-running operations
