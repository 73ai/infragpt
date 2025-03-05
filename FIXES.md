# InfraGPT CLI Command Conflicts Fix

## Issue Summary

This fix addresses a critical issue where the `history` command was not working properly when executed with a command like `python -m infragpt history --limit 1`. This was happening because there were two competing CLI implementations:

1. A legacy implementation in `main.py` with a simpler history command
2. A newer, modular implementation in the `cli/` package with an advanced history command

When using the module directly (`python -m infragpt`), the code was importing from the legacy implementation instead of the modular one. This was causing conflicts and preventing proper access to the improved history commands.

## Changes Made

The main changes involved explicitly importing the CLI components from the modular implementation:

1. Updated import paths in multiple files:
   - `__main__.py`: Changed from `.cli import cli` to `.cli.commands import cli`
   - `__init__.py`: Changed from `.cli import cli, main` to `.cli.commands import cli, main`
   - `bin/infragpt`: Changed from `infragpt.cli import cli, main` to `infragpt.cli.commands import cli, main`
   - `bin/launcher.py`: Changed from `..main import cli, main` to `..cli.commands import cli, main`

2. Fixed function name in entry points:
   - Changed `main` to `main_launcher` in `bin/launcher.py`
   - Updated `setup.py` to reference the correct function

3. Fixed import issues in the core implementation:
   - Updated `core/__init__.py` to import `get_provider` and `get_credentials` instead of `get_llm`
   - Fixed `ui/prompts.py` to use the provider system instead of the old LLM system

## Testing

A test script (`test_history.py`) was created to verify the fix. The test confirmed that the history command now works correctly, showing that the newer implementation is being used.

## Results

After these changes, the `history` command now works correctly when running:
```
python -m infragpt history --limit 1
```

The command successfully displays the history entries with all the advanced features from the modular implementation.

## Additional Benefits

1. These changes help maintain consistency between different ways of invoking the CLI:
   - Running as a Python module: `python -m infragpt`
   - Running as an installed script: `infragpt`
   - Running from source: `./bin/infragpt`

2. The command path is now more explicitly directed to the current implementation, making future maintenance easier.