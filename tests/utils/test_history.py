"""
Tests for history management functions.
"""
import json
import pathlib
from unittest.mock import patch, mock_open

import pytest

# Import helper functions for testing
from tests.utils.helpers import log_interaction_test, get_interaction_history_test


class TestHistory:
    """Tests for history management functions."""

    def test_log_interaction(self, temp_dir):
        """Test logging an interaction."""
        # Create temporary paths
        history_dir = temp_dir
        history_file = history_dir / "history.jsonl"
        
        # Log a test interaction
        test_data = {
            "model": "test-model",
            "prompt": "test prompt",
            "result": "test result"
        }
        log_interaction_test("test_type", test_data, history_file, history_dir)
        
        # Verify file was created with correct content
        assert history_file.exists(), "History file should be created"
        
        with open(history_file, "r") as f:
            line = f.readline().strip()
            entry = json.loads(line)
            
            assert entry["type"] == "test_type", "Entry type should match"
            assert entry["data"] == test_data, "Entry data should match"
            assert entry["id"] == "test-id", "Entry should have an ID"
            assert entry["timestamp"] == "2023-01-01T00:00:00", "Entry should have a timestamp"

    def test_log_interaction_exception(self, temp_dir):
        """Test handling exceptions during interaction logging."""
        # Create temporary paths
        history_dir = temp_dir
        history_file = history_dir / "history.jsonl"
        
        # Mock open to raise an exception
        m = mock_open()
        m.side_effect = Exception("Test exception")
        
        with patch("builtins.open", m):
            # This should not raise an exception
            log_interaction_test("test_type", {"test": "value"}, history_file, history_dir)
            # No assertion needed - we're testing that the exception is handled

    def test_get_interaction_history_not_exists(self, temp_dir):
        """Test getting history when file doesn't exist."""
        # Use a non-existent file in a temporary directory
        history_file = temp_dir / "nonexistent.jsonl"
        
        # Call the function and check result
        result = get_interaction_history_test(history_file)
        assert result == [], "Should return empty list when file doesn't exist"

    def test_get_interaction_history(self, history_file):
        """Test getting history entries."""
        file_path, entries = history_file
        
        # Get the history
        result = get_interaction_history_test(file_path)
        
        # Check result
        assert len(result) == len(entries), "Should return all entries"
        
        # Check the entries are in reverse order (newest first)
        assert result[0]["id"] == "test-id-3", "First entry should be the last one added"
        assert result[-1]["id"] == "test-id-1", "Last entry should be the first one added"

    def test_get_interaction_history_with_limit(self, history_file):
        """Test getting history entries with a limit."""
        file_path, entries = history_file
        
        # Get the history with limit
        result = get_interaction_history_test(file_path, limit=2)
        
        # Check result
        assert len(result) == 2, "Should return limited number of entries"
        
        # Check the entries are in reverse order (newest first)
        assert result[0]["id"] == "test-id-3", "First entry should be the last one added"
        assert result[1]["id"] == "test-id-2", "Second entry should be the second-to-last one added"

    def test_get_interaction_history_exception(self, temp_dir):
        """Test handling exceptions during history retrieval."""
        # Create a temporary file
        history_file = temp_dir / "history.jsonl"
        
        # Create an empty file
        with open(history_file, "w") as f:
            pass
        
        # Mock open function to raise an exception
        m = mock_open()
        m.side_effect = Exception("Test exception")
        
        with patch("builtins.open", m):
            result = get_interaction_history_test(history_file)
            assert result == [], "Should return empty list on exception"