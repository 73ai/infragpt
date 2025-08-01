#!/usr/bin/env python3
"""
Core LLM client functionality for making requests to language models.
"""

import datetime
import json
from typing import Optional, Dict, Any, Union

from langchain_openai import ChatOpenAI
from langchain_anthropic import ChatAnthropic
from langchain_core.language_models import BaseChatModel
from langchain_core.output_parsers import StrOutputParser
from langchain_core.prompts import ChatPromptTemplate

from infragpt.llm.models import MODEL_TYPE, MODEL_MAP, DEFAULT_PARAMS
from infragpt.llm.errors import AuthenticationError, GenerationError, ValidationError
from infragpt.llm.auth import validate_api_key
from infragpt.llm.prompts import get_prompt_template


def get_llm_client(
    model_type: MODEL_TYPE,
    api_key: str,
    *,
    temperature: Optional[float] = None,
    max_tokens: Optional[int] = None,
    validate_key: bool = False,
) -> BaseChatModel:
    """
    Creates and returns an appropriate LLM client based on model type.

    Args:
        model_type: The type of model to use ("gpt4o" or "claude")
        api_key: API key for the specified model
        temperature: Controls randomness in the response (0.0-1.0)
        max_tokens: Maximum tokens in the response (None for model default)
        validate_key: Whether to validate the API key before creating the client

    Returns:
        A LangChain chat model instance configured with the provided settings

    Raises:
        ValueError: If model_type is not supported
        AuthenticationError: If API key validation is requested and the key is invalid
    """
    # Validate model type
    if model_type not in MODEL_MAP:
        raise ValueError(f"Unsupported model type: {model_type}")

    # Validate API key if requested
    if validate_key and not validate_api_key(model_type, api_key):
        raise AuthenticationError(f"Invalid API key for {model_type}")

    # Get default parameters for the model
    params = DEFAULT_PARAMS[model_type].copy()

    # Override with provided parameters if specified
    if temperature is not None:
        params["temperature"] = temperature
    if max_tokens is not None:
        params["max_tokens"] = max_tokens

    # Create the appropriate client based on model type
    if model_type == "gpt4o":
        client_params = {
            "model": MODEL_MAP[model_type],
            "api_key": api_key,
            "temperature": params["temperature"]
        }
        if params["max_tokens"] is not None:
            client_params["max_tokens"] = params["max_tokens"]
        return ChatOpenAI(**client_params)
    elif model_type == "claude":
        client_params = {
            "model": MODEL_MAP[model_type],
            "api_key": api_key,
            "temperature": params["temperature"]
        }
        if params["max_tokens"] is not None:
            client_params["max_tokens"] = params["max_tokens"]
        return ChatAnthropic(**client_params)
    else:
        # This should never happen due to the validation above
        raise ValueError(f"Unsupported model type: {model_type}")