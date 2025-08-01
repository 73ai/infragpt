#!/usr/bin/env python3
"""
Shell command execution module for InfraGPT CLI agent.

This module provides functionality for executing shell commands with:
- Real-time streaming output
- TTY support for interactive commands
- Timeout handling (1 minute default)
- ESC key for early termination
- Environment variable persistence across commands
"""

import os
import sys
import signal
import subprocess
import threading
import time
from typing import Optional, Dict, Any, Tuple
import pty
import select

from rich.console import Console
from rich.live import Live
from rich.text import Text

# Initialize console for rich output
console = Console()


class CommandExecutor:
    """Handles shell command execution with streaming and timeout."""
    
    def __init__(self, timeout: int = 60, env: Optional[Dict[str, str]] = None):
        """
        Initialize command executor.
        
        Args:
            timeout: Command timeout in seconds (default: 60)
            env: Environment variables to persist across commands
        """
        self.timeout = timeout
        self.env = env or os.environ.copy()
        self.current_process = None
        self.cancelled = False
        self.output_buffer = []
        
    def execute_command(self, command: str) -> Tuple[int, str, bool]:
        """
        Execute a shell command with streaming output and timeout.
        
        Args:
            command: Shell command to execute
            
        Returns:
            Tuple of (exit_code, output, was_cancelled)
        """
        self.cancelled = False
        self.output_buffer = []
        
        console.print(f"[bold cyan]Executing:[/bold cyan] {command}")
        console.print("[dim]Press ESC to cancel command...[/dim]\n")
        
        try:
            # Create a pseudo-terminal for better command interaction
            master_fd, slave_fd = pty.openpty()
            
            # Start the command
            # Note: preexec_fn might not work on all platforms, so we handle it carefully
            popen_args = {
                'shell': True,
                'stdin': slave_fd,
                'stdout': slave_fd,
                'stderr': slave_fd,
                'env': self.env
            }
            
            # Only use preexec_fn on Unix-like systems (not Windows)
            if hasattr(os, 'setsid'):
                try:
                    popen_args['preexec_fn'] = os.setsid
                except:
                    pass  # Skip if not supported
            
            self.current_process = subprocess.Popen(command, **popen_args)
            
            # Close slave fd in parent process
            os.close(slave_fd)
            
            # Start timeout timer
            timer = threading.Timer(self.timeout, self._timeout_handler)
            timer.start()
            
            # Start ESC key listener
            esc_thread = threading.Thread(target=self._esc_listener, daemon=True)
            esc_thread.start()
            
            # Stream output
            output = self._stream_output(master_fd)
            
            # Wait for process to complete
            exit_code = self.current_process.wait()
            
            # Cancel timer
            timer.cancel()
            
            # Close master fd
            os.close(master_fd)
            
            if self.cancelled:
                console.print("\n[bold yellow]Command cancelled by user[/bold yellow]")
                return -1, output, True
            
            # Display exit code
            if exit_code == 0:
                console.print(f"\n[bold green]Command completed successfully (exit code: {exit_code})[/bold green]")
            else:
                console.print(f"\n[bold red]Command failed (exit code: {exit_code})[/bold red]")
                
            return exit_code, output, False
            
        except Exception as e:
            console.print(f"[bold red]Error executing command:[/bold red] {e}")
            return -1, str(e), False
        finally:
            self.current_process = None
    
    def _stream_output(self, fd: int) -> str:
        """Stream output from command in real-time."""
        output_lines = []
        
        try:
            while True:
                # Check if process is still running
                if self.current_process and self.current_process.poll() is not None:
                    # Process finished, read any remaining output
                    try:
                        ready, _, _ = select.select([fd], [], [], 0.1)
                        if ready:
                            data = os.read(fd, 4096).decode('utf-8', errors='replace')
                            if data:
                                output_lines.append(data)
                                console.print(data, end='')
                    except (OSError, ValueError):
                        pass
                    break
                
                # Check for available data
                try:
                    ready, _, _ = select.select([fd], [], [], 0.1)
                    if ready:
                        data = os.read(fd, 4096).decode('utf-8', errors='replace')
                        if data:
                            output_lines.append(data)
                            console.print(data, end='')
                    
                    # Check if cancelled
                    if self.cancelled:
                        break
                        
                except (OSError, ValueError):
                    # FD closed or invalid
                    break
                except KeyboardInterrupt:
                    # Handle Ctrl+C gracefully
                    self._terminate_command()
                    break
                    
        except Exception as e:
            console.print(f"[bold red]Error streaming output:[/bold red] {e}")
        
        return ''.join(output_lines)
    
    def _timeout_handler(self):
        """Handle command timeout."""
        if self.current_process and self.current_process.poll() is None:
            console.print(f"\n[bold yellow]Command timed out after {self.timeout} seconds[/bold yellow]")
            self._terminate_command()
    
    def _terminate_command(self):
        """Terminate the current command."""
        if self.current_process:
            try:
                # Send SIGTERM to the process group
                os.killpg(os.getpgid(self.current_process.pid), signal.SIGTERM)
                
                # Wait a bit for graceful shutdown
                time.sleep(1)
                
                # Force kill if still running
                if self.current_process.poll() is None:
                    os.killpg(os.getpgid(self.current_process.pid), signal.SIGKILL)
            except (OSError, ProcessLookupError):
                # Process already terminated
                pass
    
    def _esc_listener(self):
        """Listen for ESC key to cancel command."""
        try:
            import termios
            import tty
            
            # Save terminal settings
            old_settings = termios.tcgetattr(sys.stdin)
            
            try:
                # Set terminal to raw mode for immediate key detection
                tty.cbreak(sys.stdin.fileno())
                
                while self.current_process and self.current_process.poll() is None:
                    if sys.stdin in select.select([sys.stdin], [], [], 0.1)[0]:
                        key = sys.stdin.read(1)
                        if key == '\x1b':  # ESC key
                            self.cancelled = True
                            self._terminate_command()
                            break
            finally:
                # Restore terminal settings
                termios.tcsetattr(sys.stdin, termios.TCSADRAIN, old_settings)
                
        except ImportError:
            # termios not available (Windows), skip ESC handling
            pass
        except Exception:
            # Other errors, skip ESC handling
            pass
    
    def update_environment(self, env_vars: Dict[str, str]):
        """Update environment variables for future commands."""
        self.env.update(env_vars)
    
    def get_environment(self) -> Dict[str, str]:
        """Get current environment variables."""
        return self.env.copy()


def execute_shell_command(command: str, timeout: int = 60, 
                         executor: Optional[CommandExecutor] = None) -> Tuple[int, str, bool]:
    """
    Execute a shell command with the given executor or create a new one.
    
    Args:
        command: Shell command to execute
        timeout: Command timeout in seconds
        executor: Existing CommandExecutor instance to reuse environment
        
    Returns:
        Tuple of (exit_code, output, was_cancelled)
    """
    if executor is None:
        executor = CommandExecutor(timeout=timeout)
    
    return executor.execute_command(command)


def parse_environment_changes(output: str) -> Dict[str, str]:
    """
    Parse command output for environment variable changes.
    
    This is a simple implementation that looks for export statements
    in the output. More sophisticated parsing could be added later.
    
    Args:
        output: Command output to parse
        
    Returns:
        Dictionary of environment variable changes
    """
    env_changes = {}
    
    # Look for export statements in output
    lines = output.split('\n')
    for line in lines:
        line = line.strip()
        if line.startswith('export ') and '=' in line:
            try:
                # Parse export VAR=value
                export_part = line[7:]  # Remove 'export '
                if '=' in export_part:
                    var, value = export_part.split('=', 1)
                    # Remove quotes if present
                    value = value.strip('"\'')
                    env_changes[var] = value
            except Exception:
                # Skip malformed export statements
                pass
    
    return env_changes