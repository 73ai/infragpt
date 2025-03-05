"""LLM provider system for InfraGPT."""

import os
import sys
import time
import json
import hashlib
import pathlib
from abc import ABC, abstractmethod
from typing import Optional, Dict, Any, Literal, Tuple, List, Union

from langchain_openai import ChatOpenAI
from langchain_anthropic import ChatAnthropic
from langchain_core.messages import AIMessage
from langchain_core.output_parsers import StrOutputParser

from ..ui.console import console
from ..utils.config import Config
from ..utils.logging import logger, APIError, AuthenticationError, handle_error

# Define types for model selection
MODEL_TYPE = Literal["gpt4o", "claude"]

# Cache directory
CACHE_DIR = pathlib.Path.home() / ".config" / "infragpt" / "cache"


class LLMResponseCache:
    """Cache for LLM responses to avoid redundant API calls."""

    def __init__(self, cache_dir: pathlib.Path = CACHE_DIR, ttl: int = 86400):
        """Initialize the cache with a directory and time-to-live for entries.
        
        Args:
            cache_dir: Directory to store cache files
            ttl: Time-to-live in seconds (default: 24 hours)
        """
        self.cache_dir = cache_dir
        self.ttl = ttl
        self.cache_dir.mkdir(parents=True, exist_ok=True)

    def _get_cache_key(self, model: str, prompt: str, system_prompt: Optional[str] = None) -> str:
        """Generate a unique cache key based on model and prompts."""
        # Combine all inputs that affect the response
        key_data = f"{model}:{prompt}"
        if system_prompt:
            key_data += f":{system_prompt}"
        
        # Create a hash of the combined data
        return hashlib.md5(key_data.encode()).hexdigest()

    def _get_cache_path(self, cache_key: str) -> pathlib.Path:
        """Get the file path for a cache key."""
        return self.cache_dir / f"{cache_key}.json"

    def get(self, model: str, prompt: str, system_prompt: Optional[str] = None) -> Optional[str]:
        """Retrieve a cached response if available and not expired."""
        cache_key = self._get_cache_key(model, prompt, system_prompt)
        cache_path = self._get_cache_path(cache_key)
        
        if not cache_path.exists():
            logger.debug(f"Cache miss: no entry for key {cache_key}")
            return None
        
        try:
            with open(cache_path, "r") as f:
                cache_data = json.load(f)
            
            # Check if cache entry has expired
            if time.time() - cache_data["timestamp"] > self.ttl:
                logger.debug(f"Cache entry expired for key {cache_key}")
                # Cache expired, remove it
                cache_path.unlink(missing_ok=True)
                return None
            
            logger.debug(f"Cache hit for {model}")
            return cache_data["response"]
        except Exception as e:
            logger.warning(f"Cache read error for key {cache_key}: {str(e)}", exc_info=True)
            console.print(f"[dim]Cache read error: {e}[/dim]")
            return None

    def set(self, model: str, prompt: str, response: str, system_prompt: Optional[str] = None) -> None:
        """Store a response in the cache."""
        cache_key = self._get_cache_key(model, prompt, system_prompt)
        cache_path = self._get_cache_path(cache_key)
        
        try:
            cache_data = {
                "model": model,
                "prompt": prompt,
                "system_prompt": system_prompt,
                "response": response,
                "timestamp": time.time()
            }
            
            with open(cache_path, "w") as f:
                json.dump(cache_data, f)
            logger.debug(f"Cached response for {model}")
        except Exception as e:
            logger.warning(f"Cache write error for key {cache_key}: {str(e)}", exc_info=True)
            console.print(f"[dim]Cache write error: {e}[/dim]")

    def clear(self, max_age: Optional[int] = None) -> int:
        """Clear cache entries.
        
        Args:
            max_age: If provided, only clear entries older than max_age seconds
            
        Returns:
            Number of entries cleared
        """
        count = 0
        
        for cache_file in self.cache_dir.glob("*.json"):
            try:
                if max_age is not None:
                    # Check file age
                    with open(cache_file, "r") as f:
                        cache_data = json.load(f)
                    
                    if time.time() - cache_data["timestamp"] <= max_age:
                        continue
                
                cache_file.unlink()
                count += 1
            except Exception:
                # Skip files that can't be processed
                pass
        
        return count




class BaseLLMProvider(ABC):
    """Base class for LLM providers."""
    
    def __init__(self, api_key: str, verbose: bool = False, use_cache: bool = True):
        """Initialize with API key and options."""
        self.api_key = api_key
        self.verbose = verbose
        self.use_cache = use_cache
        self.cache = LLMResponseCache() if use_cache else None
    
    @abstractmethod
    def get_name(self) -> str:
        """Get the provider name."""
        pass
    
    @abstractmethod
    def get_model_name(self) -> str:
        """Get the model name being used."""
        pass
    
    @abstractmethod
    def validate_key(self) -> bool:
        """Validate the API key."""
        pass
    
    @abstractmethod
    def generate(self, prompt: str, system_prompt: Optional[str] = None) -> str:
        """Generate a response for the given prompt."""
        pass


class OpenAIProvider(BaseLLMProvider):
    """OpenAI provider using the GPT-4o model."""
    
    def __init__(self, api_key: str, verbose: bool = False, use_cache: bool = True):
        """Initialize OpenAI provider with API key."""
        super().__init__(api_key, verbose, use_cache)
        self.model_name = "gpt-4o"
    
    def get_name(self) -> str:
        """Get the provider name."""
        return "openai"
    
    def get_model_name(self) -> str:
        """Get the model name being used."""
        return self.model_name
    
    def validate_key(self) -> bool:
        """Validate the OpenAI API key."""
        try:
            # Create a minimal client for validation
            llm = ChatOpenAI(
                model=self.model_name,
                temperature=0,
                api_key=self.api_key,
                max_tokens=5
            )
            # Make a minimal request
            response = llm.invoke("Say OK")
            logger.debug("OpenAI API key validated successfully")
            return True
        except Exception as e:
            error_str = str(e).lower()
            is_auth_error = any(term in error_str for term in ["api key", "auth", "key", "token", "authentication"])
            
            if is_auth_error:
                error_msg = f"Invalid OpenAI API key: {e}"
                logger.error(error_msg)
                console.print(f"[bold red]Invalid OpenAI API key:[/bold red] {e}")
                
                # Create a structured error but don't raise it - just log it
                auth_error = AuthenticationError(
                    message=error_msg,
                    provider="openai",
                    details={"error": str(e)}
                )
                logger.error(f"Authentication error: {auth_error.error_code}", 
                             extra={"error_code": auth_error.error_code, "details": auth_error.details})
                return False
            else:
                # If the error is not related to authentication, log and still return True
                warning_msg = f"OpenAI API connection error: {e}"
                logger.warning(warning_msg, exc_info=True)
                console.print(f"[bold yellow]Warning:[/bold yellow] {warning_msg}")
                return True
    
    def generate(self, prompt: str, system_prompt: Optional[str] = None) -> str:
        """Generate a response for the given prompt."""
        # Check cache first
        if self.use_cache and self.cache:
            cached_response = self.cache.get(self.model_name, prompt, system_prompt)
            if cached_response:
                if self.verbose:
                    console.print("[dim]Using cached response[/dim]")
                return cached_response
        
        # Set up messages
        messages = []
        if system_prompt:
            messages.append({"role": "system", "content": system_prompt})
        messages.append({"role": "user", "content": prompt})
        
        try:
            # Log the request (without the actual prompt contents for privacy)
            logger.debug(f"Generating response with OpenAI model {self.model_name}")
            
            # Create LLM instance
            llm = ChatOpenAI(
                model=self.model_name,
                temperature=0,
                api_key=self.api_key
            )
            
            # Generate response
            response = llm.invoke(prompt)
            
            # Extract the text content
            if isinstance(response, AIMessage):
                result = response.content
            elif isinstance(response, str):
                result = response
            else:
                result = str(response)
            
            # Cache the result
            if self.use_cache and self.cache:
                self.cache.set(self.model_name, prompt, result, system_prompt)
            
            logger.debug("Successfully generated response from OpenAI")
            return result
        except Exception as e:
            error_str = str(e).lower()
            is_auth_error = any(term in error_str for term in ["api key", "auth", "key", "token", "unauthorized"])
            
            if is_auth_error:
                error_msg = f"OpenAI authentication error: {e}"
                logger.error(error_msg, exc_info=True)
                
                # Log structured error
                auth_error = AuthenticationError(
                    message=error_msg,
                    provider="openai",
                    details={"error": str(e)}
                )
                # We don't raise the error here to maintain backward compatibility
                handle_error(auth_error, exit_app=False, show_traceback=False)
            else:
                error_msg = f"OpenAI API error: {e}"
                logger.error(error_msg, exc_info=True)
                
                # Log structured error
                api_error = APIError(
                    message=error_msg,
                    provider="openai",
                    details={"error": str(e)}
                )
                # We don't raise the error here to maintain backward compatibility
                handle_error(api_error, exit_app=False, show_traceback=False)
            
            console.print(f"[bold red]Error generating response:[/bold red] {e}")
            return f"Error: {str(e)}"


class AnthropicProvider(BaseLLMProvider):
    """Anthropic provider using the Claude model."""
    
    def __init__(self, api_key: str, verbose: bool = False, use_cache: bool = True):
        """Initialize Anthropic provider with API key."""
        super().__init__(api_key, verbose, use_cache)
        self.model_name = "claude-3-sonnet-20240229"
    
    def get_name(self) -> str:
        """Get the provider name."""
        return "anthropic"
    
    def get_model_name(self) -> str:
        """Get the model name being used."""
        return self.model_name
    
    def validate_key(self) -> bool:
        """Validate the Anthropic API key."""
        try:
            # Create a minimal client for validation
            llm = ChatAnthropic(
                model=self.model_name,
                temperature=0,
                api_key=self.api_key,
                max_tokens=5
            )
            # Make a minimal request
            response = llm.invoke("Say OK")
            return True
        except Exception as e:
            if "API key" in str(e) or "auth" in str(e).lower() or "key" in str(e).lower() or "token" in str(e).lower():
                console.print(f"[bold red]Invalid Anthropic API key:[/bold red] {e}")
                return False
            else:
                # If the error is not related to authentication, log and still return True
                console.print(f"[bold yellow]Warning:[/bold yellow] API connection error: {e}")
                return True
    
    def generate(self, prompt: str, system_prompt: Optional[str] = None) -> str:
        """Generate a response for the given prompt."""
        # Check cache first
        if self.use_cache and self.cache:
            cached_response = self.cache.get(self.model_name, prompt, system_prompt)
            if cached_response:
                if self.verbose:
                    console.print("[dim]Using cached response[/dim]")
                return cached_response
        
        try:
            # Create LLM instance
            llm = ChatAnthropic(
                model=self.model_name,
                temperature=0,
                api_key=self.api_key
            )
            
            # Generate response
            response = llm.invoke(prompt)
            
            # Extract the text content
            if isinstance(response, AIMessage):
                result = response.content
            elif isinstance(response, str):
                result = response
            else:
                result = str(response)
            
            # Cache the result
            if self.use_cache and self.cache:
                self.cache.set(self.model_name, prompt, result, system_prompt)
            
            return result
        except Exception as e:
            console.print(f"[bold red]Error generating response:[/bold red] {e}")
            return f"Error: {str(e)}"


def create_provider(model_type: MODEL_TYPE, api_key: str, verbose: bool = False, use_cache: bool = True) -> BaseLLMProvider:
    """Create the appropriate LLM provider based on the model type."""
    if model_type == "gpt4o":
        return OpenAIProvider(api_key, verbose, use_cache)
    elif model_type == "claude":
        return AnthropicProvider(api_key, verbose, use_cache)
    else:
        raise ValueError(f"Unsupported model type: {model_type}")


def validate_api_key(model_type: MODEL_TYPE, api_key: str) -> bool:
    """Validate if the API key is correct by making a minimal API call."""
    provider = create_provider(model_type, api_key, verbose=False)
    return provider.validate_key()


def get_provider(model_type: Optional[MODEL_TYPE] = None, api_key: Optional[str] = None, 
               verbose: bool = False, validate: bool = True, use_cache: bool = True) -> BaseLLMProvider:
    """Get the appropriate LLM provider with validated credentials."""
    # Import here to avoid circular imports
    from ..ui.prompts import prompt_credentials
    
    # Get config and credentials
    config = Config()
    
    # Determine the model and API key
    resolved_model, resolved_api_key = get_credentials(model_type, api_key, verbose)
    
    # Validate API key if requested
    if validate:
        # If key is invalid, prompt for a new one
        while not validate_api_key(resolved_model, resolved_api_key):
            console.print("[bold red]API key validation failed.[/bold red]")
            resolved_model, resolved_api_key = prompt_credentials(resolved_model)
    
    # Create and return the provider
    return create_provider(resolved_model, resolved_api_key, verbose, use_cache)


def get_credentials(model_type: Optional[MODEL_TYPE] = None, api_key: Optional[str] = None, verbose: bool = False) -> Tuple[MODEL_TYPE, str]:
    """
    Get API credentials based on priority:
    1. Command line parameters
    2. Stored config
    3. Environment variables
    4. Interactive prompt
    """
    # Import here to avoid circular imports
    from ..ui.prompts import prompt_credentials
    
    config = Config()
    
    # Priority 1: Command line parameters
    if model_type and api_key and api_key.strip():  # Ensure API key is not empty
        # Update config for future use
        config.update_credentials(model_type, api_key)
        return model_type, api_key
    
    # Priority 2: Check stored config
    credentials = config.get_credentials()
    stored_model = credentials.get('model')
    stored_api_key = credentials.get('api_key')
    if stored_model and stored_api_key and stored_api_key.strip():
        if verbose:
            console.print(f"[dim]Using credentials from config file[/dim]")
        return stored_model, stored_api_key
    
    # Priority 3: Check environment variables
    openai_key = os.getenv("OPENAI_API_KEY")
    anthropic_key = os.getenv("ANTHROPIC_API_KEY")
    env_model = os.getenv("INFRAGPT_MODEL")
    
    # Command line model takes precedence over env var model
    resolved_model = model_type or env_model
    
    # Validate environment credentials
    if anthropic_key and openai_key:
        # If both keys are provided, use the model to decide
        if resolved_model == "claude":
            if verbose:
                console.print(f"[dim]Using Anthropic API key from environment[/dim]")
            # Save to config for future use
            config.update_credentials("claude", anthropic_key)
            return "claude", anthropic_key
        elif resolved_model == "gpt4o":
            if verbose:
                console.print(f"[dim]Using OpenAI API key from environment[/dim]")
            # Save to config for future use
            config.update_credentials("gpt4o", openai_key)
            return "gpt4o", openai_key
        elif not resolved_model:
            # Default to OpenAI if model not specified
            if verbose:
                console.print(f"[dim]Multiple API keys found, defaulting to OpenAI[/dim]")
            # Save to config for future use
            config.update_credentials("gpt4o", openai_key)
            return "gpt4o", openai_key
    elif anthropic_key:
        if resolved_model and resolved_model != "claude":
            console.print("[bold red]Error:[/bold red] Anthropic API key is set but model is not claude.")
            sys.exit(1)
        if verbose:
            console.print(f"[dim]Using Anthropic API key from environment[/dim]")
        # Save to config for future use
        config.update_credentials("claude", anthropic_key)
        return "claude", anthropic_key
    elif openai_key:
        if resolved_model and resolved_model != "gpt4o":
            console.print("[bold red]Error:[/bold red] OpenAI API key is set but model is not gpt4o.")
            sys.exit(1)
        if verbose:
            console.print(f"[dim]Using OpenAI API key from environment[/dim]")
        # Save to config for future use
        config.update_credentials("gpt4o", openai_key)
        return "gpt4o", openai_key
    
    # Priority 4: Prompt user interactively
    console.print("\n[bold yellow]API credentials required[/bold yellow]")
    
    # Get credentials through prompting
    return prompt_credentials(model_type)