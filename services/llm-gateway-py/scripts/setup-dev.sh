#!/bin/bash

# Setup script for LLM Gateway Python development environment
set -e

echo "🛠️  Setting up LLM Gateway Python development environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if running from correct directory
if [ ! -f "requirements.txt" ]; then
    echo -e "${RED}❌ Error: Please run this script from the llm-gateway-py directory${NC}"
    exit 1
fi

# Check Python version
echo -e "${BLUE}🐍 Checking Python installation...${NC}"

if ! command -v python3 &> /dev/null; then
    echo -e "${RED}❌ Python 3 is not installed${NC}"
    echo "Please install Python 3.11+ from https://python.org"
    exit 1
fi

PYTHON_VERSION=$(python3 --version 2>&1 | cut -d' ' -f2)
if ! python3 -c "import sys; exit(0 if sys.version_info >= (3, 11) else 1)" 2>/dev/null; then
    echo -e "${RED}❌ Python 3.11+ required. Found: $PYTHON_VERSION${NC}"
    echo "Please upgrade Python to 3.11 or later"
    exit 1
fi

echo -e "${GREEN}✅ Python version: $PYTHON_VERSION${NC}"

# Create virtual environment
echo -e "${BLUE}📦 Setting up Python virtual environment...${NC}"

if [ ! -d "venv" ]; then
    python3 -m venv venv
    echo -e "${GREEN}✅ Virtual environment created${NC}"
else
    echo -e "${YELLOW}⚠️  Virtual environment already exists${NC}"
fi

# Activate virtual environment
source venv/bin/activate
echo -e "${GREEN}✅ Virtual environment activated${NC}"

# Upgrade pip
echo -e "${BLUE}🔧 Upgrading pip...${NC}"
pip install --upgrade pip

# Install dependencies
echo -e "${BLUE}📚 Installing Python dependencies...${NC}"
pip install -r requirements.txt
echo -e "${GREEN}✅ Dependencies installed${NC}"

# Install development dependencies
echo -e "${BLUE}🧪 Installing development dependencies...${NC}"
pip install pytest pytest-asyncio pytest-cov black flake8 mypy types-requests
echo -e "${GREEN}✅ Development dependencies installed${NC}"

# Create .env file if it doesn't exist
echo -e "${BLUE}⚙️  Setting up environment configuration...${NC}"

if [ ! -f ".env" ]; then
    cp .env.example .env
    echo -e "${GREEN}✅ Created .env file from template${NC}"
    echo -e "${YELLOW}⚠️  Please edit .env file and add your API keys:${NC}"
    echo "   - OPENAI_API_KEY"
    echo "   - PINECONE_API_KEY"
    echo "   - TAVILY_API_KEY"
else
    echo -e "${YELLOW}⚠️  .env file already exists${NC}"
fi

# Create logs directory
echo -e "${BLUE}📁 Creating directories...${NC}"
mkdir -p logs
mkdir -p tests
echo -e "${GREEN}✅ Directories created${NC}"

# Make scripts executable
echo -e "${BLUE}🔐 Setting script permissions...${NC}"
chmod +x scripts/*.sh
echo -e "${GREEN}✅ Scripts made executable${NC}"

# Check proto files
echo -e "${BLUE}🔍 Checking proto files...${NC}"
if [ -d "../../../proto/llm/v1" ]; then
    echo -e "${GREEN}✅ Proto files found${NC}"
else
    echo -e "${YELLOW}⚠️  Proto files not found at ../../../proto/llm/v1${NC}"
    echo "Make sure you're running this from the correct directory"
fi

# Generate Python proto files if protoc is available
if command -v protoc &> /dev/null; then
    echo -e "${BLUE}🛠️  Generating Python proto files...${NC}"
    
    # Create proto output directory
    mkdir -p proto_gen
    
    # Generate Python files
    protoc --python_out=proto_gen \
           --grpc_python_out=proto_gen \
           --proto_path=../../../proto \
           ../../../proto/llm/v1/*.proto
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Proto files generated${NC}"
    else
        echo -e "${YELLOW}⚠️  Proto generation failed (this is optional)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  protoc not found, skipping proto generation${NC}"
    echo "Install protoc if you need to regenerate proto files"
fi

# Create basic test file
if [ ! -f "tests/test_basic.py" ]; then
    echo -e "${BLUE}🧪 Creating basic test file...${NC}"
    
cat > tests/test_basic.py << 'EOF'
"""Basic tests for LLM Gateway Python service."""

import pytest
from unittest.mock import Mock
import sys
import os

# Add parent directory to path for imports
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

def test_import_config():
    """Test that we can import configuration."""
    from config.settings import get_settings
    settings = get_settings()
    assert settings.service_name == "llm-gateway-py"

def test_import_vietnamese_utils():
    """Test that we can import Vietnamese utilities."""
    from utils.vietnamese import is_vietnamese_text, normalize_vietnamese_query
    
    assert is_vietnamese_text("Xin chào") == True
    assert is_vietnamese_text("Hello") == False
    
    normalized = normalize_vietnamese_query("  Xin   chào  ")
    assert normalized == "Xin chào"

def test_import_logger():
    """Test that we can import logging utilities."""
    from utils.logger import setup_logger, get_logger
    
    logger = setup_logger("test")
    assert logger is not None
    
    logger2 = get_logger("test")
    assert logger2 is not None

def test_import_metrics():
    """Test that we can import metrics utilities."""
    from utils.metrics import MetricsCollector
    
    collector = MetricsCollector()
    assert collector is not None
    
    stats = collector.get_current_stats()
    assert 'total_requests' in stats

if __name__ == "__main__":
    pytest.main([__file__])
EOF

    echo -e "${GREEN}✅ Basic test file created${NC}"
fi

# Create __init__.py files if missing
echo -e "${BLUE}📝 Creating __init__.py files...${NC}"
touch tests/__init__.py
echo -e "${GREEN}✅ __init__.py files created${NC}"

# Run basic tests to verify setup
echo -e "${BLUE}🧪 Running basic tests to verify setup...${NC}"
if python -m pytest tests/test_basic.py -v; then
    echo -e "${GREEN}✅ Basic tests passed${NC}"
else
    echo -e "${YELLOW}⚠️  Some tests failed, but setup is mostly complete${NC}"
fi

# Display setup summary
echo ""
echo -e "${GREEN}🎉 Setup completed successfully!${NC}"
echo ""
echo "📋 Next steps:"
echo "1. Edit .env file and add your API keys:"
echo "   - OPENAI_API_KEY=your-openai-key"
echo "   - PINECONE_API_KEY=your-pinecone-key"
echo "   - TAVILY_API_KEY=your-tavily-key"
echo ""
echo "2. Activate the virtual environment:"
echo "   source venv/bin/activate"
echo ""
echo "3. Start the development server:"
echo "   ./scripts/start-dev.sh"
echo ""
echo "4. Test the service:"
echo "   ./scripts/test-service.sh"
echo ""
echo "📚 Additional commands:"
echo "   Format code: black ."
echo "   Lint code: flake8 ."
echo "   Type check: mypy ."
echo "   Run tests: python -m pytest tests/ -v"
echo ""
echo "📖 Documentation:"
echo "   README.md - Complete service documentation"
echo "   Admin API docs: http://localhost:8091/admin/docs (when running)"
echo ""

# Check if API keys are set
echo -e "${BLUE}🔑 Checking API key configuration...${NC}"
source .env 2>/dev/null || true

missing_keys=""
if [ -z "$OPENAI_API_KEY" ]; then
    missing_keys="$missing_keys OPENAI_API_KEY"
fi
if [ -z "$PINECONE_API_KEY" ]; then
    missing_keys="$missing_keys PINECONE_API_KEY"
fi
if [ -z "$TAVILY_API_KEY" ]; then
    missing_keys="$missing_keys TAVILY_API_KEY"
fi

if [ -n "$missing_keys" ]; then
    echo -e "${YELLOW}⚠️  Missing API keys:$missing_keys${NC}"
    echo "Please add them to your .env file before starting the service"
else
    echo -e "${GREEN}✅ All API keys are configured${NC}"
fi

echo ""
echo -e "${GREEN}Setup complete! You're ready to develop with LLM Gateway Python.${NC}"
