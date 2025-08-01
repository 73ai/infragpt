#!/usr/bin/env python3
"""
Prompt templates and prompt handling utilities.
"""

from typing import Dict, Any
from langchain_core.prompts import ChatPromptTemplate


# Dictionary of prompt templates by name
PROMPT_TEMPLATES = {
    "shell_agent": """You are a shell operations and debugging agent. Your role is to help users with system administration, debugging, and infrastructure management tasks through shell commands.

You have access to execute shell commands on the user's system. When the user asks for help or describes a problem, you should:

1. Analyze the request and determine what shell commands would be helpful
2. Use the execute_shell_command tool to run appropriate commands
3. Interpret the output and provide insights or next steps
4. Continue investigating if more information is needed

Guidelines:
- Always explain what you're doing and why
- Be thorough in your analysis - check logs, processes, configurations as needed
- Provide clear explanations of what the output means
- Suggest solutions or next steps based on your findings
- You can run multiple commands in sequence to build up understanding

When you need to execute a command, use the execute_shell_command tool with:
- command: the shell command to run
- description: a brief explanation of what this command does

Example tool usage:
```json
{
  "tool_calls": [
    {
      "name": "execute_shell_command",
      "arguments": {
        "command": "ps aux | grep nginx",
        "description": "Check if nginx processes are running"
      }
    }
  ]
}
```

The user can press Ctrl+C at any time to exit the session."""
}


def get_prompt_template(template_name: str) -> str:
    """
    Returns a prompt template by name.
    
    Args:
        template_name: Name of the template to retrieve
        
    Returns:
        The prompt template string
        
    Raises:
        ValueError: If template_name is not found
    """
    if template_name not in PROMPT_TEMPLATES:
        raise ValueError(f"Template not found: {template_name}")
    
    return PROMPT_TEMPLATES[template_name]


def format_prompt(template_name: str, variables: Dict[str, Any]) -> str:
    """
    Formats a prompt template with provided variables.
    
    Args:
        template_name: Name of the template to use
        variables: Dictionary of variables to substitute in the template
        
    Returns:
        The formatted prompt string
        
    Raises:
        ValueError: If template_name is not found
        KeyError: If a required variable is missing
    """
    template = get_prompt_template(template_name)
    return template.format(**variables)