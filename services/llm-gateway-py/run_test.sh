#!/bin/bash

# Test runner script for LLM Gateway Python Service

cd "$(dirname "$0")"

echo "🧪 Running LLM Gateway Python Service Tests..."
echo "================================================="

# Activate virtual environment if it exists
if [ -d "venv" ]; then
    echo "📦 Activating virtual environment..."
    source venv/bin/activate
    echo "✅ Virtual environment activated: $(which python)"
else
    echo "⚠️  No virtual environment found, using system Python: $(which python)"
fi

# Run the test
echo ""
echo "🚀 Starting tests..."

# Install critical missing dependencies
echo "📦 Installing critical dependencies..."
pip install PyJWT fastapi uvicorn >/dev/null 2>&1 || echo "Note: Some optional dependencies couldn't be installed"

echo ""
echo "🧪 Running basic tests..."
python test_simple.py

echo ""
echo "🧪 Running advanced service logic tests..."
python test_service_logic.py

# Capture exit code
exit_code=$?

# Deactivate virtual environment if we activated it
if [ -d "venv" ]; then
    deactivate 2>/dev/null || true
fi

exit $exit_code
