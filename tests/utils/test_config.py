"""
Tests for configuration utility functions.
"""
import pathlib
import sys
import os
from unittest.mock import patch, mock_open

import pytest
import yaml

# Import helper functions for testing
from tests.utils.helpers import load_config_test, save_config_test


class TestConfig:
    """Tests for configuration management functions."""

    def test_load_config_file_not_exists(self):
        """Test loading config when file doesn't exist."""
        # Use a non-existent file path
        config_file = pathlib.Path("/non/existent/path")
        
        # Call the function and check result
        result = load_config_test(config_file)
        assert result == {}, "Should return empty dict when file doesn't exist"

    def test_load_config_valid(self, temp_dir):
        """Test loading valid config file."""
        # Create a temporary config file
        config_file = temp_dir / "config.yaml"
        test_config = {"model": "test-model", "api_key": "test-key"}
        
        with open(config_file, "w") as f:
            yaml.dump(test_config, f)
        
        # Load the config
        result = load_config_test(config_file)
        
        # Check result
        assert result == test_config, "Should return config contents as dict"

    def test_load_config_exception(self, tmp_path):
        """Test handling exceptions during config loading."""
        # Create a real file
        config_file = tmp_path / "config.yaml"
        with open(config_file, "w") as f:
            f.write("test")
            
        # Mock open function to raise an exception
        m = mock_open()
        m.side_effect = Exception("Test exception")
        
        with patch("builtins.open", m):
            result = load_config_test(config_file)
            assert result == {}, "Should return empty dict on exception"

    def test_save_config(self, temp_dir):
        """Test saving config to file."""
        # Create temporary paths
        config_dir = temp_dir
        config_file = config_dir / "config.yaml"
        
        # Save a test config
        test_config = {"model": "test-model", "api_key": "test-key"}
        save_config_test(test_config, config_file, config_dir)
        
        # Verify file was created with correct content
        assert config_file.exists(), "Config file should be created"
        
        with open(config_file, "r") as f:
            saved_config = yaml.safe_load(f)
            assert saved_config == test_config, "Saved config should match input"

    def test_save_config_exception(self, temp_dir):
        """Test handling exceptions during config saving."""
        # Use real temporary directory
        config_dir = temp_dir
        config_file = config_dir / "config.yaml"
        
        # Mock open function to raise an exception
        m = mock_open()
        m.side_effect = Exception("Test exception")
        
        with patch("builtins.open", m):
            # This should not raise an exception
            save_config_test({"test": "value"}, config_file, config_dir)
            # No assertion needed - we're testing that the exception is handled