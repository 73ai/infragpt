"""
InfraGPT API Client

Simple Python client for communicating with InfraGPT service.
"""

from .client import InfraGPTClient
from .exceptions import InfraGPTError

__version__ = "1.0.0"
__all__ = ["InfraGPTClient", "InfraGPTError"]