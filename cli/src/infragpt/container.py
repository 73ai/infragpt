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
import platform
from abc import ABC, abstractmethod
from pathlib import Path
from typing import Tuple, Dict, Optional
from rich.console import Console

from infragpt.api_client import GKEClusterInfo

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


def cleanup_old_containers() -> int:
    """Remove any existing sandbox containers from previous CLI sessions."""
    try:
        import docker
    except ImportError:
        return 0

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
        client.close()
        return removed
    except Exception:
        return 0


def get_executor(
    gcp_credentials_path: Optional[Path] = None,
    gke_cluster_info: Optional[GKEClusterInfo] = None,
) -> "ContainerRunner":
    """Get or create the ContainerRunner singleton."""
    global _executor
    if _executor is None:
        _executor = ContainerRunner(
            gcp_credentials_path=gcp_credentials_path,
            gke_cluster_info=gke_cluster_info,
        )
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
        gcp_credentials_path: Optional[Path] = None,
        gke_cluster_info: Optional[GKEClusterInfo] = None,
    ):
        """
        Initialize container runner.

        Args:
            image: Docker image to use
            workdir: Working directory inside container
            env: Additional environment variables
            volumes: Additional volume mounts {host_path: {"bind": container_path, "mode": "rw"}}
            timeout: Command timeout in seconds
            gcp_credentials_path: Path to GCP service account JSON file to mount
            gke_cluster_info: GKE cluster info for kubectl configuration
        """
        self.image = image or get_sandbox_image()
        self.workdir = workdir
        self.user_env = env or {}
        self.user_volumes = volumes or {}
        self.timeout = timeout
        self.gcp_credentials_path = gcp_credentials_path
        self.gke_cluster_info = gke_cluster_info

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
                self.client.images.pull(self.image)
            except Exception as e:
                raise DockerNotAvailableError(
                    f"Failed to pull sandbox image: {e}\nRun: docker pull {self.image}"
                )

        # Build volume mounts
        mounts = {os.getcwd(): {"bind": "/workspace", "mode": "rw"}}
        mounts.update(self.user_volumes)

        # Build environment variables
        env = dict(self.user_env)

        # Mount GCP credentials if available
        if self.gcp_credentials_path and self.gcp_credentials_path.exists():
            mounts[str(self.gcp_credentials_path)] = {
                "bind": "/credentials/gcp_sa.json",
                "mode": "ro",
            }
            env["GOOGLE_APPLICATION_CREDENTIALS"] = "/credentials/gcp_sa.json"

        self.container = self.client.containers.run(
            self.image,
            command="tail -f /dev/null",
            detach=True,
            tty=True,
            working_dir=self.workdir,
            volumes=mounts,
            environment=env,
            remove=True,
        )

        # Configure GCP tools if credentials are mounted
        if self.gcp_credentials_path and self.gcp_credentials_path.exists():
            self._configure_gcp_tools()

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
            full_command = (
                f"cd {self.current_cwd} 2>/dev/null || cd /workspace; {command}"
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

    def _configure_gcp_tools(self) -> None:
        """Configure gcloud and kubectl with injected credentials."""
        if self.container is None:
            return

        commands = []

        # Activate service account
        commands.append(
            "gcloud auth activate-service-account --key-file=/credentials/gcp_sa.json 2>/dev/null"
        )

        # Set project if we have GKE cluster info
        if self.gke_cluster_info:
            commands.append(
                f"gcloud config set project {self.gke_cluster_info.project_id} 2>/dev/null"
            )

            # Configure kubectl for GKE cluster
            if self.gke_cluster_info.zone:
                commands.append(
                    f"gcloud container clusters get-credentials {self.gke_cluster_info.cluster_name} "
                    f"--zone {self.gke_cluster_info.zone} "
                    f"--project {self.gke_cluster_info.project_id} 2>/dev/null"
                )
            elif self.gke_cluster_info.region:
                commands.append(
                    f"gcloud container clusters get-credentials {self.gke_cluster_info.cluster_name} "
                    f"--region {self.gke_cluster_info.region} "
                    f"--project {self.gke_cluster_info.project_id} 2>/dev/null"
                )

        # Execute all commands
        full_command = " && ".join(commands)
        try:
            exec_id = self.client.api.exec_create(
                container=self.container.id,
                cmd=["/bin/sh", "-c", full_command],
                tty=False,
                stdout=True,
                stderr=True,
            )
            self.client.api.exec_start(exec_id, stream=False)
        except Exception:
            pass  # Best effort configuration

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
