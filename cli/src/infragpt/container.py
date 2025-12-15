#!/usr/bin/env python3
"""
Docker container execution module for InfraGPT CLI agent.

This module provides isolated command execution in Docker containers with:
- Real-time streaming output
- Working directory tracking
- Container lifecycle management
"""

import os
import platform
import shlex
from abc import ABC, abstractmethod
from typing import Tuple, Dict, Optional

import docker
from rich.console import Console

console = Console()


def get_sandbox_image() -> str:
    """Get the full sandbox image name for current platform."""
    machine = platform.machine().lower()
    if machine in ("arm64", "aarch64"):
        arch = "arm64"
    else:
        arch = "amd64"
    return f"ghcr.io/73ai/infragpt-sandbox:latest-{arch}"


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


def ensure_docker_available() -> None:
    """Ensure Docker daemon is available and running. Raises DockerNotAvailableError if not."""
    try:
        client = docker.from_env()
        client.ping()
        client.close()
    except docker.errors.DockerException as e:
        raise DockerNotAvailableError(f"Docker error: {e}")


# Module-level executor singleton
_executor: Optional["ContainerRunner"] = None


def cleanup_old_containers() -> int:
    """Remove any existing sandbox containers from previous CLI sessions."""
    client = None
    try:
        client = docker.from_env()
        image_prefix = "ghcr.io/73ai/infragpt-sandbox:"
        containers = client.containers.list(all=True)
        removed = 0
        for container in containers:
            if container.image.tags and any(
                tag.startswith(image_prefix) for tag in container.image.tags
            ):
                try:
                    container.stop(timeout=5)
                except Exception:
                    pass
                try:
                    container.remove(force=True)
                except Exception:
                    pass
                removed += 1
        return removed
    except Exception:
        return 0
    finally:
        if client is not None:
            client.close()


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
        image: Optional[str] = None,
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
        self.image = image or get_sandbox_image()
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
        ensure_docker_available()

        self.client = docker.from_env()

        # Check if image exists, if not pull from registry
        try:
            self.client.images.get(self.image)
        except docker.errors.ImageNotFound:
            try:
                self.client.images.pull(self.image)
            except Exception as e:
                raise DockerNotAvailableError(
                    f"Failed to pull sandbox image: {e}\n"
                    f"Run: docker pull {self.image}"
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
            # Prepend cd to current working directory, append pwd capture for tracking
            cwd_marker = "__INFRAGPT_CWD__"
            # Use timeout command if timeout is set (124 is timeout's exit code)
            timeout_prefix = f"timeout {self.timeout} " if self.timeout > 0 else ""
            full_command = (
                f"cd {shlex.quote(self.current_cwd)} 2>/dev/null || cd /workspace; "
                f"{timeout_prefix}{command}; _exit_code=$?; echo {cwd_marker}; pwd; exit $_exit_code"
            )

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
                    # Don't print the cwd marker and path
                    if cwd_marker not in decoded:
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

            # Extract and update working directory from output
            self._update_cwd_from_output(output, cwd_marker)

            return exit_code, output, self.cancelled

        except Exception as e:
            console.print(f"[bold red]Error executing command:[/bold red] {e}")
            return -1, str(e), False

    def _update_cwd_from_output(self, output: str, marker: str) -> None:
        """
        Extract working directory from command output using the marker.

        Args:
            output: Full command output containing the marker and pwd
            marker: The marker string used to identify pwd output
        """
        try:
            if marker in output:
                # Find the line after the marker
                lines = output.split(marker)
                if len(lines) > 1:
                    pwd_output = lines[-1].strip().split("\n")[0].strip()
                    if pwd_output and pwd_output.startswith("/"):
                        self.current_cwd = pwd_output
        except Exception:
            pass

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
