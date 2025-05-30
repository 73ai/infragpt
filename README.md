# InfraGPT - AI-Powered Infrastructure Management Platform 🤖

InfraGPT is a multi-service platform that provides AI-powered infrastructure management through Slack integration. The system consists of multiple services that work together to deliver intelligent DevOps workflows.

![PyPI](https://img.shields.io/pypi/v/infragpt)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/priyanshujain/infragpt/deploy.yml)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/priyanshujain/infragpt/publish.yml)

## Platform Architecture

InfraGPT consists of four main services:

### 1. 🖥️ CLI Tool (`/cli/`)
**Language**: Python  
**Purpose**: Interactive terminal interface for infrastructure command generation

- Natural language to Google Cloud commands conversion
- Interactive mode with command history
- Support for OpenAI GPT-4o and Anthropic Claude models
- Install with: `pipx install infragpt`

[**📖 CLI Documentation**](cli/README.md)

### 2. 🤖 Agent Service (`/services/agent/`)
**Language**: Python  
**Purpose**: AI-powered message processing and response generation

- Multi-agent framework with LLM integration
- Conversation management and RCA analysis
- FastAPI + gRPC dual server architecture
- Integration with InfraGPT core service

### 3. 🌐 Core Service (`/services/infragpt/`)
**Language**: Go  
**Purpose**: Main Slack bot and infrastructure management service

- Slack Socket Mode integration
- PostgreSQL-backed persistence
- GitHub PR management
- Terraform code generation
- Clean architecture with domain/infrastructure layers

### 4. 🖼️ Web Application (`/services/app/`)
**Language**: TypeScript/React  
**Purpose**: Web client interface for InfraGPT platform

- Modern React with Vite and TypeScript
- Radix UI components with Tailwind CSS
- Authentication via Clerk
- Real-time integration with platform services

## Quick Start

### CLI Tool (Most Common Entry Point)

```bash
# Install using pipx (recommended)
pipx install infragpt

# Launch interactive mode
infragpt

# Example usage
> create a new VM instance called test-vm in us-central1 with 2 CPUs
```

### Full Platform Development

```bash
# Clone the repository
git clone https://github.com/priyanshujain/infragpt.git
cd infragpt

# Run individual services (see service-specific READMEs)
# - CLI: see cli/README.md
# - Agent: see services/agent/README.md  
# - Core: see services/infragpt/README.md
# - Web: see services/app/README.md
```

## Integration Flow

The services work together in this message flow:
1. User posts in Slack channel or uses CLI
2. InfraGPT Core Service receives requests via Socket Mode
3. Core Service calls Agent Service via gRPC for AI processing
4. Agent Service processes with LLM intelligence
5. Responses flow back through the system to Slack or CLI

## Features

- **🗣️ Natural Language Processing**: Convert natural language to infrastructure commands
- **🔗 Slack Integration**: Seamless Slack bot for team collaboration
- **🧠 Multi-Agent AI**: Intelligent routing and specialized agent responses
- **📊 Web Dashboard**: Modern web interface for platform management
- **🏗️ Infrastructure as Code**: Generate Terraform and other IaC
- **📈 Analytics**: Track usage and infrastructure changes
- **🔐 Enterprise Security**: Authentication, authorization, and audit trails

## Service Documentation

Each service has its own detailed documentation:

- **[CLI Documentation](cli/README.md)** - Terminal interface, installation, and usage
- **[Agent Service](services/agent/README.md)** - AI processing and multi-agent framework
- **[Core Service](services/infragpt/README.md)** - Slack integration and main API
- **[Web Application](services/app/README.md)** - React frontend and UI components

## Development

Each service can be developed independently with its own toolchain:

- **CLI**: Python with uv package manager
- **Agent**: Python with FastAPI and gRPC
- **Core**: Go with clean architecture patterns
- **Web**: TypeScript/React with Vite

See individual service READMEs for specific development setup instructions.

## Contributing

For information on how to contribute to InfraGPT, including development setup, release process, and CI/CD configuration, please see the [CONTRIBUTING.md](CONTRIBUTING.md) file.

## License

This project is licensed under the GPL-3.0 License - see the [LICENSE](LICENSE) file for details.