#!/bin/bash
"""
Script to run Vietnamese university data ingestion inside the Docker container.
This ensures all dependencies are available and properly configured.
"""

echo "🚀 Running Vietnamese university data ingestion in Docker container..."
echo "📅 Date: $(date)"
echo ""

# Check if the container is running
if ! docker ps | grep -q "careerup-monorepo-llm-gateway-py-1"; then
    echo "❌ LLM Gateway container is not running!"
    echo "Please start the services first with: docker-compose up -d"
    exit 1
fi

echo "✅ LLM Gateway container is running"
echo ""

# Copy the ingestion script into the container
echo "📋 Copying ingestion script to container..."
docker cp ./ingest_vietnamese_data.py careerup-monorepo-llm-gateway-py-1:/app/

# Run the ingestion script inside the container
echo "🔄 Running Vietnamese data ingestion..."
docker exec -it careerup-monorepo-llm-gateway-py-1 python /app/ingest_vietnamese_data.py

echo ""
echo "✅ Vietnamese data ingestion completed!"
