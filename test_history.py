#!/usr/bin/env python3
"""Test script to verify that the history command is working correctly."""

import sys
import subprocess

def main():
    """Run a test of the history command."""
    print("Testing the history command...")
    
    # Run the history command with limit=1
    cmd = ["python", "-m", "infragpt", "history", "--limit", "1"]
    print(f"Running command: {' '.join(cmd)}")
    
    try:
        result = subprocess.run(cmd, check=True, capture_output=True, text=True)
        print("\nCommand output:")
        print(result.stdout)
        print("History command succeeded!")
    except subprocess.CalledProcessError as e:
        print(f"\nCommand failed with exit code {e.returncode}")
        print("STDOUT:")
        print(e.stdout)
        print("STDERR:")
        print(e.stderr)
        return 1
    
    return 0

if __name__ == "__main__":
    sys.exit(main())