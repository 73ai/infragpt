"""
Custom exceptions for InfraGPT client
"""


class InfraGPTError(Exception):
    """Base exception for InfraGPT client operations."""
    pass


class ConnectionError(InfraGPTError):
    """Raised when unable to connect to InfraGPT service."""
    pass


class RequestError(InfraGPTError):
    """Raised when a request to InfraGPT service fails."""
    pass