"""History tracking for InfraGPT."""

import os
import json
import uuid
import datetime
import pathlib
import time
from typing import List, Dict, Any, Optional, Tuple, Union, Iterator

from ..ui.console import console
from .config import CONFIG_DIR, config

# Define history file locations
HISTORY_DIR = CONFIG_DIR / "history"
HISTORY_DB_FILE = HISTORY_DIR / "history.jsonl"

class HistoryManager:
    """Manager for interaction history."""
    
    def __init__(self, history_file: Optional[pathlib.Path] = None):
        """Initialize history manager."""
        self.history_file = history_file or HISTORY_DB_FILE
        self._ensure_history_dir()
    
    def _ensure_history_dir(self) -> None:
        """Ensure history directory exists."""
        self.history_file.parent.mkdir(parents=True, exist_ok=True)
    
    def add_entry(self, interaction_type: str, data: Dict[str, Any]) -> str:
        """Add an entry to the history.
        
        Returns:
            The ID of the added entry.
        """
        # Create the entry
        entry_id = str(uuid.uuid4())
        entry = {
            "id": entry_id,
            "timestamp": datetime.datetime.now().isoformat(),
            "type": interaction_type,
            "data": data,
            "session_id": os.environ.get("INFRAGPT_SESSION_ID", str(uuid.uuid4()))
        }
        
        # Append to history file
        try:
            self._ensure_history_dir()
            with open(self.history_file, "a") as f:
                f.write(json.dumps(entry) + "\n")
            return entry_id
        except Exception as e:
            # Silently fail - history logging should not interrupt user flow
            if 'verbose' in data and data.get('verbose'):
                console.print(f"[dim]Warning: Could not log interaction: {e}[/dim]")
            return entry_id
    
    def get_entries(self, 
                   limit: int = 100, 
                   interaction_type: Optional[str] = None,
                   start_date: Optional[Union[str, datetime.datetime]] = None,
                   end_date: Optional[Union[str, datetime.datetime]] = None,
                   search_term: Optional[str] = None) -> List[Dict[str, Any]]:
        """Get history entries with optional filtering.
        
        Args:
            limit: Maximum number of entries to return
            interaction_type: Filter by interaction type
            start_date: Filter entries after this date (inclusive)
            end_date: Filter entries before this date (inclusive)
            search_term: Filter entries containing this term
            
        Returns:
            List of history entries, newest first
        """
        if not self.history_file.exists():
            return []
        
        # Parse date filters if they're strings
        if isinstance(start_date, str):
            try:
                start_date = datetime.datetime.fromisoformat(start_date)
            except ValueError:
                start_date = None
        
        if isinstance(end_date, str):
            try:
                end_date = datetime.datetime.fromisoformat(end_date)
            except ValueError:
                end_date = None
        
        try:
            entries = []
            with open(self.history_file, "r") as f:
                for line_num, line in enumerate(f):
                    if not line.strip():
                        continue
                        
                    try:
                        entry = json.loads(line)
                        
                        # Apply filters
                        if interaction_type and entry.get("type") != interaction_type:
                            continue
                            
                        if start_date or end_date:
                            try:
                                entry_date = datetime.datetime.fromisoformat(entry.get("timestamp", ""))
                                if start_date and entry_date < start_date:
                                    continue
                                if end_date and entry_date > end_date:
                                    continue
                            except (ValueError, TypeError):
                                # Skip entries with invalid dates
                                continue
                        
                        if search_term:
                            # Very simple search implementation
                            entry_str = json.dumps(entry).lower()
                            if search_term.lower() not in entry_str:
                                continue
                        
                        entries.append(entry)
                    except json.JSONDecodeError:
                        console.print(f"[yellow]Warning:[/yellow] Invalid JSON at line {line_num + 1}")
                        continue
            
            # Return most recent entries first, up to the limit
            entries.sort(key=lambda x: x.get("timestamp", ""), reverse=True)
            return entries[:limit]
            
        except Exception as e:
            console.print(f"[yellow]Warning:[/yellow] Could not read history: {e}")
            return []
    
    def get_statistics(self) -> Dict[str, Any]:
        """Get statistics about the history.
        
        Returns:
            Dictionary with statistics like:
            - total_entries: Total number of entries
            - entries_by_type: Count of entries by type
            - first_entry_date: Date of first entry
            - last_entry_date: Date of last entry
        """
        if not self.history_file.exists():
            return {
                "total_entries": 0,
                "entries_by_type": {},
                "first_entry_date": None,
                "last_entry_date": None
            }
        
        try:
            # Initialize statistics
            stats = {
                "total_entries": 0,
                "entries_by_type": {},
                "first_entry_date": None,
                "last_entry_date": None,
                "top_commands": [],
                "top_parameters": {},
            }
            
            command_counts = {}
            parameter_counts = {}
            
            # First pass: count entries and get basic stats
            with open(self.history_file, "r") as f:
                for line in f:
                    if not line.strip():
                        continue
                    
                    try:
                        entry = json.loads(line)
                        stats["total_entries"] += 1
                        
                        # Count by type
                        entry_type = entry.get("type", "unknown")
                        stats["entries_by_type"][entry_type] = stats["entries_by_type"].get(entry_type, 0) + 1
                        
                        # Track dates
                        try:
                            entry_date = datetime.datetime.fromisoformat(entry.get("timestamp", ""))
                            if stats["first_entry_date"] is None or entry_date < stats["first_entry_date"]:
                                stats["first_entry_date"] = entry_date
                            if stats["last_entry_date"] is None or entry_date > stats["last_entry_date"]:
                                stats["last_entry_date"] = entry_date
                        except (ValueError, TypeError):
                            pass
                        
                        # Track command usage
                        if entry_type == "command_generation":
                            command = entry.get("data", {}).get("result", "").strip()
                            if command:
                                command_base = command.split()[0] if " " in command else command
                                command_counts[command_base] = command_counts.get(command_base, 0) + 1
                        
                        # Track parameter usage
                        if entry_type == "command_action":
                            params = entry.get("data", {}).get("parameters", {})
                            for param, value in params.items():
                                if param not in parameter_counts:
                                    parameter_counts[param] = {}
                                parameter_counts[param][value] = parameter_counts[param].get(value, 0) + 1
                        
                    except json.JSONDecodeError:
                        continue
            
            # Process top commands
            top_commands = sorted(
                [{"command": cmd, "count": count} for cmd, count in command_counts.items()],
                key=lambda x: x["count"], 
                reverse=True
            )[:10]
            stats["top_commands"] = top_commands
            
            # Process top parameters
            stats["top_parameters"] = {}
            for param, values in parameter_counts.items():
                top_values = sorted(
                    [{"value": val, "count": count} for val, count in values.items()],
                    key=lambda x: x["count"],
                    reverse=True
                )[:5]
                stats["top_parameters"][param] = top_values
            
            return stats
                
        except Exception as e:
            console.print(f"[yellow]Warning:[/yellow] Could not calculate statistics: {e}")
            return {
                "total_entries": 0,
                "entries_by_type": {},
                "first_entry_date": None,
                "last_entry_date": None,
                "error": str(e)
            }
    
    def export_entries(self, file_path: str, 
                      format: str = "jsonl", 
                      **filter_kwargs) -> Tuple[bool, str]:
        """Export entries to a file.
        
        Args:
            file_path: Path to export to
            format: Format to export in (jsonl, csv, json)
            **filter_kwargs: Filters to apply (passed to get_entries)
            
        Returns:
            (success, message) tuple
        """
        entries = self.get_entries(**filter_kwargs)
        
        if not entries:
            return False, "No entries to export"
        
        try:
            if format == "jsonl":
                with open(file_path, "w") as f:
                    for entry in entries:
                        f.write(json.dumps(entry) + "\n")
            elif format == "json":
                with open(file_path, "w") as f:
                    json.dump(entries, f, indent=2)
            elif format == "csv":
                import csv
                fieldnames = ["id", "timestamp", "type"]
                
                # Get all data fields across all entries
                data_fields = set()
                for entry in entries:
                    data_fields.update(entry.get("data", {}).keys())
                
                for field in sorted(data_fields):
                    fieldnames.append(f"data.{field}")
                
                with open(file_path, "w", newline="") as f:
                    writer = csv.DictWriter(f, fieldnames=fieldnames)
                    writer.writeheader()
                    
                    for entry in entries:
                        row = {
                            "id": entry.get("id", ""),
                            "timestamp": entry.get("timestamp", ""),
                            "type": entry.get("type", "")
                        }
                        
                        # Flatten data fields
                        for field in data_fields:
                            row[f"data.{field}"] = entry.get("data", {}).get(field, "")
                        
                        writer.writerow(row)
            else:
                return False, f"Unsupported format: {format}"
            
            return True, f"Exported {len(entries)} entries to {file_path}"
            
        except Exception as e:
            return False, f"Export failed: {e}"
    
    def clear_history(self, older_than: Optional[Union[str, datetime.datetime]] = None) -> Tuple[bool, str]:
        """Clear history entries.
        
        Args:
            older_than: Clear entries older than this date
            
        Returns:
            (success, message) tuple
        """
        if not self.history_file.exists():
            return True, "No history to clear"
        
        # Parse date if it's a string
        if isinstance(older_than, str):
            try:
                older_than = datetime.datetime.fromisoformat(older_than)
            except ValueError:
                return False, f"Invalid date format: {older_than}"
        
        # If no date specified, delete the whole file
        if older_than is None:
            try:
                self.history_file.unlink()
                return True, "History cleared successfully"
            except Exception as e:
                return False, f"Failed to clear history: {e}"
        
        # Otherwise, filter entries
        try:
            # Get all entries
            all_entries = self.get_entries(limit=999999)  # Very high limit to get all entries
            
            # Filter entries to keep
            keep_entries = []
            for entry in all_entries:
                try:
                    entry_date = datetime.datetime.fromisoformat(entry.get("timestamp", ""))
                    if entry_date >= older_than:
                        keep_entries.append(entry)
                except (ValueError, TypeError):
                    # Keep entries with invalid dates (better safe than sorry)
                    keep_entries.append(entry)
            
            # Write back the entries to keep
            self.history_file.unlink()
            for entry in sorted(keep_entries, key=lambda x: x.get("timestamp", "")):
                self.add_entry(entry.get("type", "unknown"), entry.get("data", {}))
            
            cleared_count = len(all_entries) - len(keep_entries)
            return True, f"Cleared {cleared_count} entries older than {older_than.isoformat()}"
            
        except Exception as e:
            return False, f"Failed to clear history: {e}"

# Global history manager instance
history_manager = HistoryManager()

# Legacy functions for backward compatibility

def log_interaction(interaction_type: str, data: Dict[str, Any]) -> None:
    """Log user interaction to the history database file."""
    history_manager.add_entry(interaction_type, data)

def get_interaction_history(limit: int = 100, interaction_type: Optional[str] = None) -> List[Dict[str, Any]]:
    """Retrieve the most recent interaction history entries."""
    return history_manager.get_entries(limit=limit, interaction_type=interaction_type)