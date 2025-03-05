"""Logging and error handling for InfraGPT."""

import os
import sys
import logging
import pathlib
import traceback
import datetime
from typing import Dict, Any, Optional, List, Type, Union, Callable

from rich.logging import RichHandler
from rich.traceback import Traceback

from ..ui.console import console

# Define log directory
LOG_DIR = pathlib.Path.home() / ".config" / "infragpt" / "logs"
LOG_FILE = LOG_DIR / "infragpt.log"

# Define custom exception classes
class InfraGPTError(Exception):
    """Base exception class for InfraGPT errors."""
    
    def __init__(self, message: str, error_code: Optional[str] = None, details: Optional[Dict[str, Any]] = None):
        """Initialize with error message, code, and details.
        
        Args:
            message: Human-readable error description
            error_code: Optional error code for programmatic handling
            details: Optional dictionary with additional error context
        """
        self.message = message
        self.error_code = error_code or "UNKNOWN_ERROR"
        self.details = details or {}
        super().__init__(message)


class ConfigError(InfraGPTError):
    """Error related to configuration loading/saving."""
    
    def __init__(self, message: str, details: Optional[Dict[str, Any]] = None):
        super().__init__(message, "CONFIG_ERROR", details)


class APIError(InfraGPTError):
    """Error related to API calls (OpenAI, Anthropic, etc.)."""
    
    def __init__(self, message: str, provider: Optional[str] = None, details: Optional[Dict[str, Any]] = None):
        error_details = details or {}
        if provider:
            error_details["provider"] = provider
        super().__init__(message, "API_ERROR", error_details)


class AuthenticationError(InfraGPTError):
    """Error related to authentication (API keys, etc.)."""
    
    def __init__(self, message: str, provider: Optional[str] = None, details: Optional[Dict[str, Any]] = None):
        error_details = details or {}
        if provider:
            error_details["provider"] = provider
        super().__init__(message, "AUTH_ERROR", error_details)


class CommandError(InfraGPTError):
    """Error related to command generation or execution."""
    
    def __init__(self, message: str, command: Optional[str] = None, details: Optional[Dict[str, Any]] = None):
        error_details = details or {}
        if command:
            error_details["command"] = command
        super().__init__(message, "COMMAND_ERROR", error_details)


# Set up logging
def setup_logging(
    level: int = logging.INFO, 
    log_file: bool = True,
    console_output: bool = True,
    log_file_path: Optional[pathlib.Path] = None
) -> logging.Logger:
    """Set up logging configuration.
    
    Args:
        level: Logging level (e.g., logging.DEBUG, logging.INFO)
        log_file: Whether to log to a file
        console_output: Whether to log to the console
        log_file_path: Custom log file path
        
    Returns:
        Configured logger instance
    """
    # Create logger
    logger = logging.getLogger("infragpt")
    logger.setLevel(level)
    
    # Remove existing handlers to avoid duplicates
    for handler in logger.handlers[:]:
        logger.removeHandler(handler)
    
    # Create formatters
    verbose_formatter = logging.Formatter(
        '%(asctime)s [%(name)s] [%(levelname)s] %(message)s'
    )
    
    # Add console handler if requested
    if console_output:
        console_handler = RichHandler(
            rich_tracebacks=True,
            console=console,
            tracebacks_show_locals=True,
            show_time=False,
            omit_repeated_times=True
        )
        console_handler.setLevel(level)
        logger.addHandler(console_handler)
    
    # Add file handler if requested
    if log_file:
        # Use provided path or default
        log_path = log_file_path or LOG_FILE
        
        # Ensure log directory exists
        log_path.parent.mkdir(parents=True, exist_ok=True)
        
        # Create file handler
        file_handler = logging.FileHandler(log_path)
        file_handler.setLevel(level)
        file_handler.setFormatter(verbose_formatter)
        logger.addHandler(file_handler)
    
    return logger


# Global logger instance
logger = setup_logging(
    level=logging.INFO,
    log_file=True,
    console_output=False  # We'll handle console output separately
)


# Error handling functions
def handle_error(
    error: Exception,
    log_level: int = logging.ERROR,
    exit_app: bool = False,
    show_traceback: bool = False,
    console_msg: Optional[str] = None
) -> None:
    """Handle an exception with appropriate logging and user feedback.
    
    Args:
        error: The exception to handle
        log_level: Logging level for this error
        exit_app: Whether to exit the application
        show_traceback: Whether to show traceback to the user
        console_msg: Optional custom message for console output
    """
    # Create error details
    error_type = type(error).__name__
    error_msg = str(error)
    
    # Add InfraGPT-specific details if available
    if isinstance(error, InfraGPTError):
        error_code = error.error_code
        error_details = error.details
    else:
        error_code = "UNKNOWN_ERROR"
        error_details = {}
    
    # Log the error
    if log_level >= logging.ERROR:
        logger.error(
            f"Error: {error_type} - {error_msg}",
            exc_info=True,
            extra={"error_code": error_code, "details": error_details}
        )
    elif log_level >= logging.WARNING:
        logger.warning(
            f"Warning: {error_type} - {error_msg}",
            extra={"error_code": error_code, "details": error_details}
        )
    elif log_level >= logging.INFO:
        logger.info(
            f"Info: {error_type} - {error_msg}",
            extra={"error_code": error_code, "details": error_details}
        )
    
    # Show console message if provided
    if console_msg:
        console.print(f"[bold red]Error:[/bold red] {console_msg}")
    else:
        console.print(f"[bold red]Error:[/bold red] {error_msg}")
    
    # Show traceback if requested
    if show_traceback:
        console.print(Traceback.from_exception(
            type(error),
            error,
            traceback.extract_tb(error.__traceback__)
        ))
    
    # Exit if requested
    if exit_app:
        sys.exit(1)


def error_boundary(
    func: Callable,
    error_types: Union[Type[Exception], List[Type[Exception]]] = Exception,
    log_level: int = logging.ERROR,
    exit_app: bool = False,
    show_traceback: bool = False,
    console_msg: Optional[str] = None
) -> Callable:
    """Decorator that creates an error boundary around a function.
    
    Args:
        func: The function to wrap
        error_types: Exception type(s) to catch
        log_level: Logging level for caught errors
        exit_app: Whether to exit the application on error
        show_traceback: Whether to show traceback to the user
        console_msg: Optional custom message for console output
        
    Returns:
        Wrapped function with error handling
    """
    def wrapper(*args, **kwargs):
        try:
            return func(*args, **kwargs)
        except error_types as e:
            handle_error(e, log_level, exit_app, show_traceback, console_msg)
            return None
    
    return wrapper


# Aliases for common log levels
def debug(msg: str, *args, **kwargs) -> None:
    """Log a debug message."""
    logger.debug(msg, *args, **kwargs)

def info(msg: str, *args, **kwargs) -> None:
    """Log an info message."""
    logger.info(msg, *args, **kwargs)

def warning(msg: str, *args, **kwargs) -> None:
    """Log a warning message."""
    logger.warning(msg, *args, **kwargs)

def error(msg: str, *args, **kwargs) -> None:
    """Log an error message."""
    logger.error(msg, *args, **kwargs)

def critical(msg: str, *args, **kwargs) -> None:
    """Log a critical message."""
    logger.critical(msg, *args, **kwargs)


# Function to get a copy of the log file content
def get_logs(
    max_lines: int = 100,
    level: int = logging.INFO,
    log_file_path: Optional[pathlib.Path] = None
) -> List[str]:
    """Get recent log entries from the log file.
    
    Args:
        max_lines: Maximum number of lines to return
        level: Minimum log level to include
        log_file_path: Custom log file path
        
    Returns:
        List of log lines
    """
    log_path = log_file_path or LOG_FILE
    
    if not log_path.exists():
        return []
    
    try:
        with open(log_path, "r") as f:
            lines = f.readlines()
        
        # Filter by log level if needed
        if level > logging.DEBUG:
            filtered_lines = []
            for line in lines:
                # Simple log level detection
                if "[ERROR]" in line and level <= logging.ERROR:
                    filtered_lines.append(line)
                elif "[WARNING]" in line and level <= logging.WARNING:
                    filtered_lines.append(line)
                elif "[INFO]" in line and level <= logging.INFO:
                    filtered_lines.append(line)
                elif "[DEBUG]" in line and level <= logging.DEBUG:
                    filtered_lines.append(line)
            lines = filtered_lines
        
        # Return last N lines
        return lines[-max_lines:]
    except Exception as e:
        logger.warning(f"Could not read log file: {e}")
        return []