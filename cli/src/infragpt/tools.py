"""
Tool system without LangChain dependencies.
"""

import inspect
import json
from typing import Callable, Dict, Any, Optional, List
from functools import wraps
from rich.console import Console
from .shell import CommandExecutor


class ToolExecutionCancelled(Exception):
    """Raised when user cancels tool execution to break the agentic loop."""
    pass


console = Console()


def tool(name: Optional[str] = None, description: Optional[str] = None):
    """
    Decorator to convert a function into a tool definition.
    Replacement for LangChain's @tool decorator.
    """
    def decorator(func: Callable) -> Callable:
        # Get function signature
        sig = inspect.signature(func)
        
        # Build tool schema
        tool_name = name or func.__name__
        tool_description = description or func.__doc__ or f"Execute {tool_name}"
        
        # Build parameters schema from function signature
        parameters = {
            "type": "object",
            "properties": {},
            "required": []
        }
        
        for param_name, param in sig.parameters.items():
            param_type = "string"  # Default type
            param_description = f"Parameter {param_name}"
            
            # Try to infer type from annotation
            if param.annotation != inspect.Parameter.empty:
                if param.annotation == str:
                    param_type = "string"
                elif param.annotation == int:
                    param_type = "integer"
                elif param.annotation == float:
                    param_type = "number"
                elif param.annotation == bool:
                    param_type = "boolean"
            
            parameters["properties"][param_name] = {
                "type": param_type,
                "description": param_description
            }
            
            # Mark as required if no default value
            if param.default == inspect.Parameter.empty:
                parameters["required"].append(param_name)
        
        # Store tool metadata on function
        func._tool_schema = {
            "name": tool_name,
            "description": tool_description,
            "input_schema": parameters  # Anthropic format
        }
        
        func._tool_schema_openai = {
            "type": "function",
            "function": {
                "name": tool_name,
                "description": tool_description,
                "parameters": parameters  # OpenAI format
            }
        }
        
        @wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)
        
        return wrapper
    
    return decorator


@tool(name="execute_shell_command", description="Execute a shell command and return the output")
def execute_shell_command(command: str, description: Optional[str] = None) -> str:
    """
    Execute a shell command and return the output.
    
    Args:
        command: The shell command to execute
        description: Optional description of what the command does
        
    Returns:
        The command output or error message
    """
    console.print(f"\n[bold cyan]Tool Call: execute_shell_command[/bold cyan]")
    console.print(f"[dim]Command:[/dim] {command}")
    if description:
        console.print(f"[dim]Description:[/dim] {description}")
    
    # Ask for user confirmation with proper interrupt handling
    # Use standard input() for tool confirmation to ensure reliable Ctrl+C handling
    console.print(f"\n[yellow]Execute this command? (Y/n):[/yellow] ", end="")
    console.file.flush()
    
    try:
        user_input = input().strip().lower()
    except (KeyboardInterrupt, EOFError):
        console.print("\n[yellow]Command execution cancelled.[/yellow]")
        raise ToolExecutionCancelled("User cancelled command execution")
    
    # Check for explicit denial
    if user_input in ['n', 'no']:
        console.print("\n[yellow]Command execution cancelled.[/yellow]")
        raise ToolExecutionCancelled("User cancelled command execution")
    # Empty input defaults to "Yes" (execute command)
    
    # If we get here, user input is 'y', 'yes', or any other non-empty input - execute command
    
    try:
        # Use the proper shell executor for streaming output
        # Pass our console instance to ensure consistent output
        executor = CommandExecutor(timeout=60)
        
        console.print(f"\n[bold blue]Executing command...[/bold blue]")
        console.file.flush()  # Ensure prompt is displayed immediately
        
        # Execute the command - output should stream in real-time via shell.py console
        exit_code, output, was_cancelled = executor.execute_command(command)
        
        if was_cancelled:
            return "Command execution was cancelled by user."
        elif exit_code == 0:
            console.print(f"\n[green]✓ Command executed successfully (exit code: {exit_code})[/green]")
            return output or "Command executed successfully (no output)"
        else:
            console.print(f"\n[red]✗ Command failed with exit code {exit_code}[/red]")
            return f"Command failed with exit code {exit_code}. Output:\n{output}"
            
    except Exception as e:
        error_msg = f"Failed to execute command: {e}"
        console.print(f"[red]✗ {error_msg}[/red]")
        return error_msg


def get_available_tools() -> List[Dict[str, Any]]:
    """Get list of available tools in unified format."""
    tools = []
    
    # Add the shell command tool
    tools.append(execute_shell_command._tool_schema)
    
    return tools


def get_available_tools_openai() -> List[Dict[str, Any]]:
    """Get list of available tools in OpenAI format."""
    tools = []
    
    # Add the shell command tool
    tools.append(execute_shell_command._tool_schema_openai)
    
    return tools


def execute_tool_call(tool_name: str, arguments: Dict[str, Any]) -> str:
    """Execute a tool call by name and arguments."""
    if tool_name == "execute_shell_command":
        return execute_shell_command(**arguments)
    else:
        return f"Unknown tool: {tool_name}"