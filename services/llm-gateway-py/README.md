# LLM Gateway Python Service

A Python-based LLM Gateway microservice, providing prompt orchestration, RAG (Retrieval-Augmented Generation), OpenAI integration, and Tavily web search with Vietnamese language support.

## Features

- **gRPC Service**: Compatible with existing Go implementation interface
- **RAG (Retrieval-Augmented Generation)**: Document retrieval and grading using Pinecone vector store
- **Vietnamese Language Support**: Specialized prompting and response formatting
- **Adaptive Query Routing**: Intelligent routing between vectorstore, web search, and direct LLM
- **Web Search Integration**: Tavily API for real-time information retrieval
- **OpenAI Integration**: GPT models with embeddings and chat completions
- **Admin HTTP API**: FastAPI-based management endpoints
- **Comprehensive Monitoring**: Metrics collection and health checks
- **Docker Support**: Containerized deployment with docker-compose

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   gRPC Client   │───▶│  LLM Gateway Py  │───▶│   OpenAI API    │
└─────────────────┘    │                  │    └─────────────────┘
                       │  ┌─────────────┐ │    ┌─────────────────┐
                       │  │ Query Router│ │───▶│  Pinecone VDB   │
                       │  └─────────────┘ │    └─────────────────┘
                       │  ┌─────────────┐ │    ┌─────────────────┐
                       │  │ RAG Engine  │ │───▶│   Tavily API    │
                       │  └─────────────┘ │    └─────────────────┘
                       └──────────────────┘
                                │
                       ┌──────────────────┐
                       │   Admin HTTP     │
                       │   FastAPI        │
                       └──────────────────┘
```

## Quick Start

### Prerequisites

- Python 3.11+
- OpenAI API key
- Pinecone account and API key
- Tavily API key (for web search)

### Installation

1. **Clone and navigate to the service directory:**
   ```bash
   cd services/llm-gateway-py
   ```

2. **Install dependencies:**
   ```bash
   pip install -r requirements.txt
   ```

3. **Configure environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your API keys
   ```

4. **Run the service:**
   ```bash
   python main.py
   ```

### Using Docker

1. **Build and run with docker-compose:**
   ```bash
   docker-compose up --build
   ```

2. **Or build and run manually:**
   ```bash
   docker build -t llm-gateway-py .
   docker run -p 50054:50054 -p 8091:8091 --env-file .env llm-gateway-py
   ```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `OPENAI_API_KEY` | OpenAI API key | - | Yes |
| `PINECONE_API_KEY` | Pinecone API key | - | Yes |
| `TAVILY_API_KEY` | Tavily API key | - | Yes |
| `GRPC_PORT` | gRPC server port | 50054 | No |
| `HTTP_PORT` | HTTP admin port | 8091 | No |
| `ENVIRONMENT` | Environment name | development | No |
| `DEBUG` | Debug mode | false | No |
| `LOG_LEVEL` | Logging level | INFO | No |
| `ADMIN_API_KEY` | Admin API key | admin-secret-key | No |

### RAG Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `RAG_CHUNK_SIZE` | Document chunk size | 1000 |
| `RAG_CHUNK_OVERLAP` | Chunk overlap | 200 |
| `RAG_RETRIEVAL_TOP_K` | Top K results | 5 |
| `RAG_TEMPERATURE` | LLM temperature | 0.7 |
| `RAG_MAX_TOKENS` | Max response tokens | 1000 |

## API Reference

### gRPC Service

The service implements the `LLMService` interface defined in `proto/llm/v1/llm.proto`:

#### ProcessPrompt
```protobuf
rpc ProcessPrompt(PromptRequest) returns (PromptResponse);
```

**Example Request:**
```json
{
  "query": "What are the admission requirements for computer science?",
  "context": "Additional context about the query",
  "use_rag": true,
  "language": "vi"
}
```

**Example Response:**
```json
{
  "request_id": "uuid-string",
  "response": "Detailed response text in Vietnamese",
  "query_type": "rag",
  "language": "vietnamese",
  "tokens_used": 150,
  "sources": [
    {
      "content": "Relevant document content",
      "url": "https://example.com/doc",
      "title": "Document Title",
      "score": 0.95
    }
  ]
}
```

#### IngestDocuments
```protobuf
rpc IngestDocuments(IngestRequest) returns (IngestResponse);
```

Ingest documents into the vector store for RAG.

### HTTP Admin API

The admin API is available at `http://localhost:8091/admin/` when enabled.

#### Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/health` | Health check | No |
| GET | `/admin/config` | Service configuration | Yes |
| GET | `/admin/metrics` | Service metrics | Yes |
| GET | `/admin/metrics/export` | Export metrics (JSON/Prometheus) | Yes |
| POST | `/admin/test` | Test query processing | Yes |
| POST | `/admin/ingest` | Ingest documents | Yes |
| GET | `/admin/status` | Detailed service status | Yes |

**API Documentation:** Available at `http://localhost:8091/admin/docs`

#### Authentication

Use the `Authorization: Bearer <ADMIN_API_KEY>` header for authenticated endpoints.

## Features

### Vietnamese Language Support

The service includes comprehensive Vietnamese language processing:

- **Language Detection**: Automatic Vietnamese text detection
- **Query Normalization**: Unicode normalization and common error corrections
- **Keyword Extraction**: Vietnamese-aware keyword extraction
- **Response Formatting**: Vietnamese formatting conventions
- **Intent Detection**: Vietnamese query intent analysis

### Adaptive Query Routing

The service intelligently routes queries based on content analysis:

1. **Document Retrieval**: For factual questions about stored documents
2. **Web Search**: For current events, recent information, or when documents lack relevance
3. **Direct LLM**: For general questions, conversations, or creative tasks

### RAG (Retrieval-Augmented Generation)

Advanced RAG implementation with:

- **Document Grading**: Relevance scoring and filtering
- **Fallback Mechanisms**: Web search when vector store results are insufficient
- **Multi-source Fusion**: Combining multiple information sources
- **Context Optimization**: Smart context window management

## Monitoring and Observability

### Metrics

The service collects comprehensive metrics:

- Request counts and success rates
- Response times and token usage
- Error distribution and types
- Language and query type statistics
- Model usage patterns

### Health Checks

- **HTTP Health Endpoint**: `/health`
- **Docker Health Check**: Built into Docker container
- **gRPC Reflection**: Enabled for service discovery

### Logging

Structured JSON logging with:

- Request tracing with unique IDs
- Performance metrics
- Error tracking with stack traces
- Vietnamese language detection results

## Development

### Project Structure

```
llm-gateway-py/
├── admin/              # FastAPI admin endpoints
├── config/             # Configuration management
├── services/           # Core gRPC service implementation
├── utils/              # Utility modules
│   ├── logger.py       # Logging utilities
│   ├── vietnamese.py   # Vietnamese language support
│   ├── security.py     # Security utilities
│   ├── metrics.py      # Metrics collection
│   └── helpers.py      # Helper functions
├── main.py            # Service entry point
├── requirements.txt   # Python dependencies
├── Dockerfile         # Docker configuration
└── docker-compose.yml # Docker Compose setup
```

### Running Tests

```bash
# Install test dependencies
pip install pytest pytest-asyncio pytest-cov

# Run tests
pytest tests/ -v

# Run with coverage
pytest tests/ --cov=. --cov-report=html
```

### Code Quality

```bash
# Format code
black .

# Lint code
flake8 .

# Type checking
mypy .
```

## Deployment

### Docker Deployment

1. **Production docker-compose:**
   ```yaml
   # Use production environment variables
   environment:
     - ENVIRONMENT=production
     - DEBUG=false
     - LOG_LEVEL=INFO
   ```

2. **Health checks and restart policies:**
   ```yaml
   restart: unless-stopped
   healthcheck:
     interval: 30s
     timeout: 10s
     retries: 3
   ```

### Kubernetes Deployment

Example Kubernetes manifests are available in the `k8s/` directory.

### Environment-Specific Configuration

- **Development**: Full logging, debug mode enabled
- **Staging**: Production-like with detailed logging
- **Production**: Optimized performance, minimal logging

## Performance Tuning

### Configuration Optimization

```bash
# Increase worker threads for high load
MAX_WORKERS=20

# Optimize RAG parameters
RAG_CHUNK_SIZE=1500
RAG_RETRIEVAL_TOP_K=3
RAG_MAX_TOKENS=800
```

### Monitoring Performance

```bash
# View real-time metrics
curl -H "Authorization: Bearer $ADMIN_API_KEY" \
     http://localhost:8091/admin/metrics

# Export Prometheus metrics
curl -H "Authorization: Bearer $ADMIN_API_KEY" \
     http://localhost:8091/admin/metrics/export?format=prometheus
```

## Troubleshooting

### Common Issues

1. **Import errors for proto modules:**
   ```bash
   export PYTHONPATH=/path/to/careerup-monorepo/proto:$PYTHONPATH
   ```

2. **Missing API keys:**
   ```bash
   # Check configuration
   curl http://localhost:8091/health
   ```

3. **Vector store connection issues:**
   ```bash
   # Verify Pinecone configuration
   curl -H "Authorization: Bearer $ADMIN_API_KEY" \
        http://localhost:8091/admin/status
   ```

### Debugging

1. **Enable debug logging:**
   ```bash
   export DEBUG=true
   export LOG_LEVEL=DEBUG
   ```

2. **Test individual components:**
   ```bash
   # Test query processing
   curl -X POST http://localhost:8091/admin/test \
        -H "Authorization: Bearer $ADMIN_API_KEY" \
        -H "Content-Type: application/json" \
        -d '{"query": "Test query", "use_rag": true}'
   ```

## Contributing

1. **Follow code style:**
   - Use Black for formatting
   - Follow PEP 8 guidelines
   - Add type hints

2. **Write tests:**
   - Unit tests for utilities
   - Integration tests for services
   - End-to-end tests for gRPC

3. **Update documentation:**
   - Update README for new features
   - Add docstrings to functions
   - Update API documentation

## License

This service is part of the CareerUp monorepo and follows the same licensing terms.

## Support

For issues and questions:
1. Check the troubleshooting section
2. Review logs in debug mode
3. Use the admin API for diagnostics
4. Create an issue in the monorepo
