"""Command generation and processing pipeline for InfraGPT."""

import re
import datetime
import json
import abc
from typing import List, Optional, Dict, Any, Tuple, Union, Protocol, Callable, TypeVar

from langchain_core.output_parsers import StrOutputParser

from ..ui.console import console
from ..utils.config import Config
from ..utils.history import log_interaction
from .llm import get_provider, MODEL_TYPE, BaseLLMProvider
from .prompt import create_prompt, create_parameter_prompt

# Type definitions
T = TypeVar('T')
CommandInput = str
CommandOutput = str
ProcessedCommand = str
ParameterDict = Dict[str, Any]

class PipelineComponent(Protocol[T]):
    """Protocol defining a component in the command processing pipeline."""

    def process(self, input_data: Any, context: Dict[str, Any]) -> T:
        """Process input data and return transformed output.

        Args:
            input_data: The input data to process
            context: Additional context data for the processing

        Returns:
            Processed output data
        """
        ...


class CommandInputProcessor(PipelineComponent[CommandInput]):
    """Component that processes raw user input."""

    def process(self, input_data: str, context: Dict[str, Any]) -> CommandInput:
        """Process user input for command generation.

        Args:
            input_data: Raw user input
            context: Additional context for processing

        Returns:
            Processed input ready for command generation
        """
        # Basic preprocessing - trim whitespace, handle special characters
        processed_input = input_data.strip()
        context['original_input'] = input_data
        context['processed_input'] = processed_input
        return processed_input


class CommandGenerator(PipelineComponent[CommandOutput]):
    """Component that generates commands from processed input using LLM."""

    def __init__(self, provider: Optional[BaseLLMProvider] = None, system_prompt: Optional[str] = None):
        """Initialize with an optional LLM provider and system prompt.

        Args:
            provider: The LLM provider to use for command generation
            system_prompt: Optional custom system prompt to override the default
        """
        self.provider = provider
        self.system_prompt = system_prompt

    def process(self, input_data: CommandInput, context: Dict[str, Any]) -> CommandOutput:
        """Generate commands using the LLM provider.

        Args:
            input_data: Processed user input
            context: Additional context including LLM settings

        Returns:
            Generated command string
        """
        # Get or create provider if not already initialized
        provider = self.provider
        if not provider:
            model_type = context.get('model_type')
            api_key = context.get('api_key')
            verbose = context.get('verbose', False)
            use_cache = context.get('use_cache', True)
            provider = get_provider(model_type, api_key, verbose, validate=True, use_cache=use_cache)

        # Get model info for logging
        model_name = provider.get_model_name()
        provider_name = provider.get_name()

        # Log verbose info
        if context.get('verbose'):
            console.print(f"[dim]Generating command using {provider_name} ({model_name})...[/dim]")

        # Get prompt with optional system prompt override
        prompt_template = create_prompt(self.system_prompt)
        prompt_text = prompt_template.format(prompt=input_data)

        # Generate response
        start_time = datetime.datetime.now()

        # Don't show a status display here since it's shown in the calling function
        # Just generate the command directly
        result = provider.generate(prompt_text)

        end_time = datetime.datetime.now()

        # Store result and metadata in context
        context['raw_result'] = result.strip()
        context['generation_time'] = (end_time - start_time).total_seconds() * 1000
        context['provider'] = provider_name
        context['model'] = model_name

        # Log the interaction
        try:
            interaction_data = {
                "provider": provider_name,
                "model": model_name,
                "prompt": input_data,
                "result": result.strip(),
                "duration_ms": context['generation_time'],
                "verbose": context.get('verbose', False),
                "cached": provider.use_cache,
            }
            log_interaction("command_generation", interaction_data)
        except Exception:
            # Log failures should not interrupt the flow
            pass

        return result.strip()


class CommandParser(PipelineComponent[List[str]]):
    """Component that parses and splits raw command output."""

    def process(self, input_data: CommandOutput, context: Dict[str, Any]) -> List[str]:
        """Parse and split commands from LLM response.

        Args:
            input_data: Raw command output from LLM
            context: Additional context

        Returns:
            List of individual commands
        """
        if "Request cannot be fulfilled." in input_data:
            commands = [input_data]
            context['is_error'] = True
        else:
            # Split by newlines and filter out empty lines
            commands = [cmd.strip() for cmd in input_data.splitlines() if cmd.strip()]
            context['is_error'] = False

        context['commands'] = commands
        context['command_count'] = len(commands)
        return commands


class CommandParameterProcessor(PipelineComponent[Dict[str, ProcessedCommand]]):
    """Component that processes parameters in commands."""

    def __init__(self, provider: Optional[BaseLLMProvider] = None):
        """Initialize with an optional LLM provider.

        Args:
            provider: The LLM provider to use for parameter info
        """
        self.provider = provider

    def process(self, input_data: List[str], context: Dict[str, Any]) -> Dict[str, ProcessedCommand]:
        """Process parameters in commands.

        Args:
            input_data: List of commands to process
            context: Additional context

        Returns:
            Dictionary mapping command IDs to processed commands
        """
        # If error or no commands, return as is
        if context.get('is_error', False) or not input_data:
            return {f"command_{i+1}": cmd for i, cmd in enumerate(input_data)}

        processed_commands = {}
        parameter_values = {}

        # Get or create provider if not already initialized
        provider = self.provider
        if not provider and '[' in ''.join(input_data):
            model_type = context.get('model_type')
            api_key = context.get('api_key')
            verbose = context.get('verbose', False)
            use_cache = context.get('use_cache', True)
            provider = get_provider(model_type, api_key, verbose, validate=False, use_cache=use_cache)

        for i, command in enumerate(input_data):
            command_id = f"command_{i+1}"

            # Check if command has parameters
            if '[' in command:
                # Get parameter info for this command
                parameter_info = self._get_parameter_info(command, provider, context)

                # Parse command to extract parts
                base_command, params, bracket_params = parse_command_parameters(command)

                # If no parameters, just keep the command as is
                if not bracket_params:
                    processed_commands[command_id] = command
                    parameter_values[command_id] = {}
                    continue

                # Otherwise, store the extracted parameter info in context
                context.setdefault('parameter_info', {})[command_id] = {
                    'command': command,
                    'base_command': base_command,
                    'params': params,
                    'bracket_params': bracket_params,
                    'info': parameter_info
                }
            else:
                # Command has no parameters to process
                processed_commands[command_id] = command
                parameter_values[command_id] = {}

        context['processed_commands'] = processed_commands
        context['parameter_values'] = parameter_values

        return processed_commands

    def _get_parameter_info(self, command: str, provider: Optional[BaseLLMProvider],
                           context: Dict[str, Any]) -> Dict[str, Dict[str, Any]]:
        """Get information about parameters in a command.

        Args:
            command: The command to analyze
            provider: LLM provider to use
            context: Additional context

        Returns:
            Dictionary of parameter information
        """
        # Extract parameters that need filling in
        bracket_params = re.findall(r'\[([A-Z_]+)\]', command)

        if not bracket_params or not provider:
            return {}

        # Build a system prompt manually instead of using the template
        system_prompt = """You are InfraGPT Parameter Helper, a specialized assistant that helps users understand Google Cloud CLI command parameters.

Analyze the Google Cloud CLI command below and provide information about each parameter that needs to be filled in.
For each parameter in square brackets, provide:
1. A brief description of what this parameter is
2. Examples of valid values
3. Any constraints or requirements

Format your response as JSON with the parameter name as key."""

        # Build a user prompt manually
        user_prompt = f"Command: {command}\n\nParameter JSON:"

        # Don't show a status display here since it's shown in the calling function
        # Just generate the parameters directly
        result = provider.generate(user_prompt, system_prompt=system_prompt)

        # Extract the JSON part
        try:
            # Find JSON part between triple backticks if present
            if "```json" in result:
                json_part = result.split("```json")[1].split("```")[0].strip()
            elif "```" in result:
                json_part = result.split("```")[1].strip()
            else:
                json_part = result.strip()

            parameter_info = json.loads(json_part)
            return parameter_info
        except Exception as e:
            console.print(f"[bold yellow]Warning:[/bold yellow] Could not parse parameter info: {e}")
            return {}


class CommandPipeline:
    """Main pipeline for processing commands end-to-end."""

    def __init__(self):
        """Initialize the command pipeline with default components."""
        self.components = [
            CommandInputProcessor(),
            CommandGenerator(),
            CommandParser(),
            CommandParameterProcessor()
        ]
        self.pre_processors = []
        self.post_processors = []

    def add_pre_processor(self, processor: Callable[[Dict[str, Any]], Dict[str, Any]]):
        """Add a pre-processor to the pipeline.

        Args:
            processor: Function that processes the context before the main pipeline
        """
        self.pre_processors.append(processor)

    def add_post_processor(self, processor: Callable[[Dict[str, Any]], Dict[str, Any]]):
        """Add a post-processor to the pipeline.

        Args:
            processor: Function that processes the context after the main pipeline
        """
        self.post_processors.append(processor)

    def replace_component(self, index: int, component: PipelineComponent):
        """Replace a component in the pipeline.

        Args:
            index: Index of the component to replace
            component: New component to use
        """
        if 0 <= index < len(self.components):
            self.components[index] = component

    def process(self, input_data: str, **kwargs) -> Dict[str, Any]:
        """Process input through the entire pipeline.

        Args:
            input_data: Raw user input
            **kwargs: Additional context parameters

        Returns:
            Context dictionary with processing results
        """
        # Initialize context with kwargs
        context = dict(kwargs)

        # Run pre-processors
        for processor in self.pre_processors:
            context = processor(context)

        # Run main pipeline
        current_input = input_data
        for component in self.components:
            current_input = component.process(current_input, context)

        # Run post-processors
        for processor in self.post_processors:
            context = processor(context)

        return context


# Create a global pipeline instance
command_pipeline = CommandPipeline()

def generate_gcloud_command(prompt: str, model_type: Optional[MODEL_TYPE] = None,
                             api_key: Optional[str] = None, verbose: bool = False,
                             use_cache: bool = True) -> str:
    """Generate a gcloud command based on the user's natural language prompt.

    Args:
        prompt: The natural language prompt requesting a gcloud command
        model_type: The LLM model type to use (if None, use default from config)
        api_key: The API key for the LLM (if None, use default from config or env)
        verbose: Whether to print verbose output
        use_cache: Whether to use response caching

    Returns:
        The generated gcloud command(s) as a string
    """
    # Run the pipeline with context parameters
    context = command_pipeline.process(
        prompt,
        model_type=model_type,
        api_key=api_key,
        verbose=verbose,
        use_cache=use_cache
    )

    # Return the raw generated result
    return context.get('raw_result', '')

def get_parameter_info(command: str, model_type: Optional[MODEL_TYPE] = None,
                       api_key: Optional[str] = None, verbose: bool = False,
                       use_cache: bool = True) -> Dict[str, Dict[str, Any]]:
    """Get information about parameters from the LLM.

    Args:
        command: The command containing parameters to explain
        model_type: The LLM model type to use
        api_key: The API key for the LLM
        verbose: Whether to print verbose output
        use_cache: Whether to use response caching

    Returns:
        A dictionary mapping parameter names to their descriptions
    """
    # Create a parameter processor component
    processor = CommandParameterProcessor()

    # Create a context with the necessary parameters
    context = {
        'model_type': model_type,
        'api_key': api_key,
        'verbose': verbose,
        'use_cache': use_cache
    }

    # Process the single command
    result = processor.process([command], context)

    # Check if we have parameter info for the command
    command_id = 'command_1'
    if command_id in context.get('parameter_info', {}):
        return context['parameter_info'][command_id].get('info', {})

    return {}

def parse_command_parameters(command: str) -> Tuple[str, Dict[str, str], List[str]]:
    """Parse a command to extract its parameters and bracket placeholders.

    Args:
        command: The command to parse

    Returns:
        A tuple of (base_command, params, bracket_params)
    """
    # Extract base command and arguments
    parts = command.split()
    base_command = []

    params = {}
    current_param = None
    bracket_params = []

    for part in parts:
        # Extract parameters in square brackets (could be in any part of the command)
        bracket_matches = re.findall(r'\[([A-Z_]+)\]', part)
        if bracket_matches:
            for match in bracket_matches:
                bracket_params.append(match)

        if part.startswith('--'):
            # Handle --param=value format
            if '=' in part:
                param_name, param_value = part.split('=', 1)
                params[param_name[2:]] = param_value
            else:
                current_param = part[2:]
                params[current_param] = None
        elif current_param is not None:
            # This is a value for the previous parameter
            params[current_param] = part
            current_param = None
        else:
            # This is part of the base command
            base_command.append(part)

    return ' '.join(base_command), params, bracket_params

def split_commands(result: str) -> List[str]:
    """Split multiple commands from the response.

    Args:
        result: The LLM response containing one or more commands

    Returns:
        A list of individual commands
    """
    # Create a parser component
    parser = CommandParser()

    # Process the result
    context = {}
    commands = parser.process(result, context)

    return commands


class InteractiveParameterProcessor(PipelineComponent[Dict[str, ProcessedCommand]]):
    """Component that interactively prompts for parameters."""

    def __init__(self, provider: Optional[BaseLLMProvider] = None):
        """Initialize with an optional LLM provider.

        Args:
            provider: The LLM provider to use for parameter info
        """
        self.provider = provider
        self.parameter_processor = CommandParameterProcessor(provider)

    def process(self, input_data: List[str], context: Dict[str, Any]) -> Dict[str, ProcessedCommand]:
        """Process parameters in commands by prompting the user.

        Args:
            input_data: List of commands to process
            context: Additional context

        Returns:
            Dictionary mapping command IDs to processed commands
        """
        from ..ui.prompts import prompt_for_parameter

        # Process parameters to get info
        self.parameter_processor.process(input_data, context)

        # If error or no commands, return as is
        if context.get('is_error', False) or not input_data:
            return {f"command_{i+1}": cmd for i, cmd in enumerate(input_data)}

        processed_commands = {}
        parameter_values = {}

        # Get parameter info from context
        parameter_info = context.get('parameter_info', {})

        for i, command in enumerate(input_data):
            command_id = f"command_{i+1}"

            # Skip if no parameters or already processed
            if command_id not in parameter_info:
                processed_commands[command_id] = command
                parameter_values[command_id] = {}
                continue

            # Get command info
            cmd_info = parameter_info[command_id]
            base_command = cmd_info['base_command']
            params = cmd_info['params']
            bracket_params = cmd_info['bracket_params']
            param_info = cmd_info['info']

            if not bracket_params:
                processed_commands[command_id] = command
                parameter_values[command_id] = {}
                continue

            # Show the original command template
            console.print("\n[bold blue]Command template:[/bold blue]")
            from rich.panel import Panel
            
            console.print(Panel(command, border_style="blue"))

            # Now prompt for each parameter
            console.print("\n[bold magenta]Command requires the following parameters:[/bold magenta]")

            # Replace bracket parameters in command
            command_with_replacements = command
            collected_params = {}

            for param in bracket_params:
                # Get parameter info
                info = param_info.get(param, {})
                description = info.get('description', f"Value for {param}")
                examples = info.get('examples', [])
                default = info.get('default', None)

                # Check if parameter is required
                is_required = info.get('required', True)  # Default to required if not specified
                
                # Prompt for value
                value = prompt_for_parameter(param, description, examples, default, required=is_required)

                # Store parameter value
                collected_params[param] = value

                # Replace all occurrences of [PARAM] with the value
                command_with_replacements = command_with_replacements.replace(f"[{param}]", value)

            # Store processed command and parameters
            processed_commands[command_id] = command_with_replacements
            parameter_values[command_id] = collected_params

            # Show the final command
            console.print(Panel(command_with_replacements, border_style="green", title=f"Final Command {i+1}"))

        context['processed_commands'] = processed_commands
        context['parameter_values'] = parameter_values

        return processed_commands


def prompt_for_parameters(command: str, model_type: Optional[MODEL_TYPE] = None, return_params: bool = False) -> Union[str, Tuple[str, Dict[str, str]]]:
    """Prompt the user for each parameter in the command with AI assistance.

    Args:
        command: The command template with parameters
        model_type: The LLM model type to use
        return_params: Whether to return parameter values

    Returns:
        The processed command, or a tuple of (command, params) if return_params is True
    """
    # Create a processor pipeline for this specific task
    processor = InteractiveParameterProcessor()

    # Create context
    context = {
        'model_type': model_type,
        'verbose': False
    }

    # Process the command
    processed_commands = processor.process([command], context)

    # Get the processed command and params
    processed_command = processed_commands.get('command_1', command)
    params = context.get('parameter_values', {}).get('command_1', {})

    if return_params:
        return processed_command, params
    return processed_command
