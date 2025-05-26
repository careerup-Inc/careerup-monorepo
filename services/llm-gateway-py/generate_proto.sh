#!/bin/bash

# Generate Python gRPC files from proto definitions

set -e

echo "🔧 Generating Python gRPC files from proto definitions..."

# Get the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PROTO_DIR="$PROJECT_ROOT/proto"
OUTPUT_DIR="$SCRIPT_DIR"

# Ensure we're in the right directory
cd "$SCRIPT_DIR"

# Activate virtual environment if it exists
if [ -d "venv" ]; then
    echo "📦 Activating virtual environment..."
    source venv/bin/activate
fi

# Install grpcio-tools if not already installed
echo "📦 Installing grpcio-tools..."
pip install grpcio-tools >/dev/null 2>&1 || true

# Generate Python files
echo "🏗️  Generating Python gRPC files..."

# Create output directory for generated files
mkdir -p "$OUTPUT_DIR/llm/v1"

# Generate the Python files
python -m grpc_tools.protoc \
    --proto_path="$PROTO_DIR" \
    --python_out="$OUTPUT_DIR" \
    --grpc_python_out="$OUTPUT_DIR" \
    llm/v1/llm.proto

# Create __init__.py files
touch "$OUTPUT_DIR/llm/__init__.py"
touch "$OUTPUT_DIR/llm/v1/__init__.py"

echo "✅ Python gRPC files generated successfully!"
echo "📁 Generated files:"
find "$OUTPUT_DIR/llm" -name "*.py" | sort

echo ""
echo "🎉 Ready to run the Python LLM Gateway service!"
echo "💡 Next step: python main.py"
