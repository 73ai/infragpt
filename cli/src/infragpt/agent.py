#!/usr/bin/env python3
"""
Modern InfraGPT Shell Agent using LangChain 2025 standards.
"""

import sys
import signal
from typing import List, Dict, Any, Optional
from dataclasses import dataclass
from datetime import datetime
from collections import deque

from rich.console import Console
from rich.panel import Panel

from infragpt.tools import get_tools
from infragpt.history import log_interaction
from infragpt.llm.models import MODEL_TYPE
from infragpt.llm.prompts import get_prompt_template
from infragpt.llm_adapter import get_llm_with_tools

# Initialize console for rich output
console = Console()


@dataclass
class Message:
    """Represents a message in the conversation."""
    role: str  # 'user', 'assistant', 'system', 'tool'
    content: str
    timestamp: datetime
    tool_calls: Optional[List[Dict]] = None
    tool_call_id: Optional[str] = None


class ConversationContext:
    """Manages conversation context with message history."""
    
    def __init__(self, max_messages: int = 5):
        """Initialize conversation context."""
        self.max_messages = max_messages
        self.messages: List[Message] = []
        self.system_message = None
    
    def add_message(self, role: str, content: str, tool_calls: Optional[List[Dict]] = None,
                   tool_call_id: Optional[str] = None):
        """Add a message to the conversation context."""
        message = Message(
            role=role,
            content=content,
            timestamp=datetime.now(),
            tool_calls=tool_calls,
            tool_call_id=tool_call_id
        )
        
        # Keep system message separate
        if role == 'system':
            self.system_message = message
        else:
            self.messages.append(message)
            
            # Maintain context window
            if len(self.messages) > self.max_messages:
                self.messages = self.messages[-self.max_messages:]
    
    def get_context_messages(self) -> List[Dict[str, Any]]:
        """Get messages formatted for LLM API."""
        context = []
        
        # Add system message first
        if self.system_message:
            context.append({
                'role': 'system',
                'content': self.system_message.content
            })
        
        # Add conversation messages
        for msg in self.messages:
            message_dict = {
                'role': msg.role,
                'content': msg.content
            }
            
            # Add tool calls if present (skip for Claude as it uses content format)
            if msg.tool_calls and msg.role == 'assistant':
                # For Claude, tool calls are part of the content structure
                # Skip adding tool_calls to the message dict for now
                pass
            
            # Add tool call ID if this is a tool result
            if msg.tool_call_id and msg.role == 'tool':
                message_dict['tool_call_id'] = msg.tool_call_id
                message_dict['name'] = 'execute_shell_command'  # Add tool name for Claude
            
            context.append(message_dict)
        
        return context
    
    def clear(self):
        """Clear conversation context."""
        self.messages = []


class ModernShellAgent:
    """Modern shell agent using LangChain 2025 standards."""
    
    def __init__(self, model_type: MODEL_TYPE, api_key: str, verbose: bool = False):
        """Initialize shell agent."""
        self.model_type = model_type
        self.api_key = api_key
        self.verbose = verbose
        self.context = ConversationContext()
        self.should_exit = False
        self.tools = get_tools()
        
        # Create model with tools bound using modern LangChain
        self.model = get_llm_with_tools(
            model_type=model_type,
            api_key=api_key,
            verbose=verbose,
            tools=self.tools
        )
        
        # Set up signal handlers
        signal.signal(signal.SIGINT, self._signal_handler)
        
        # Initialize system message
        self._initialize_system_message()
    
    def _initialize_system_message(self):
        """Initialize the system message for the agent."""
        system_prompt = get_prompt_template('shell_agent')
        self.context.add_message('system', system_prompt)
    
    def _signal_handler(self, signum, frame):
        """Handle interrupt signals."""
        self.should_exit = True
        console.print("\n[yellow]Exiting...[/yellow]")
    
    def run_interactive_session(self):
        """Run the main interactive agent session."""
        console.print(Panel.fit(
            f"InfraGPT Shell Agent - Interactive debugging and operations assistant",
            border_style="blue",
            title="[bold green]Shell Agent[/bold green]"
        ))
        
        console.print(f"[yellow]Model:[/yellow] [bold]{self.model_type}[/bold]")
        console.print("[dim]Press Ctrl+C to exit the session[/dim]\n")
        
        # Show initial prompt
        console.print("[bold cyan]What would you like me to help with?[/bold cyan]")
        
        while not self.should_exit:
            try:
                # Get user input
                user_input = self._get_user_input()
                if not user_input or self.should_exit:
                    break
                
                # Add user message to context
                self.context.add_message('user', user_input)
                
                # Process with LLM
                self._process_user_input(user_input)
                
            except KeyboardInterrupt:
                break
            except EOFError:
                break
        
        console.print("\n[bold]Goodbye![/bold]")
    
    def _get_user_input(self) -> str:
        """Get user input with prompt."""
        try:
            return console.input("[bold cyan]> [/bold cyan]")
        except (KeyboardInterrupt, EOFError):
            return ""
    
    def _process_user_input(self, user_input: str):
        """Process user input with modern LangChain streaming."""
        try:
            # Continue processing until no more tool calls
            while True:
                # Get context messages
                messages = self.context.get_context_messages()
                
                # Show thinking and stream response
                console.print("\n[dim]Thinking...[/dim]")
                
                response_content = ""
                tool_calls = []
                first_content = True
                tool_accumulator = {}  # Track partial tool calls being built
                
                # Use modern LangChain streaming
                for chunk in self.model.stream(messages):
                    # Handle content streaming
                    if chunk.content:
                        if first_content:
                            # Clear thinking message and show A: prefix
                            console.print("\033[1A\033[K", end="")  # Move up and clear line
                            console.print("[bold green]A:[/bold green] ", end="")
                            first_content = False
                    
                        # Handle both string and list content (Claude returns list)
                        if isinstance(chunk.content, str):
                            content_text = chunk.content
                        elif isinstance(chunk.content, list):
                            # Extract text from Claude's structured format
                            content_text = ""
                            for item in chunk.content:
                                if isinstance(item, dict):
                                    if item.get('type') == 'text':
                                        content_text += item.get('text', '')
                                    elif item.get('type') == 'tool_use':
                                        # Start tracking a new tool call
                                        tool_id = item.get('id', '')
                                        tool_accumulator[tool_id] = {
                                            'id': str(tool_id),
                                            'name': item.get('name', ''),
                                            'args': {},
                                            'json_parts': []
                                        }
                                    elif item.get('type') == 'input_json_delta':
                                        # Accumulate JSON parts
                                        if tool_accumulator:
                                            # Add to the last tool being built
                                            last_tool_id = list(tool_accumulator.keys())[-1]
                                            tool_accumulator[last_tool_id]['json_parts'].append(
                                                item.get('partial_json', '')
                                            )
                                elif isinstance(item, str):
                                    content_text += item
                        else:
                            content_text = str(chunk.content)
                        
                        response_content += content_text
                        console.print(content_text, end="")
                    
                    # Standard tool calls (OpenAI format)
                    if hasattr(chunk, 'tool_calls') and chunk.tool_calls:
                        tool_calls = chunk.tool_calls
            
                # Build final tool calls from accumulated data
                if tool_accumulator:
                    import json
                    for tool_data in tool_accumulator.values():
                        # Reconstruct the JSON arguments
                        if tool_data['json_parts']:
                            json_str = ''.join(tool_data['json_parts'])
                            try:
                                tool_data['args'] = json.loads(json_str)
                            except json.JSONDecodeError:
                                if self.verbose:
                                    console.print(f"[dim]Failed to parse tool JSON: {json_str}[/dim]")
                        
                        tool_calls.append({
                            'id': tool_data['id'],
                            'name': tool_data['name'],
                            'args': tool_data['args']
                        })
                
                # Add newline after streaming
                if response_content:
                    console.print()
                
                # Debug tool calls
                if self.verbose and tool_calls:
                    console.print(f"[dim]Found {len(tool_calls)} tool calls[/dim]")
                    for tc in tool_calls:
                        console.print(f"[dim]Tool call: {tc}[/dim]")
            
                # Handle tool calls if present
                if tool_calls:
                    self._handle_tool_calls(tool_calls, response_content)
                    # Continue the loop to process LLM's response to tool results
                    if self.should_exit:
                        break
                else:
                    # No tool calls, just add assistant message and exit loop
                    self.context.add_message('assistant', response_content)
                    break
            
            # Log interaction (moved outside the loop)
            self._log_interaction(user_input, response_content, [])
            
        except Exception as e:
            console.print(f"[bold red]Error processing input:[/bold red] {e}")
            if self.verbose:
                import traceback
                console.print(traceback.format_exc())
    
    def _handle_tool_calls(self, tool_calls: List[Dict], response_content: str):
        """Handle tool calls with user confirmation using modern patterns."""
        try:
            # Add assistant message with tool calls
            self.context.add_message('assistant', response_content, tool_calls=tool_calls)
            
            for tool_call in tool_calls:
                if self.should_exit:
                    break
                
                # Extract command from tool call
                tool_name = tool_call.get('name', '')
                tool_args = tool_call.get('args', {})
                
                # Skip empty or invalid tool calls
                if not tool_name or tool_name not in [t.name for t in self.tools]:
                    if self.verbose:
                        console.print(f"[dim]Skipping invalid tool call: {tool_call}[/dim]")
                    continue
                
                command = tool_args.get('command', '')
                
                if not command:
                    console.print(f"[bold red]No command found in tool call[/bold red]")
                    continue
                
                # Ask for confirmation
                console.print(f"\n[bold yellow]Running command:[/bold yellow] {command}")
                
                try:
                    user_response = console.input(f'Run "{command}"? (Y/n): ').strip().lower()
                    if user_response == "n":
                        console.print("[dim]Exiting...[/dim]")
                        self.should_exit = True
                        return
                    elif user_response not in ["", "y"]:
                        console.print("[bold red]Invalid input. Please enter 'y' or 'n'.[/bold red]")
                        continue
                except (KeyboardInterrupt, EOFError):
                    self.should_exit = True
                    return
                
                # Execute the tool using modern LangChain
                try:
                    # Get the actual tool function
                    tool_func = next((t for t in self.tools if t.name == tool_name), None)
                    if tool_func:
                        result = tool_func.invoke(tool_args)
                        
                        # Add tool result to context
                        self.context.add_message(
                            'tool',
                            result,
                            tool_call_id=str(tool_call.get('id', ''))  # Ensure string ID
                        )
                        
                        # The output is already displayed by the shell executor in real-time
                        # so we don't need to print it again here
                    else:
                        console.print(f"[bold red]Unknown tool: {tool_name}[/bold red]")
                        
                except Exception as e:
                    error_msg = f"Error executing command: {str(e)}"
                    console.print(f"[bold red]{error_msg}[/bold red]")
                    self.context.add_message(
                        'tool',
                        error_msg,
                        tool_call_id=str(tool_call.get('id', ''))  # Ensure string ID
                    )
                    
        except Exception as e:
            console.print(f"[bold red]Error handling tool calls:[/bold red] {e}")
            if self.verbose:
                import traceback
                console.print(traceback.format_exc())
    
    def _log_interaction(self, user_input: str, response: str, tool_calls: List[Dict]):
        """Log the interaction for history."""
        try:
            interaction_data = {
                "user_input": user_input,
                "assistant_response": response,
                "tool_calls": [
                    {
                        "name": tc.get('name', ''),
                        "arguments": tc.get('args', {})
                    }
                    for tc in tool_calls
                ],
                "model": self.model_type,
                "timestamp": datetime.now().isoformat()
            }
            log_interaction("agent_conversation", interaction_data)
        except Exception:
            # Don't let logging failures interrupt the session
            pass


def run_shell_agent(model_type: MODEL_TYPE, api_key: str, verbose: bool = False):
    """Run the modern shell agent."""
    agent = ModernShellAgent(model_type, api_key, verbose)
    agent.run_interactive_session()