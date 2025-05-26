#!/bin/bash
# Start script for LLM Gateway Python service

set -e

echo "🚀 Starting LLM Gateway Python Service..."

# Get the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "❌ .env file not found!"
    echo "📝 Please copy .env.example to .env and configure your API keys:"
    echo "   cp .env.example .env"
    echo "   # Then edit .env with your actual API keys"
    exit 1
fi

# Check for required API keys
echo "🔍 Checking API key configuration..."
source .env

if [[ -z "$OPENAI_API_KEY" || "$OPENAI_API_KEY" == "sk-your-openai-api-key-here" ]]; then
    echo "❌ OPENAI_API_KEY not configured in .env file"
    echo "💡 Get your API key from: https://platform.openai.com/api-keys"
    exit 1
fi

if [[ -z "$PINECONE_API_KEY" || "$PINECONE_API_KEY" == "your-pinecone-api-key-here" ]]; then
    echo "❌ PINECONE_API_KEY not configured in .env file" 
    echo "💡 Get your API key from: https://www.pinecone.io/"
    exit 1
fi

if [[ -z "$TAVILY_API_KEY" || "$TAVILY_API_KEY" == "your-tavily-api-key-here" ]]; then
    echo "❌ TAVILY_API_KEY not configured in .env file"
    echo "💡 Get your API key from: https://tavily.com/"
    exit 1
fi

echo "✅ API keys configured"

# Check if virtual environment exists
if [ ! -d "venv" ]; then
    echo "🔧 Creating virtual environment..."
    python -m venv venv --upgrade-deps
fi

# Activate virtual environment
echo "📦 Activating virtual environment..."
source venv/bin/activate

# Check if proto files exist
if [ ! -f "llm/v1/llm_pb2.py" ]; then
    echo "🔧 Generating proto files..."
    ./generate_proto.sh
fi

# Install/upgrade dependencies if needed
echo "📦 Checking dependencies..."
pip install --quiet --upgrade -r requirements.txt

# Run a quick health check
echo "🩺 Running health check..."
if ! python -c "from services.llm_service import LLMServicer; print('✅ Service imports working')"; then
    echo "❌ Service health check failed"
    exit 1
fi

# Start the service
echo "🎯 Starting gRPC server on port ${GRPC_PORT:-50054}..."
echo "🌐 Admin API will be available on port ${HTTP_PORT:-8091}"
echo ""
echo "🔗 Service endpoints:"
echo "   - gRPC: localhost:${GRPC_PORT:-50054}"
echo "   - Admin API: http://localhost:${HTTP_PORT:-8091}"
echo "   - Health check: http://localhost:${HTTP_PORT:-8091}/health"
echo ""
echo "📝 Logs will appear below. Press Ctrl+C to stop the service."
echo "=" * 70

exec python main.py
