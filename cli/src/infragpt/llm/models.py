#!/usr/bin/env python3
"""
Type definitions and model-related constants for LLM interactions.
"""

from typing import Literal, Dict, Any

# Model type definitions
MODEL_TYPE = Literal["gpt4o", "claude"]

# Provider-specific model identifiers
OPENAI_MODEL = Literal["o4-mini"]
ANTHROPIC_MODEL = Literal["claude-sonnet-4-20250514"]

# Mapping from MODEL_TYPE to actual model identifier
MODEL_MAP: Dict[MODEL_TYPE, str] = {
    "gpt4o": "o4-mini",
    "claude": "claude-sonnet-4-20250514"
}

# Default parameters for each model
DEFAULT_PARAMS: Dict[MODEL_TYPE, Dict[str, Any]] = {
    "gpt4o": {
        "temperature": 0.0,
        "max_tokens": None,  # Use model default
        "top_p": 1.0,
    },
    "claude": {
        "temperature": 0.0,
        "max_tokens": 4096,  # Claude requires max_tokens to be specified
        "top_p": 1.0,
    }
}