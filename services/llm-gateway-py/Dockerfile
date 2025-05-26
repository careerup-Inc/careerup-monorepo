# Python LLM Gateway Service Dockerfile
FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first for better caching
COPY services/llm-gateway-py/requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy proto files
COPY proto/ ./proto/

# Copy application code
COPY services/llm-gateway-py/ ./

# Create non-root user for security
RUN useradd --create-home --shell /bin/bash app \
    && chown -R app:app /app
USER app

# Expose ports
EXPOSE 50054 8091

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import grpc; grpc.channel_ready_future(grpc.insecure_channel('localhost:50054')).result(timeout=10)"

# Set environment variables
ENV PYTHONPATH=/app/proto:/app
ENV PYTHONUNBUFFERED=1

# Run the service
CMD ["python", "main.py"]
