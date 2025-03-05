# InfraGPT Testing Framework

This directory contains tests for the InfraGPT project, organized by module.

## Test Structure

- `utils/`: Tests for utility modules (config, history, etc.)
- `core/`: Tests for core functionality (commands, LLM integration, etc.)
- `cli/`: Tests for command-line interface components

## Running Tests

To run all tests:

```bash
pytest
```

To run tests with coverage report:

```bash
pytest --cov=infragpt --cov-report=term-missing
```

To run a specific test file:

```bash
pytest tests/unit/utils/test_config.py
```

## Test Configuration

Configuration for pytest is defined in `pyproject.toml`. Key settings include:

- Test discovery paths
- Coverage reporting
- Test file naming patterns

## Fixtures

Common fixtures are defined in `conftest.py` and include:

- `temp_config_dir`: Creates a temporary directory for test configuration files
- `mock_config_file`: Creates a mock configuration file with test values
- `mock_environment`: Sets up environment variables for testing
- `mock_history_file`: Creates a mock history file with test entries

## Mocking Strategy

Tests use pytest's monkeypatch and unittest.mock to:

1. Mock filesystem operations
2. Mock configuration files and environment variables
3. Mock LLM API calls to avoid actual API usage
4. Mock user input/output for CLI testing

## Coverage Goals

The test suite aims for high coverage across:

- Configuration management
- API credential handling
- Command parsing and generation
- User interaction flows
- Error handling and edge cases

## Adding New Tests

When adding new functionality to InfraGPT, please also add corresponding tests that:

1. Test the happy path (normal operation)
2. Test edge cases and error handling
3. Mock external dependencies appropriately
4. Validate both inputs and outputs

Follow the existing naming conventions and organization to maintain consistency.