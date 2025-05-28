# InfraGPT Python Client

Simple Python client for communicating with the InfraGPT service. This client hides all gRPC implementation details and provides a clean interface for Python agent services.

## Installation

### Development Setup
```bash
# Clone and navigate to the client directory
cd infragptapi/client/python

# Run development setup (creates venv, installs deps)
./setup_dev.sh

# Activate the virtual environment
source venv/bin/activate

# Generate gRPC client files
./generate.sh
```

### Production Installation
```bash
pip install ./infragptapi/client/python
```

## Usage

### Basic Usage
```python
from infragptapi import InfraGPTClient

# Create client
client = InfraGPTClient(host="localhost", port=9090)

# Send a reply to a conversation
success = client.send_reply(
    conversation_id="550e8400-e29b-41d4-a716-446655440000",
    message="Hello from Python agent!"
)

if success:
    print("Reply sent successfully!")

# Clean up
client.close()
```

### Context Manager Usage
```python
from infragptapi import InfraGPTClient

with InfraGPTClient(host="localhost", port=9090) as client:
    client.send_reply(
        conversation_id="550e8400-e29b-41d4-a716-446655440000", 
        message="Another reply!"
    )
# Automatically closes connection
```

### Error Handling
```python
from infragptapi import InfraGPTClient, InfraGPTError, ConnectionError, RequestError

client = InfraGPTClient()

try:
    client.send_reply(conversation_id, message)
except ConnectionError:
    print("Failed to connect to InfraGPT service")
except RequestError as e:
    print(f"Service returned error: {e}")
except InfraGPTError as e:
    print(f"General error: {e}")
```

## Configuration

### Default Settings
- **Host**: `localhost`
- **Port**: `9090` (gRPC port)

### Environment Variables
The client uses the default values above, but you can override them when creating the client:

```python
client = InfraGPTClient(host="infragpt-service", port=9090)
```

## Development

### Regenerating gRPC Files
If the protobuf definitions change, regenerate the client:

```bash
# Activate virtual environment
source venv/bin/activate

# Regenerate gRPC files
./generate.sh
```

### Package Structure
```
infragptapi/
├── __init__.py              # Package exports
├── client.py                # Main client class
├── exceptions.py            # Custom exceptions
└── generated/               # Auto-generated gRPC files
    ├── infragpt_pb2.py      # Protobuf messages
    └── infragpt_pb2_grpc.py # gRPC client stub
```

## Requirements

- Python 3.8+
- grpcio >= 1.50.0
- grpcio-tools >= 1.50.0

## License

Part of the InfraGPT project.