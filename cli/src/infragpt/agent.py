"""
Modern InfraGPT Shell Agent using direct SDKs instead of LangChain.
"""

from typing import List, Dict, Any, Optional
from datetime import datetime

from rich.console import Console
from rich.panel import Panel

from .llm_adapter import get_llm_adapter
from .history import log_interaction
from .tools import ToolExecutionCancelled

# Import prompt_toolkit for better input handling
import pathlib
from prompt_toolkit import PromptSession
from prompt_toolkit.history import FileHistory


# Create a simple system prompt instead of using the old prompts module
def get_system_prompt():
    return """You are an intelligent shell operations assistant. You help users with:

1. Infrastructure and system administration tasks
2. Debugging and troubleshooting issues  
3. Running shell commands safely with user confirmation
4. Analyzing system logs and performance

You have access to shell command execution tools. Always:
- Ask for confirmation before running commands
- Explain what commands do before executing them
- Be cautious with destructive operations
- Provide helpful context and suggestions

Be concise but thorough in your responses."""


# Initialize console for rich output
console = Console()


class ConversationContext:
    """Manages conversation context with message history."""

    def __init__(self, max_messages: int = 5):
        """Initialize conversation context."""
        self.max_messages = max_messages
        self.messages: List[Dict[str, Any]] = []
        self.system_message = None

    def add_message(
        self,
        role: str,
        content: str,
        tool_calls: Optional[List[Dict]] = None,
        tool_call_id: Optional[str] = None,
    ):
        """Add a message to the conversation context."""
        message_dict = {"role": role, "content": content}

        if tool_calls:
            message_dict["tool_calls"] = tool_calls

        if tool_call_id:
            message_dict["tool_call_id"] = tool_call_id
            message_dict["name"] = (
                "execute_shell_command"  # Add tool name for compatibility
            )

        # Keep system message separate
        if role == "system":
            self.system_message = message_dict
        else:
            self.messages.append(message_dict)

            # Maintain context window
            if len(self.messages) > self.max_messages:
                self.messages = self.messages[-self.max_messages :]

    def get_context_messages(self) -> List[Dict[str, Any]]:
        """Get messages formatted for LLM API."""
        context = []

        # Add system message first
        if self.system_message:
            context.append(self.system_message)

        # Add conversation messages
        context.extend(self.messages)

        return context

    def clear(self):
        """Clear conversation context."""
        self.messages = []


class ModernShellAgent:
    """Modern shell agent using direct SDK integration."""

    def __init__(self, model_string: str, api_key: str, verbose: bool = False):
        """Initialize shell agent."""
        self.model_string = model_string
        self.api_key = api_key
        self.verbose = verbose
        self.context = ConversationContext()

        # Create LLM adapter
        self.llm_adapter = get_llm_adapter(
            model_string=model_string, api_key=api_key, verbose=verbose
        )

        # Initialize command history session
        self._setup_command_history()

        # Don't set up custom signal handlers - let prompt_toolkit handle them

        # Initialize system message
        self._initialize_system_message()

    def _setup_command_history(self):
        """Setup command history with persistent storage."""
        try:
            # Create history directory following InfraGPT conventions
            history_dir = pathlib.Path.home() / ".infragpt"
            history_dir.mkdir(exist_ok=True)

            # History file for command-line input
            history_file = history_dir / "history"

            # WARNING: FileHistory writes clear-text inputs to disk.
            # This is a security risk if users enter secrets at the prompt.
            # Create PromptSession with FileHistory
            self.prompt_session = PromptSession(history=FileHistory(str(history_file)))

            # Best-effort: restrict file permissions to user-only
            try:
                history_file.chmod(0o600)
            except Exception:
                pass

            if self.verbose:
                console.print(f"[dim]Command history: {history_file}[/dim]")

        except Exception as e:
            # Fallback to no history if setup fails
            self.prompt_session = PromptSession()
            if self.verbose:
                console.print(
                    f"[dim]Warning: Could not setup command history: {e}[/dim]"
                )

    def _initialize_system_message(self):
        """Initialize the system message for the agent."""
        system_prompt = get_system_prompt()
        self.context.add_message("system", system_prompt)

    def run_interactive_session(self):
        """Run the main interactive agent session."""
        console.print(
            Panel.fit(
                "InfraGPT Shell Agent V2 - Direct SDK Integration",
                border_style="blue",
                title="[bold green]Shell Agent V2[/bold green]",
            )
        )

        console.print(f"[yellow]Model:[/yellow] [bold]{self.model_string}[/bold]")

        # Validate API key
        console.print("[dim]Validating API key...[/dim]")
        try:
            if self.llm_adapter.validate_api_key():
                console.print("[green]✓ API key validated[/green]")
            else:
                console.print("[red]✗ API key validation failed[/red]")
                return
        except Exception as e:
            console.print(f"[red]✗ API key validation failed: {e}[/red]")
            return

        # Show initial prompt
        console.print("[bold cyan]What would you like me to help with?[/bold cyan]")
        console.print(
            "[dim]Press Ctrl+D to exit, Ctrl+C to interrupt operations[/dim]\n"
        )

        while True:
            try:
                # Get user input
                user_input = self._get_user_input()
                if not user_input:
                    continue  # Go back to prompt for empty input

                # Check for exit commands
                if user_input.lower() in ["exit", "quit", "bye"]:
                    break

                # Add user message to context
                self.context.add_message("user", user_input)

                # Process with LLM
                self._process_user_input(user_input)

            except KeyboardInterrupt:
                # This shouldn't happen with prompt_toolkit, but handle just in case
                continue
            except EOFError:
                # Ctrl+D - exit the application
                console.print("\n[dim]EOF received (Ctrl+D). Exiting...[/dim]")
                break

        console.print("\n[bold]Goodbye![/bold]")

    def _get_user_input(self) -> str:
        """Get user input with prompt - use prompt_toolkit for proper interactive features."""
        try:
            return self.prompt_session.prompt("> ")
        except KeyboardInterrupt:
            # Ctrl+C should just return empty string to continue
            return ""
        except EOFError:
            # Ctrl+D - let this propagate for proper exit handling
            raise

    def _process_user_input(self, user_input: str):
        """Process user input with direct SDK streaming and tool execution."""
        try:
            # Get context messages
            messages = self.context.get_context_messages()

            # Debug: Show message structure if verbose
            if self.verbose:
                console.print(f"[dim]Context has {len(messages)} messages[/dim]")
                for i, msg in enumerate(messages):
                    role = msg.get("role", "unknown")
                    has_tools = "tool_calls" in msg
                    has_tool_id = "tool_call_id" in msg
                    content_len = len(str(msg.get("content", "")))
                    console.print(
                        f"[dim]  {i}: {role} (content: {content_len} chars, tools: {has_tools}, tool_id: {has_tool_id})[/dim]"
                    )

            # Show thinking and stream response
            console.print("\n[dim]Thinking...[/dim]")

            response_content = ""
            first_content = True

            # Stream response using new adapter with interrupt checking
            # Note: stream_with_tools already handles the complete tool execution cycle
            # including getting the final response after tool execution
            try:
                for chunk in self.llm_adapter.stream_with_tools(messages):
                    # Handle content streaming
                    if chunk.content:
                        if first_content:
                            # Clear thinking message and show A: prefix
                            console.print(
                                "\033[1A\033[K", end=""
                            )  # Move up and clear line
                            console.print("[bold green]A:[/bold green] ", end="")
                            first_content = False

                        response_content += chunk.content
                        console.print(chunk.content, end="")

                    # Handle finish reason
                    if chunk.finish_reason:
                        if self.verbose:
                            console.print(
                                f"\n[dim]Finish reason: {chunk.finish_reason}[/dim]"
                            )
            except KeyboardInterrupt:
                console.print("\n[yellow]Operation cancelled by user.[/yellow]")
                return

            # Add newline after streaming
            if response_content:
                console.print()

            # Add assistant message to context if we have content
            if response_content:
                self.context.add_message("assistant", response_content)

            # Log interaction
            self._log_interaction(user_input, response_content)

        except ToolExecutionCancelled:
            # User cancelled tool execution - just return to prompt
            # No need to print anything - the tool already printed a message
            return
        except KeyboardInterrupt:
            # Ctrl+C during streaming - just return to prompt (message already printed by LLM adapter)
            return
        except Exception as e:
            console.print(f"[bold red]Error processing input:[/bold red] {e}")
            if self.verbose:
                import traceback

                console.print(traceback.format_exc())

    def _log_interaction(self, user_input: str, response: str):
        """Log the interaction for history. Sensitive fields are excluded explicitly."""
        try:
            # Only safe fields, do not include api_key or any secrets
            interaction_data = {
                "user_input": user_input,
                "assistant_response": response,
                "model": self.model_string,
                "timestamp": datetime.utcnow().isoformat() + "Z",
                # Not persisted: allow-list will drop this; used only for error reporting in history.py
                "verbose": self.verbose,
            }
            # Note: self.api_key is NOT logged ever
            log_interaction("agent_conversation_v2", interaction_data)
        except Exception as e:
            # Don't let logging failures interrupt the session
            if self.verbose:
                console.print(f"[dim]Warning: Could not log interaction: {e}[/dim]")


def run_shell_agent(model_string: str, api_key: str, verbose: bool = False):
    """Run the modern shell agent."""
    agent = ModernShellAgent(model_string, api_key, verbose)
    agent.run_interactive_session()
