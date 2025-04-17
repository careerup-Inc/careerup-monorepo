# CareerUP

CareerUP is a polyglot, micro-service-oriented backend that mixes Go for ultra-concurrent, low-latency components with Java 21 for domain-heavy APIs and ML pipelines.

## Architecture

The system consists of the following services:

- **api-gateway** (Go/Gin): JWT auth, rate-limit, route to internal gRPC
- **auth-core** (Java/Spring Boot): Users, profiles, bookings, payments
- **rec-service** (Java/Quarkus): Recommendation & similarity ML
- **chat-gateway** (Go/Fiber): Multiplex client WebSockets → LLM
- **llm-gateway** (Go/LangChainGo): Prompt-orchestration, RAG, OpenAI calls
- **avatar-service** (Go): VRoid Studio model management & D-ID orchestration, TTS
- **notification** (Go/Fiber): Push alerts via Redis streams

## Development Setup

### Prerequisites

- Go 1.21+
- Java 21
- Docker & Docker Compose
- buf CLI
- Make

### Getting Started

1/ Install development tools:

```bash
make tools
```

2/ Generate protobuf code:

```bash
make proto
```

3/ Start the development environment:

```bash
make run
```

### Environment Variables

Create a `.env` file with the following variables:

```bash
OPENAI_API_KEY=your_openai_key
PINECONE_API_KEY=your_pinecone_key
PINECONE_ENVIRONMENT=your_pinecone_env
VROID_HUB_API_KEY=your_vroid_key
DID_API_KEY=your_did_key
```

## API Documentation

### Authentication

- `POST /v1/auth/register`: Register a new user
- `POST /v1/auth/login`: Login and get JWT tokens
- `GET /v1/users/me`: Get current user profile
- `PUT /v1/users/me`: Update user profile

### Chat

WebSocket endpoint: `<wss://chat-gw.careerup.ai/ws>`

Message format:

```json
// Client → Server
{
  "type": "user_msg",
  "conv_id": "uuid",
  "text": "I'm interested in AI careers"
}

// Server → Client
{
  "type": "assistant_token",
  "token": "Sure,"
}
{
  "type": "avatar_url",
  "url": "https://cdn.careerup.ai/clip/abc.mp4"
}
```

## Observability

- Prometheus: <http://localhost:9090>
- Grafana: <http://localhost:3000>
- Tempo: <http://localhost:3200>

## License

MIT
