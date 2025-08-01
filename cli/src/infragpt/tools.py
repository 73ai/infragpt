#!/usr/bin/env python3
"""
Tool definitions for InfraGPT CLI agent using modern LangChain 2025 standards.
"""

from typing import Optional
from langchain_core.tools import tool

from infragpt.shell import CommandExecutor

# Initialize command executor
_executor = CommandExecutor()


@tool
def execute_shell_command(command: str, description: Optional[str] = None) -> str:
    """Execute a shell command and return the output.
    
    Use this to run system commands, check file contents, manage processes, etc.
    
    Args:
        command: The shell command to execute
        description: A brief description of what this command does (for user confirmation)
        
    Returns:
        The command output or error message
    """
    try:
        exit_code, output, was_cancelled = _executor.execute_command(command)
        
        if was_cancelled:
            return "Command was cancelled by user"
        elif exit_code == 0:
            return output
        else:
            return f"Command failed with exit code {exit_code}:\n{output}"
    except Exception as e:
        return f"Error executing command: {str(e)}"


def get_tools():
    """Get the list of available tools for LLM binding."""
    return [execute_shell_command]