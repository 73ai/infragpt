#!/bin/bash

# Script to generate protobuf code for both Python and Go from agent.proto

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROTO_FILE="$SCRIPT_DIR/agent.proto"

echo "Generating protobuf code from $PROTO_FILE..."

# Generate Python protobuf code
echo "Generating Python protobuf code..."
python -m grpc_tools.protoc \
    --python_out="$SCRIPT_DIR" \
    --grpc_python_out="$SCRIPT_DIR" \
    --proto_path="$SCRIPT_DIR" \
    "$PROTO_FILE"

# Fix imports in the generated gRPC file
echo "Fixing import paths in generated files..."
if [ -f "$SCRIPT_DIR/agent_pb2_grpc.py" ]; then
    sed -i.bak 's/import agent_pb2 as agent__pb2/from . import agent_pb2 as agent__pb2/' "$SCRIPT_DIR/agent_pb2_grpc.py"
fi

echo "✓ Python protobuf code generated"

# Generate Go protobuf code for the client
GO_CLIENT_DIR="$SCRIPT_DIR/../client/go/proto"
echo "Generating Go protobuf code in $GO_CLIENT_DIR..."

# Generate Go messages
protoc \
    --go_out="$GO_CLIENT_DIR" \
    --go_opt=paths=source_relative \
    --proto_path="$SCRIPT_DIR" \
    "$PROTO_FILE"

echo "✓ Go protobuf messages generated"

# Try to generate Go gRPC services
if command -v protoc-gen-go-grpc >/dev/null 2>&1; then
    echo "Generating Go gRPC services..."
    
    # Add Go bin to PATH if it exists
    if [ -d "$HOME/go/bin" ]; then
        export PATH="$HOME/go/bin:$PATH"
    fi
    
    protoc \
        --go-grpc_out="$GO_CLIENT_DIR" \
        --go-grpc_opt=paths=source_relative \
        --proto_path="$SCRIPT_DIR" \
        "$PROTO_FILE"
    
    echo "✓ Go gRPC services generated"
else
    echo "⚠️  protoc-gen-go-grpc not found, skipping gRPC service generation"
    echo "   Install with: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
fi

echo ""
echo "🎉 Protobuf generation complete!"
echo ""
echo "Generated files:"
echo "  Python: $SCRIPT_DIR/agent_pb2.py"
echo "  Python: $SCRIPT_DIR/agent_pb2_grpc.py"
echo "  Go:     $GO_CLIENT_DIR/agent.pb.go"
if command -v protoc-gen-go-grpc >/dev/null 2>&1; then
    echo "  Go:     $GO_CLIENT_DIR/agent_grpc.pb.go"
fi