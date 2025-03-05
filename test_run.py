#!/usr/bin/env python3
"""Simple test script for InfraGPT."""

import sys
from infragpt.core.llm import get_provider, get_credentials
from infragpt.core.prompt import create_prompt

def main():
    """Run a simple command test."""
    print("Testing InfraGPT command generation...")
    
    # Get credentials directly
    model_type, api_key = get_credentials(verbose=True)
    
    if not model_type or not api_key:
        print("Error: Couldn't get valid credentials")
        return 1
    
    # Create provider
    provider = get_provider(model_type, api_key, verbose=True, validate=False)
    
    # Create prompt
    prompt_template = create_prompt()
    prompt = "list all my VM instances"
    
    # Format prompt
    formatted_prompt = prompt_template.format(prompt=prompt)
    
    # Generate response directly using provider
    print("\nGenerating response...")
    result = provider.generate(formatted_prompt)
    
    print(f"\nPrompt: {prompt}")
    print(f"Result: {result}")
    
    return 0

if __name__ == "__main__":
    sys.exit(main())