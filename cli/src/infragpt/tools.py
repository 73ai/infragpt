"""
Tool system without LangChain dependencies.
"""

import inspect
from typing import Callable, Dict, Any, Optional, List
from functools import wraps
from rich.console import Console
from .shell import CommandExecutor
from .container import ExecutorInterface, is_sandbox_mode, get_executor as get_container_executor
from .llm.models import Tool, InputSchema, Parameter


class ToolExecutionCancelled(Exception):
    """Raised when user cancels tool execution to break the agentic loop."""

    pass


console = Console()

_host_executor: Optional[CommandExecutor] = None


def get_executor() -> ExecutorInterface:
    """Get the appropriate executor based on sandbox mode."""
    global _host_executor

    if is_sandbox_mode():
        return get_container_executor()
    else:
        if _host_executor is None:
            _host_executor = CommandExecutor(timeout=60)
        return _host_executor


def cleanup_executor() -> None:
    """Clean up executor resources."""
    global _host_executor

    if is_sandbox_mode():
        from .container import cleanup_executor as cleanup_container
        cleanup_container()
    elif _host_executor is not None:
        _host_executor.cleanup()
        _host_executor = None


def tool(name: Optional[str] = None, description: Optional[str] = None):
    """
    Decorator to convert a function into a tool definition.
    Replacement for LangChain's @tool decorator.
    """

    def decorator(func: Callable) -> Callable:
        sig = inspect.signature(func)
        tool_name = name or func.__name__
        tool_description = description or func.__doc__ or f"Execute {tool_name}"

        properties = {}
        required = []

        for param_name, param in sig.parameters.items():
            param_type = "string"
            param_description = f"Parameter {param_name}"
            param_default = (
                None if param.default == inspect.Parameter.empty else param.default
            )

            if param.annotation != inspect.Parameter.empty:
                if param.annotation is str:
                    param_type = "string"
                elif param.annotation is int:
                    param_type = "integer"
                elif param.annotation is float:
                    param_type = "number"
                elif param.annotation is bool:
                    param_type = "boolean"

                if (
                    hasattr(param.annotation, "__args__")
                    and type(None) in param.annotation.__args__
                ):
                    actual_type = next(
                        t for t in param.annotation.__args__ if t is not type(None)
                    )
                    if actual_type is str:
                        param_type = "string"
                    elif actual_type is int:
                        param_type = "integer"
                    elif actual_type is float:
                        param_type = "number"
                    elif actual_type is bool:
                        param_type = "boolean"
                    param_description = f"Optional parameter {param_name}"

            properties[param_name] = Parameter(
                type=param_type, description=param_description, default=param_default
            )

            if param.default == inspect.Parameter.empty:
                required.append(param_name)

        input_schema = InputSchema(
            type="object",
            properties=properties,
            required=required,
            additionalProperties=False,
        )

        tool_obj = Tool(
            name=tool_name, description=tool_description, input_schema=input_schema
        )

        func._tool = tool_obj

        @wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return decorator


@tool(
    name="execute_shell_command",
    description="Execute a shell command and return the output",
)
def execute_shell_command(command: str, description: Optional[str] = None) -> str:
    """
    Execute a shell command and return the output.

    Args:
        command: The shell command to execute
        description: Optional description of what the command does

    Returns:
        The command output or error message
    """
    console.print("\n[bold cyan]Tool Call: execute_shell_command[/bold cyan]")
    console.print(f"[dim]Command:[/dim] {command}")
    if description:
        console.print(f"[dim]Description:[/dim] {description}")

    console.print("\n[yellow]Execute this command? (Y/n):[/yellow] ", end="")
    console.file.flush()

    try:
        user_input = input().strip().lower()
    except (KeyboardInterrupt, EOFError):
        console.print("\n[yellow]Command execution cancelled.[/yellow]")
        raise ToolExecutionCancelled("User cancelled command execution")

    if user_input not in ["y", "yes", ""]:
        console.print("\n[yellow]Command execution cancelled.[/yellow]")
        raise ToolExecutionCancelled("User cancelled command execution")

    try:
        executor = get_executor()
        console.print("\n[bold blue]Executing command...[/bold blue]")
        console.file.flush()

        exit_code, output, was_cancelled = executor.execute_command(command)

        if was_cancelled:
            return "Command execution was cancelled by user."
        elif exit_code == 0:
            console.print(
                f"\n[green]✓ Command executed successfully (exit code: {exit_code})[/green]"
            )
            return output or "Command executed successfully (no output)"
        else:
            console.print(f"\n[red]✗ Command failed with exit code {exit_code}[/red]")
            return f"Command failed with exit code {exit_code}. Output:\n{output}"

    except Exception as e:
        error_msg = f"Failed to execute command: {e}"
        console.print(f"[red]✗ {error_msg}[/red]")
        return error_msg


def get_available_tools() -> List[Tool]:
    """Get list of available Tool objects."""
    return [execute_shell_command._tool]


def get_tool_by_name(tool_name: str) -> Optional[Tool]:
    """Get a tool definition by name."""
    tools = get_available_tools()
    for tool in tools:
        if tool.name == tool_name:
            return tool
    return None


def execute_tool_call(tool_name: str, arguments: Dict[str, Any]) -> str:
    """Execute a tool call by name and arguments."""
    if tool_name == "execute_shell_command":
        return execute_shell_command(**arguments)
    else:
        return f"Unknown tool: {tool_name}"
