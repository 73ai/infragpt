"""
Backend API Client

Simple Python client for communicating with Backend service.
"""

from .client import BackendClient
from .exceptions import BackendError

__version__ = "1.0.0"
__all__ = ["BackendClient", "BackendError"]