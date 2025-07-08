#!/bin/bash

# Script to generate Python gRPC client from protobuf definitions
# Assumes grpcio-tools is already installed in the environment

set -e

echo "Generating Python gRPC client from proto files..."

# Check if grpc_tools is available
if ! python3 -c "import grpc_tools.protoc" 2>/dev/null; then
    echo "Error: grpcio-tools not installed. Please install it first:"
    echo "  pip install grpcio-tools"
    exit 1
fi

# Generate Python gRPC client from proto files
python3 -m grpc_tools.protoc \
    -I../../proto \
    --python_out=backendapi/generated \
    --grpc_python_out=backendapi/generated \
    ../../proto/backend.proto


sed -i '' 's/^import backend_pb2/from . import backend_pb2/' backendapi/generated/backend_pb2_grpc.py

echo "✅ Python gRPC client generated successfully!"
echo "Generated files:"
echo "  - backendapi/generated/backend_pb2.py"
echo "  - backendapi/generated/backend_pb2_grpc.py"