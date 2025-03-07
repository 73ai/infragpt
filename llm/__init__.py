#!/usr/bin/env python3
"""
LLM module for InfraGPT - provides shared LLM functionality for CLI and server.

This module centralizes functionality for interacting with language models,
providing a clean interface for model operations, API key validation,
and prompt handling.
"""

__version__ = "0.1.0"

# Public exports from client.py
from llm.client import (
    get_llm_client,
    generate_gcloud_command,
    get_parameter_info,
)

# Public exports from auth.py
from llm.auth import validate_api_key

# Public exports from prompts.py
from llm.prompts import get_prompt_template, format_prompt

# Public exports from models.py
from llm.models import MODEL_TYPE

# Public exports from errors.py
from llm.errors import (
    LLMError,
    AuthenticationError,
    GenerationError,
    ParsingError,
    ValidationError,
    ConfigurationError,
)