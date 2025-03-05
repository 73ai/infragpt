"""Console output utilities for InfraGPT.

This module provides a singleton console instance to be used throughout the application.
This prevents errors with multiple live displays being active at once.
"""

from rich.console import Console

# Initialize console for rich output as a singleton instance
console = Console()

def get_console():
    """Get the singleton console instance.
    
    Returns:
        Console: The global console instance.
    """
    global console
    return console