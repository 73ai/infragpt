#!/usr/bin/env python3
"""
Docker container execution module for InfraGPT CLI agent.

This module provides isolated command execution in Docker containers with:
- Real-time streaming output
- Environment variable passthrough for cloud credentials
- Working directory tracking
- Container lifecycle management
"""

import os
from abc import ABC, abstractmethod
from typing import Tuple, Dict, Optional
from rich.console import Console

console = Console()


class DockerNotAvailableError(Exception):
    """Raised when Docker is not installed or running."""

    pass


class ExecutorInterface(ABC):
    """Abstract interface for command executors."""

    @abstractmethod
    def execute_command(self, command: str) -> Tuple[int, str, bool]:
        """
        Execute a command and return results.

        Args:
            command: Shell command to execute

        Returns:
            Tuple of (exit_code, output, was_cancelled)
        """
        pass

    @abstractmethod
    def cleanup(self) -> None:
        """Clean up resources."""
        pass


def is_sandbox_mode() -> bool:
    """
    Check if sandbox mode is enabled.

    Sandbox mode is enabled by default. Set INFRAGPT_ISOLATED=false to disable.
    """
    return os.environ.get("INFRAGPT_ISOLATED", "").lower() != "false"


def is_docker_available() -> bool:
    """Check if Docker daemon is available and running."""
    try:
        import docker
    except ImportError:
        raise DockerNotAvailableError(
            "Docker SDK not installed. Install with: uv pip install docker"
        )

    try:
        client = docker.from_env()
        client.ping()
        return True
    except docker.errors.DockerException as e:
        raise DockerNotAvailableError(f"Docker error: {e}")


# Module-level executor singleton
_executor: Optional["ContainerRunner"] = None


def get_executor() -> "ContainerRunner":
    """Get or create the ContainerRunner singleton."""
    global _executor
    if _executor is None:
        _executor = ContainerRunner()
    return _executor


def cleanup_executor() -> None:
    """Clean up the executor and remove container."""
    global _executor
    if _executor is not None:
        _executor.cleanup()
        _executor = None


class ContainerRunner(ExecutorInterface):
    """Docker container executor with streaming support."""

    def __init__(
        self,
        image: str = "infragpt/sandbox:latest",
        workdir: str = "/workspace",
        env: Optional[Dict[str, str]] = None,
        volumes: Optional[Dict[str, Dict[str, str]]] = None,
        timeout: int = 60,
    ):
        """
        Initialize container runner.

        Args:
            image: Docker image to use
            workdir: Working directory inside container
            env: Additional environment variables
            volumes: Additional volume mounts {host_path: {"bind": container_path, "mode": "rw"}}
            timeout: Command timeout in seconds
        """
        self.image = image
        self.workdir = workdir
        self.user_env = env or {}
        self.user_volumes = volumes or {}
        self.timeout = timeout

        self.client = None
        self.container = None
        self.current_cwd = workdir
        self.cancelled = False

    def start(self) -> None:
        """Create and start the container."""
        try:
            import docker
        except ImportError:
            raise DockerNotAvailableError(
                "Docker SDK not installed. Install with: pip install docker"
            )

        if not is_docker_available():
            raise DockerNotAvailableError(
                "Docker is not running. Please start Docker to use sandbox mode."
            )

        self.client = docker.from_env()

        # Check if image exists, if not pull from registry
        try:
            self.client.images.get(self.image)
        except docker.errors.ImageNotFound:
            try:
                self.client.images.pull("ghcr.io/73ai/infragpt-sandbox", tag="latest")
                self.client.images.get("ghcr.io/73ai/infragpt-sandbox:latest").tag(
                    "infragpt/sandbox", "latest"
                )
            except Exception as e:
                raise DockerNotAvailableError(
                    f"Failed to pull sandbox image: {e}\n"
                    f"Run: docker pull ghcr.io/73ai/infragpt-sandbox:latest"
                )

        # Build volume mounts
        mounts = {os.getcwd(): {"bind": "/workspace", "mode": "rw"}}
        mounts.update(self.user_volumes)

        self.container = self.client.containers.run(
            self.image,
            command="tail -f /dev/null",
            detach=True,
            tty=True,
            working_dir=self.workdir,
            volumes=mounts,
            environment=self.user_env,
            remove=True,
        )

    def execute_command(self, command: str) -> Tuple[int, str, bool]:
        """
        Execute a command in the container with streaming output.

        Args:
            command: Shell command to execute

        Returns:
            Tuple of (exit_code, output, was_cancelled)
        """
        if self.container is None:
            raise DockerNotAvailableError("Container not started. Call start() first.")

        self.cancelled = False

        console.print(f"[bold cyan]Executing:[/bold cyan] {command}")
        console.print("[dim]Press Ctrl+C to cancel...[/dim]\n")

        try:
            # Prepend cd to current working directory
            full_command = f"cd {self.current_cwd} 2>/dev/null || cd /workspace; {command}"

            # Create exec instance using low-level API for streaming
            exec_id = self.client.api.exec_create(
                container=self.container.id,
                cmd=["/bin/sh", "-c", full_command],
                tty=True,
                stdout=True,
                stderr=True,
            )

            # Stream output
            output_chunks = []
            try:
                for chunk in self.client.api.exec_start(exec_id, stream=True):
                    decoded = chunk.decode("utf-8", errors="replace")
                    output_chunks.append(decoded)
                    console.print(decoded, end="")
                    console.file.flush()
            except KeyboardInterrupt:
                self.cancelled = True
                console.print("\n[yellow]Command cancelled by user[/yellow]")
                # Try to kill the exec process
                try:
                    self.client.api.exec_start(
                        self.client.api.exec_create(
                            container=self.container.id,
                            cmd=["/bin/sh", "-c", "pkill -P 1"],
                        )
                    )
                except Exception:
                    pass

            # Get exit code
            exec_info = self.client.api.exec_inspect(exec_id)
            exit_code = exec_info.get("ExitCode", -1) if not self.cancelled else -1
            output = "".join(output_chunks)

            # Track working directory changes
            self._track_cwd(command)

            return exit_code, output, self.cancelled

        except Exception as e:
            console.print(f"[bold red]Error executing command:[/bold red] {e}")
            return -1, str(e), False

    def _track_cwd(self, command: str) -> None:
        """
        Track working directory changes from cd commands.

        Args:
            command: The command that was executed
        """
        # Get the actual current directory from the container
        try:
            exec_id = self.client.api.exec_create(
                container=self.container.id,
                cmd=[
                    "/bin/sh",
                    "-c",
                    f"cd {self.current_cwd} 2>/dev/null || cd /workspace; {command}; pwd",
                ],
                tty=False,
                stdout=True,
                stderr=False,
            )
            # Only update if command contains cd
            if "cd " in command or command.strip() == "cd":
                result = self.client.api.exec_start(exec_id, stream=False)
                new_cwd = result.decode("utf-8").strip().split("\n")[-1]
                if new_cwd and new_cwd.startswith("/"):
                    self.current_cwd = new_cwd
        except Exception:
            pass  # Ignore errors in tracking

    def stop(self) -> None:
        """Stop and remove the container."""
        if self.container is not None:
            try:
                console.print("[dim]Stopping sandbox container...[/dim]")
                self.container.stop(timeout=5)
            except Exception:
                try:
                    self.container.kill()
                except Exception:
                    pass
            self.container = None

        if self.client is not None:
            self.client.close()
            self.client = None

    def cleanup(self) -> None:
        """Alias for stop() - implements ExecutorInterface."""
        self.stop()
