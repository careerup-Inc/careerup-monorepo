#!/bin/bash

# Development startup script for LLM Gateway Python service
set -e

echo "🚀 Starting LLM Gateway Python service in development mode..."

# Check if running from correct directory
if [ ! -f "main.py" ]; then
    echo "❌ Error: Please run this script from the llm-gateway-py directory"
    exit 1
fi

# Set development environment variables
export ENVIRONMENT=development
export DEBUG=true
export LOG_LEVEL=DEBUG
export SERVICE_NAME=llm-gateway-py

# Default ports (can be overridden)
export GRPC_PORT=${GRPC_PORT:-50054}
export HTTP_PORT=${HTTP_PORT:-8091}
export MAX_WORKERS=${MAX_WORKERS:-5}

# Admin API configuration
export ENABLE_ADMIN_API=true
export ADMIN_API_KEY=${ADMIN_API_KEY:-dev-admin-key}

# Check for required environment variables
check_env_var() {
    if [ -z "${!1}" ]; then
        echo "❌ Error: Environment variable $1 is not set"
        echo "   Please set it in your .env file or environment"
        return 1
    fi
}

echo "🔍 Checking required environment variables..."

# Load .env file if it exists
if [ -f ".env" ]; then
    echo "📁 Loading environment from .env file..."
    export $(grep -v '^#' .env | xargs)
else
    echo "⚠️  No .env file found. Make sure to set environment variables manually."
fi

# Check required API keys
if ! check_env_var "OPENAI_API_KEY"; then
    echo "   Get your API key from: https://platform.openai.com/api-keys"
    exit 1
fi

if ! check_env_var "PINECONE_API_KEY"; then
    echo "   Get your API key from: https://app.pinecone.io/"
    exit 1
fi

if ! check_env_var "TAVILY_API_KEY"; then
    echo "   Get your API key from: https://tavily.com/"
    exit 1
fi

# Check if Python virtual environment should be used
if [ -d "venv" ]; then
    echo "🐍 Activating Python virtual environment..."
    source venv/bin/activate
fi

# Check Python version
PYTHON_VERSION=$(python3 --version 2>&1 | cut -d' ' -f2)
REQUIRED_VERSION="3.11"

if ! python3 -c "import sys; exit(0 if sys.version_info >= (3, 11) else 1)" 2>/dev/null; then
    echo "❌ Error: Python 3.11+ required. Found: $PYTHON_VERSION"
    exit 1
fi

echo "✅ Python version: $PYTHON_VERSION"

# Install/check dependencies
echo "📦 Checking Python dependencies..."
if [ -f "requirements.txt" ]; then
    pip install -q -r requirements.txt
    echo "✅ Dependencies installed"
else
    echo "❌ Error: requirements.txt not found"
    exit 1
fi

# Set Python path for proto imports
export PYTHONPATH="../../../proto:$(pwd):$PYTHONPATH"

# Create logs directory
mkdir -p logs

# Check if ports are available
check_port() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null; then
        echo "⚠️  Warning: Port $1 is already in use"
        return 1
    fi
    return 0
}

echo "🔌 Checking port availability..."
if ! check_port $GRPC_PORT; then
    echo "   gRPC port $GRPC_PORT is in use. Set GRPC_PORT to use a different port."
fi

if ! check_port $HTTP_PORT; then
    echo "   HTTP port $HTTP_PORT is in use. Set HTTP_PORT to use a different port."
fi

# Display configuration
echo ""
echo "📋 Service Configuration:"
echo "   Environment: $ENVIRONMENT"
echo "   Debug Mode: $DEBUG"
echo "   Log Level: $LOG_LEVEL"
echo "   gRPC Port: $GRPC_PORT"
echo "   HTTP Port: $HTTP_PORT"
echo "   Max Workers: $MAX_WORKERS"
echo "   Admin API: $ENABLE_ADMIN_API"
echo ""

# Optional: Run tests before starting
if [ "$1" = "--test" ]; then
    echo "🧪 Running tests first..."
    if [ -d "tests" ]; then
        python -m pytest tests/ -v
        if [ $? -ne 0 ]; then
            echo "❌ Tests failed. Aborting startup."
            exit 1
        fi
        echo "✅ All tests passed"
    else
        echo "⚠️  No tests directory found, skipping tests"
    fi
fi

echo "🎯 Starting LLM Gateway Python service..."
echo "   gRPC server will be available at: localhost:$GRPC_PORT"
echo "   Admin API will be available at: http://localhost:$HTTP_PORT/admin/docs"
echo ""
echo "📝 Logs will be written to: logs/llm-gateway.log"
echo ""
echo "🛑 Press Ctrl+C to stop the service"
echo ""

# Start the service
exec python main.py
