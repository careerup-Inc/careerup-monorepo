#!/bin/bash

# Test script for Vietnamese university RAG system
# This script tests the fixed embedding system with the university-scores collection

echo "🇻🇳 Testing Vietnamese University RAG System"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test queries in Vietnamese
QUERIES=(
    # Basic university queries
    "Điểm chuẩn trường Đại học Bách khoa Hà Nội năm 2024"
    "Trường đại học nào có điểm chuẩn cao nhất?"
    "Điểm chuẩn ngành Y khoa"
    "Các trường đại học ở Hà Nội"
    
    # Abbreviation handling tests
    "ĐHBK HN có những ngành nào?"
    "Điểm chuẩn ĐHQGHN khoa CNTT"
    "ĐHKT có mấy cơ sở đào tạo?"
    
    # Complex academic queries
    "Tổ hợp môn xét tuyển vào ngành kỹ thuật điện HUST"
    "Chỉ tiêu tuyển sinh đại học Y Hà Nội 2023-2024"
    "Điểm sàn xét tuyển vào ĐHQG TPHCM"
    
    # Score pattern queries
    "Ngành nào lấy điểm trên 27 điểm năm 2023?"
    "Điểm chuẩn khối A00 cao nhất năm nay"
    "Điểm xét học bạ ngành Luật NEU"
    
    # Mixed Vietnamese-English queries
    "Top 5 university ngành engineering ở Hà Nội"
    "What are the admission scores for HUST engineering?"
    "Điểm thi admission vào RMIT như thế nào?"
    
    # Admission policy queries
    "Phương thức xét tuyển của ĐH Kinh tế Quốc dân"
    "Cách tính điểm ưu tiên khu vực trong xét tuyển đại học"
)

# Function to test WebSocket connection
test_websocket() {
    local query="$1"
    local test_name="$2"
    
    echo -e "${BLUE}Testing: $test_name${NC}"
    echo -e "${YELLOW}Query: $query${NC}"
    
    # Create WebSocket test file
    cat > temp_ws_test.html << EOF
<!DOCTYPE html>
<html>
<head>
    <title>RAG Test</title>
</head>
<body>
    <script>
        const ws = new WebSocket('ws://localhost:8082/chat');
        
        ws.onopen = function() {
            console.log('WebSocket connected');
            
            const message = {
                type: 'chat',
                content: '$query',
                conversation_id: 'test-$(date +%s)',
                user_id: 'test-user',
                adaptive: true,
                use_rag: true
            };
            
            ws.send(JSON.stringify(message));
        };
        
        ws.onmessage = function(event) {
            const data = JSON.parse(event.data);
            console.log('Response:', data);
        };
        
        ws.onerror = function(error) {
            console.error('WebSocket error:', error);
        };
        
        ws.onclose = function() {
            console.log('WebSocket closed');
        };
        
        // Close after 10 seconds
        setTimeout(() => {
            ws.close();
        }, 10000);
    </script>
</body>
</html>
EOF
    
    echo "Created WebSocket test file. Open in browser to test manually."
    echo ""
}

# Function to test gRPC connection using grpcurl (if available)
test_grpc() {
    local query="$1"
    local test_name="$2"
    
    echo -e "${BLUE}Testing gRPC: $test_name${NC}"
    echo -e "${YELLOW}Query: $query${NC}"
    
    if command -v grpcurl &> /dev/null; then
        # echo -e "${BLUE}Raw streaming response:${NC}"
        grpcurl -plaintext \
            -d "{\"prompt\":\"$query\",\"user_id\":\"test-user\",\"conversation_id\":\"test-conv\",\"rag_collection\":\"university-scores\",\"adaptive\":true}" \
            localhost:50053 \
            llm.v1.LLMService/GenerateWithRAG
    else
        echo -e "${YELLOW}grpcurl not available, skipping gRPC test${NC}"
    fi
    echo ""
}

# Function to test Python LLM Gateway with concatenated response
test_python_llm_gateway() {
    local query="$1"
    local test_name="$2"
    
    echo -e "${BLUE}Testing Python LLM Gateway: $test_name${NC}"
    echo -e "${YELLOW}Query: $query${NC}"
    
    if command -v grpcurl &> /dev/null; then
        echo -e "${BLUE}Collecting streaming tokens...${NC}"
        
        # Capture the streaming response and extract tokens
        local raw_response=$(timeout 30s grpcurl -plaintext \
            -d "{\"prompt\":\"$query\",\"user_id\":\"test-user\",\"conversation_id\":\"test-conv\",\"rag_collection\":\"university-scores\",\"adaptive\":true}" \
            localhost:50054 \
            llm.v1.LLMService/GenerateWithRAG 2>/dev/null)
        
        if [ $? -eq 0 ] && [ ! -z "$raw_response" ]; then
            # Extract and concatenate tokens
            local concatenated_response=$(echo "$raw_response" | \
                grep '"token"' | \
                sed 's/.*"token": *"\([^"]*\)".*/\1/' | \
                tr -d '\n')
            
            echo -e "${GREEN}✅ Complete Response:${NC}"
            echo -e "${YELLOW}${concatenated_response}${NC}"
            echo ""
            
            # Count tokens for statistics
            local token_count=$(echo "$raw_response" | grep -c '"token"')
            echo -e "${BLUE}📊 Token count: $token_count${NC}"
            
        else
            echo -e "${RED}❌ No response received or timeout${NC}"
        fi
    else
        echo -e "${YELLOW}grpcurl not available, skipping Python LLM Gateway test${NC}"
    fi
    echo ""
}

# Function to check service health
check_services() {
    echo -e "${BLUE}Checking service health...${NC}"
    
    # Check Chat Gateway
    if curl -s http://localhost:8082/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Chat Gateway (8082) - OK${NC}"
    else
        echo -e "${RED}❌ Chat Gateway (8082) - DOWN${NC}"
    fi
    
    # Check LLM Gateway (Go - original)
    if curl -s http://localhost:8090/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ LLM Gateway Go (8090) - OK${NC}"
    else
        echo -e "${RED}❌ LLM Gateway Go (8090) - DOWN${NC}"
    fi
    
    # Check LLM Gateway (Python)
    if curl -s http://localhost:8091/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ LLM Gateway Python (8091) - OK${NC}"
        # Test gRPC connection
        if command -v grpcurl &> /dev/null; then
            if grpcurl -plaintext localhost:50054 list > /dev/null 2>&1; then
                echo -e "${GREEN}✅ Python LLM Gateway gRPC (50054) - OK${NC}"
            else
                echo -e "${RED}❌ Python LLM Gateway gRPC (50054) - DOWN${NC}"
            fi
        fi
    else
        echo -e "${RED}❌ LLM Gateway Python (8091) - DOWN${NC}"
    fi
    
    # Check Auth Core
    if curl -s http://localhost:8081/actuator/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Auth Core (8081) - OK${NC}"
    else
        echo -e "${RED}❌ Auth Core (8081) - DOWN${NC}"
    fi
    
    echo ""
}

# Function to test Pinecone connectivity
test_pinecone() {
    echo -e "${BLUE}Testing Pinecone connectivity...${NC}"
    
    python3 -c "
import os
import sys
try:
    from pinecone import Pinecone
    
    # Load environment variables
    api_key = os.getenv('PINECONE_API_KEY', 'pcsk_2YfBuW_QqtcQCPrVihorkoVJtHgKMUGji6htGJP1qwnYySPp5NBdNkrWcrSS6jUZNMXSdC')
    
    pc = Pinecone(api_key=api_key)
    
    # List indexes
    indexes = pc.list_indexes()
    print(f'📊 Available indexes: {[idx.name for idx in indexes]}')
    
    # Test university-scores index
    if any(idx.name == 'university-scores' for idx in indexes):
        index = pc.Index('university-scores')
        stats = index.describe_index_stats()
        print(f'🎓 university-scores stats:')
        print(f'   - Total vectors: {stats.total_vector_count}')
        print(f'   - Dimensions: {stats.dimension}')
        print(f'   - Index fullness: {stats.index_fullness}')
        
        # Test query
        test_vector = [0.1] * stats.dimension
        results = index.query(vector=test_vector, top_k=3, include_metadata=True)
        print(f'   - Query test: Found {len(results.matches)} matches')
        
        if results.matches:
            print(f'   - Sample result score: {results.matches[0].score:.4f}')
    else:
        print('❌ university-scores index not found')
        
except ImportError:
    print('⚠️  Pinecone library not installed. Run: pip install pinecone-client')
except Exception as e:
    print(f'❌ Pinecone test failed: {e}')
" 2>/dev/null || echo -e "${YELLOW}⚠️  Could not test Pinecone connectivity${NC}"
    
    echo ""
}

# Function to test embedding compatibility
test_embeddings() {
    echo -e "${BLUE}Testing embedding models...${NC}"
    
    python3 -c "
import os
import requests
import json

# Test Hugging Face embedding API
hf_api_key = os.getenv('HUGGINGFACE_API_KEY', 'hf_your_token_here')

if hf_api_key != 'hf_your_token_here':
    try:
        url = 'https://api-inference.huggingface.co/pipeline/feature-extraction/sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2'
        headers = {'Authorization': f'Bearer {hf_api_key}'}
        data = {'inputs': 'Đại học Bách khoa Hà Nội'}
        
        response = requests.post(url, headers=headers, json=data, timeout=30)
        
        if response.status_code == 200:
            embeddings = response.json()
            if isinstance(embeddings, list) and len(embeddings) > 0:
                print(f'✅ Llama embeddings working: {len(embeddings[0])} dimensions')
            else:
                print(f'⚠️  Unexpected embedding format: {type(embeddings)}')
        else:
            print(f'❌ Hugging Face API error: {response.status_code}')
            
    except Exception as e:
        print(f'❌ Embedding test failed: {e}')
else:
    print('⚠️  HUGGINGFACE_API_KEY not set, using fallback')
" 2>/dev/null || echo -e "${YELLOW}⚠️  Could not test embeddings${NC}"
    
    echo ""
}

# Main execution
main() {
    echo -e "${GREEN}Starting RAG system tests...${NC}"
    echo ""
    
    # Check if services are running
    check_services
    
    # Test Pinecone connectivity
    test_pinecone
    
    # Test embedding models
    test_embeddings
    
    # Test each query
    for i in "${!QUERIES[@]}"; do
        query="${QUERIES[$i]}"
        test_name="Test $((i+1))"
        
        echo -e "${GREEN}=== $test_name ===${NC}"
        test_websocket "$query" "$test_name"
        test_grpc "$query" "$test_name"
        test_python_llm_gateway "$query" "$test_name"
        echo -e "${GREEN}=================${NC}"
        echo ""
    done
    
    echo -e "${GREEN}🎉 RAG system testing completed!${NC}"
    echo ""
    echo -e "${YELLOW}📝 Manual testing steps:${NC}"
    echo "1. Open temp_ws_test.html in your browser"
    echo "2. Check browser console for WebSocket responses"
    echo "3. Verify that responses mention Vietnamese universities"
    echo "4. Check Docker logs: docker-compose logs llm-gateway"
    echo ""
    echo -e "${YELLOW}🔍 Key things to verify:${NC}"
    echo "- Embedding model shows 'llama' in logs"
    echo "- Retrieved documents contain Vietnamese university data"
    echo "- Responses are relevant to Vietnamese education"
    echo "- No fallback to web search for university queries"
}

# Cleanup function
cleanup() {
    rm -f temp_ws_test.html
}

# Set trap for cleanup
trap cleanup EXIT

# Run main function
main "$@"
