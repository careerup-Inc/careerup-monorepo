# CareerUP Testing Suite

This directory contains the unified testing infrastructure for the CareerUP monorepo.

## Overview

The testing suite has been consolidated from multiple scattered test files into a unified, organized structure that covers all aspects of the CareerUP system.

## Test Files

### Shell Scripts
- **`test-suite.sh`** - Main unified test suite for all components
- **`test-rag-vietnamese.sh`** - Specialized Vietnamese RAG testing (kept for detailed Vietnamese-specific testing)

### Python Scripts
- **`test-python-suite.py`** - Consolidated Python testing for services and components

## Quick Start

### Run All Tests
```bash
# Shell-based comprehensive testing
./tests/test-suite.sh

# Python-based service testing
./tests/test-python-suite.py
```

### Run Specific Test Categories

#### Shell Test Suite
```bash
# Core functionality only (health + auth + llm)
./tests/test-suite.sh core

# Health checks only
./tests/test-suite.sh health

# Authentication tests only
./tests/test-suite.sh auth

# LLM Gateway tests only
./tests/test-suite.sh llm

# Vietnamese RAG tests only
./tests/test-suite.sh vietnamese

# WebSocket chat tests only
./tests/test-suite.sh websocket
```

#### Python Test Suite
```bash
# Core Python functionality
./tests/test-python-suite.py core

# Service-specific tests
./tests/test-python-suite.py services

# Individual test categories
./tests/test-python-suite.py imports
./tests/test-python-suite.py config
./tests/test-python-suite.py health
./tests/test-python-suite.py embedding
./tests/test-python-suite.py pinecone
./tests/test-python-suite.py admin
```

## Test Categories

### 1. Health Checks
- API Gateway health
- Auth Service health  
- LLM Gateway Python health
- Service connectivity validation

### 2. Authentication
- User registration
- User login
- Token validation
- Access control testing

### 3. LLM Gateway
- Basic chat completion
- RAG functionality
- Vietnamese language support
- Embedding system validation

### 4. Vietnamese RAG
- Vietnamese university data queries
- Abbreviation handling
- Complex academic queries
- Score pattern recognition

### 5. WebSocket Chat
- Connection establishment
- Message handling
- Authentication integration

### 6. Python Services
- Import validation
- Configuration loading
- Service component testing
- External dependency validation

## Prerequisites

### Required Services
Make sure these services are running before executing tests:

```bash
# Start all services
docker-compose up -d

# Check service status
docker-compose ps
```

### Required Environment Variables
Ensure these are set in your `.env` file:
- `PINECONE_API_KEY` - For vector database testing
- `EMBEDDING_DIMENSIONS=384` - Correct embedding dimensions
- `EMBEDDING_MODEL=llama` - Embedding model configuration

### Required Dependencies
For Python tests:
```bash
pip install requests pinecone
```

## Test Results

Tests provide colored output for easy interpretation:
- ✅ **Green** - Test passed
- ⚠️ **Yellow** - Warning or partial success
- ❌ **Red** - Test failed
- ℹ️ **Cyan** - Information

## Troubleshooting

### Common Issues

1. **Service Not Responding**
   ```bash
   # Check if services are running
   docker-compose ps
   
   # Restart specific service
   docker-compose restart llm-gateway-py
   ```

2. **Authentication Failures**
   ```bash
   # May need to register test user first
   # The test suite will attempt this automatically
   ```

3. **Vietnamese RAG Dimension Mismatch**
   ```bash
   # Ensure correct configuration
   grep EMBEDDING_DIMENSIONS .env
   # Should show: EMBEDDING_DIMENSIONS=384
   ```

4. **Pinecone Connection Issues**
   ```bash
   # Check API key
   echo $PINECONE_API_KEY
   
   # Verify index exists
   # Check Pinecone dashboard
   ```

## Legacy Test Cleanup

The following files have been **consolidated** into the unified test suite:

### Removed Shell Scripts
- `test-adaptive-rag-enhanced.sh` → Merged into `test-suite.sh`
- `test-basic-chat.sh` → Merged into `test-suite.sh` 
- `test-python-llm-gateway.sh` → Merged into `test-suite.sh`
- `test-auth.sh` → Merged into `test-suite.sh`
- `test-ingest-endpoint.sh` → Functionality integrated
- `test-ilo.sh` → Functionality integrated

### Removed Python Scripts
- `services/llm-gateway-py/test_simple.py` → Merged into `test-python-suite.py`
- `services/llm-gateway-py/test_integration.py` → Merged into `test-python-suite.py`
- `services/llm-gateway-py/test_standalone.py` → Merged into `test-python-suite.py`
- `services/llm-gateway-py/test_service_*.py` → Consolidated
- `test-embeddings.py` → Merged into `test-python-suite.py`

### Kept Specialized Files
- `test-rag-vietnamese.sh` - Kept for detailed Vietnamese testing (working solution)

## Integration with CI/CD

The unified test suite is designed for easy integration with CI/CD pipelines:

```yaml
# Example GitHub Actions usage
- name: Run Core Tests
  run: ./tests/test-suite.sh core

- name: Run Python Service Tests  
  run: ./tests/test-python-suite.py services

- name: Run Vietnamese RAG Tests
  run: ./tests/test-rag-vietnamese.sh
```

## Contributing

When adding new tests:
1. Add them to the appropriate unified test suite
2. Follow the existing pattern for colored output
3. Include proper error handling and cleanup
4. Update this README with new test categories
