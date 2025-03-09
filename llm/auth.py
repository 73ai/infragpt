#!/usr/bin/env python3
"""
Authentication utilities for LLM API interactions.
"""

from typing import Optional

from langchain_openai import ChatOpenAI
from langchain_anthropic import ChatAnthropic

from llm.models import MODEL_TYPE
from llm.errors import AuthenticationError


def validate_api_key(model_type: MODEL_TYPE, api_key: str) -> bool:
    """
    Validate if the API key is correct by making a minimal API call.
    
    Args:
        model_type: The type of model to use ("gpt4o" or "claude")
        api_key: API key for the specified model
        
    Returns:
        True if the API key is valid, False otherwise
        
    Note:
        This makes a minimal API call to validate the key
    """
    try:
        if model_type == "gpt4o":
            # Create a minimal OpenAI client to validate the key
            llm = ChatOpenAI(
                model="gpt-4o", 
                temperature=0, 
                api_key=api_key,
                max_tokens=5  # Minimal response to reduce token usage
            )
            # Make a minimal request
            response = llm.invoke("Say OK")
            return True
        elif model_type == "claude":
            # Create a minimal Anthropic client to validate the key
            llm = ChatAnthropic(
                model="claude-3-sonnet-20240229", 
                temperature=0, 
                api_key=api_key,
                max_tokens=5  # Minimal response to reduce token usage
            )
            # Make a minimal request
            response = llm.invoke("Say OK")
            return True
        else:
            raise ValueError(f"Unsupported model type: {model_type}")
    except Exception as e:
        if "API key" in str(e) or "auth" in str(e).lower() or "key" in str(e).lower() or "token" in str(e).lower():
            return False
        else:
            # If the error is not related to authentication, we still allow the key - it might be a temporary issue
            return True