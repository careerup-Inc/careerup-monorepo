#!/bin/bash

# Production startup script for LLM Gateway Python service
set -e

echo "🚀 Starting LLM Gateway Python service in production mode..."

# Check if running from correct directory
if [ ! -f "main.py" ]; then
    echo "❌ Error: Please run this script from the llm-gateway-py directory"
    exit 1
fi

# Set production environment variables
export ENVIRONMENT=production
export DEBUG=false
export LOG_LEVEL=INFO
export SERVICE_NAME=llm-gateway-py

# Default ports
export GRPC_PORT=${GRPC_PORT:-50054}
export HTTP_PORT=${HTTP_PORT:-8091}
export MAX_WORKERS=${MAX_WORKERS:-20}

# Admin API configuration
export ENABLE_ADMIN_API=${ENABLE_ADMIN_API:-true}

# Security: Ensure admin API key is set in production
if [ "$ENABLE_ADMIN_API" = "true" ] && [ -z "$ADMIN_API_KEY" ]; then
    echo "❌ Error: ADMIN_API_KEY must be set in production mode"
    exit 1
fi

# Check for required environment variables
check_env_var() {
    if [ -z "${!1}" ]; then
        echo "❌ Error: Environment variable $1 is not set"
        return 1
    fi
}

echo "🔍 Checking required environment variables..."

# Check required API keys
if ! check_env_var "OPENAI_API_KEY"; then
    echo "   OpenAI API key is required for production"
    exit 1
fi

if ! check_env_var "PINECONE_API_KEY"; then
    echo "   Pinecone API key is required for production"
    exit 1
fi

if ! check_env_var "TAVILY_API_KEY"; then
    echo "   Tavily API key is required for production"
    exit 1
fi

# Validate Python environment
if ! command -v python3 &> /dev/null; then
    echo "❌ Error: Python 3 is not installed"
    exit 1
fi

PYTHON_VERSION=$(python3 --version 2>&1 | cut -d' ' -f2)
if ! python3 -c "import sys; exit(0 if sys.version_info >= (3, 11) else 1)" 2>/dev/null; then
    echo "❌ Error: Python 3.11+ required. Found: $PYTHON_VERSION"
    exit 1
fi

echo "✅ Python version: $PYTHON_VERSION"

# Check dependencies
echo "📦 Checking Python dependencies..."
if [ -f "requirements.txt" ]; then
    pip install -q --no-deps -r requirements.txt
    echo "✅ Dependencies verified"
else
    echo "❌ Error: requirements.txt not found"
    exit 1
fi

# Set Python path
export PYTHONPATH="../../../proto:$(pwd):$PYTHONPATH"

# Create necessary directories
mkdir -p logs
mkdir -p /tmp/llm-gateway-py

# Set up logging
export LOG_FILE="logs/llm-gateway-$(date +%Y%m%d).log"

# Display configuration
echo ""
echo "📋 Production Configuration:"
echo "   Environment: $ENVIRONMENT"
echo "   Log Level: $LOG_LEVEL"
echo "   gRPC Port: $GRPC_PORT"
echo "   HTTP Port: $HTTP_PORT"
echo "   Max Workers: $MAX_WORKERS"
echo "   Admin API: $ENABLE_ADMIN_API"
echo "   Log File: $LOG_FILE"
echo ""

# Create systemd-style process management
create_pid_file() {
    echo $$ > /tmp/llm-gateway-py/service.pid
    echo "📝 PID file created: /tmp/llm-gateway-py/service.pid"
}

cleanup() {
    echo ""
    echo "🛑 Shutting down LLM Gateway Python service..."
    rm -f /tmp/llm-gateway-py/service.pid
    echo "✅ Cleanup completed"
    exit 0
}

# Set up signal handlers
trap cleanup SIGTERM SIGINT

# Create PID file
create_pid_file

echo "🎯 Starting LLM Gateway Python service in production mode..."
echo "   Process ID: $$"
echo "   gRPC server: localhost:$GRPC_PORT"

if [ "$ENABLE_ADMIN_API" = "true" ]; then
    echo "   Admin API: http://localhost:$HTTP_PORT/admin/"
fi

echo ""
echo "📊 Monitor service health:"
echo "   Health check: curl http://localhost:$HTTP_PORT/health"

if [ "$ENABLE_ADMIN_API" = "true" ]; then
    echo "   Metrics: curl -H 'Authorization: Bearer \$ADMIN_API_KEY' http://localhost:$HTTP_PORT/admin/metrics"
fi

echo ""
echo "🛑 To stop the service: kill -TERM $$"
echo ""

# Start the service with production settings
exec python main.py 2>&1 | tee -a "$LOG_FILE"
