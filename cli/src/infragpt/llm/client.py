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
        return ChatOpenAI(
            model=MODEL_MAP[model_type],
            api_key=api_key,
            temperature=params["temperature"],
            max_tokens=params["max_tokens"]
        )
    elif model_type == "claude":
        return ChatAnthropic(
            model=MODEL_MAP[model_type],
            api_key=api_key,
            temperature=params["temperature"],
            max_tokens=params["max_tokens"]
        )
    else:
        # This should never happen due to the validation above
        raise ValueError(f"Unsupported model type: {model_type}")


def generate_gcloud_command(
    prompt: str,
    model_type: MODEL_TYPE,
    api_key: str,
    *,
    temperature: Optional[float] = None,
    max_tokens: Optional[int] = None,
) -> str:
    """
    Generate a gcloud command based on the user's natural language prompt.

    Args:
        prompt: Natural language description of the desired command
        model_type: The type of model to use
        api_key: API key for the specified model
        temperature: Controls randomness in the response (0.0-1.0)
        max_tokens: Maximum tokens in the response (None for model default)

    Returns:
        Generated gcloud command string

    Raises:
        ValueError: If model_type is not supported
        AuthenticationError: If the API key is invalid
        GenerationError: If command generation fails
    """
    # Get command generation prompt template
    prompt_template = get_prompt_template("command_generation")

    try:
        # Initialize the LLM
        llm = get_llm_client(
            model_type=model_type,
            api_key=api_key,
            temperature=temperature,
            max_tokens=max_tokens,
            validate_key=True  # Always validate the key for command generation
        )

        # Create and execute the chain
        chain = prompt_template | llm | StrOutputParser()
        start_time = datetime.datetime.now()
        result = chain.invoke({"prompt": prompt})
        end_time = datetime.datetime.now()

        return result.strip()

    except AuthenticationError as e:
        # Re-raise authentication errors
        raise
    except Exception as e:
        # Wrap other errors in GenerationError
        raise GenerationError(f"Failed to generate command: {str(e)}") from e


def get_parameter_info(
    command: str,
    model_type: MODEL_TYPE,
    api_key: str,
) -> Dict[str, Dict[str, Any]]:
    """
    Get information about parameters from the LLM.

    Args:
        command: The command to analyze
        model_type: The type of model to use
        api_key: API key for the specified model

    Returns:
        Dictionary of parameter information with structure:
        {
            "PARAM_NAME": {
                "description": "Parameter description",
                "examples": ["example1", "example2"],
                "required": true,
                "default": "default value or null"
            }
        }

    Raises:
        ValueError: If model_type is not supported
        AuthenticationError: If the API key is invalid
        ParsingError: If response parsing fails
    """
    # Get parameter info prompt template
    prompt_template = get_prompt_template("parameter_info")

    try:
        # Initialize the LLM
        llm = get_llm_client(
            model_type=model_type,
            api_key=api_key,
            validate_key=True
        )

        # Create and execute the chain
        chain = prompt_template | llm | StrOutputParser()
        result = chain.invoke({"command": command})

        # Find JSON part between triple backticks if present
        if "```json" in result:
            json_part = result.split("```json")[1].split("```")[0].strip()
        elif "```" in result:
            json_part = result.split("```")[1].strip()
        else:
            json_part = result.strip()

        parameter_info = json.loads(json_part)
        return parameter_info

    except AuthenticationError as e:
        # Re-raise authentication errors
        raise
    except json.JSONDecodeError as e:
        from infragpt.llm.errors import ParsingError
        raise ParsingError(f"Failed to parse parameter info: {str(e)}") from e
    except Exception as e:
        # Wrap other errors
        from infragpt.llm.errors import ParsingError
        raise ParsingError(f"Failed to get parameter info: {str(e)}") from e
