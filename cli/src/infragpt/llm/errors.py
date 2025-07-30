#!/usr/bin/env python3
"""
Custom exception classes for LLM-related errors.
"""

class LLMError(Exception):
    """Base exception for LLM-related errors."""
    pass


class AuthenticationError(LLMError):
    """Raised when authentication fails with the LLM provider."""
    pass


class GenerationError(LLMError):
    """Raised when command generation fails."""
    pass


class ParsingError(LLMError):
    """Raised when parsing LLM response fails."""
    pass


class ValidationError(LLMError):
    """Raised when validation of input or output fails."""
    pass


class ConfigurationError(LLMError):
    """Raised when there's an issue with the LLM configuration."""
    pass