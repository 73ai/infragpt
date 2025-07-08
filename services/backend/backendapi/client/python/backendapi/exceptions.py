"""
Custom exceptions for Backend client
"""


class BackendError(Exception):
    """Base exception for Backend client operations."""
    pass


class ConnectionError(BackendError):
    """Raised when unable to connect to Backend service."""
    pass


class RequestError(BackendError):
    """Raised when a request to Backend service fails."""
    pass