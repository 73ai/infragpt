#!/usr/bin/env python3
"""
Prompt handling and command processing for InfraGPT CLI.
"""

# Import from local LLM module
from infragpt.llm.prompts import get_prompt_template


def get_agent_system_prompt() -> str:
    """Get the system prompt for the shell agent."""
    return get_prompt_template('shell_agent')








