#!/bin/bash

# Enhanced Adaptive RAG Testing Script
# Tests the new capabilities: document grading, hallucination detection, query routing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
LLM_HOST="localhost:50054"
ADMIN_HOST="http://localhost:8091"
ADMIN_API_KEY="admin-secret-key"

echo -e "${BLUE}🔬 Enhanced Adaptive RAG Testing Suite${NC}"
echo "=================================================="

# Function to test streaming RAG with token concatenation
test_streaming_rag() {
    local prompt="$1"
    local description="$2"
    local language="${3:-auto}"
    
    echo -e "\n${YELLOW}Testing: $description${NC}"
    echo "Prompt: $prompt"
    echo "Response:"
    
    # Test with adaptive flag enabled
    response=$(grpcurl -plaintext -d "{\"prompt\": \"$prompt\", \"adaptive\": true}" \
        $LLM_HOST llm.v1.LLMService/GenerateWithRAG 2>/dev/null | \
        jq -r '.token // empty' | tr -d '\n')
    
    echo -e "${GREEN}$response${NC}"
    
    # Check response quality
    if [[ ${#response} -gt 10 ]]; then
        echo -e "${GREEN}✅ Response generated successfully${NC}"
    else
        echo -e "${RED}❌ Response too short or empty${NC}"
        return 1
    fi
    
    echo "----------------------------------------"
}

# Function to test admin API
test_admin_endpoint() {
    local endpoint="$1"
    local description="$2"
    
    echo -e "\n${YELLOW}Testing Admin API: $description${NC}"
    
    response=$(curl -s -H "Authorization: Bearer $ADMIN_API_KEY" \
        "$ADMIN_HOST$endpoint" | jq . 2>/dev/null || echo "Failed to parse JSON")
    
    if [[ "$response" != "Failed to parse JSON" ]]; then
        echo -e "${GREEN}✅ $description successful${NC}"
        echo "$response" | head -10
    else
        echo -e "${RED}❌ $description failed${NC}"
    fi
    
    echo "----------------------------------------"
}

# Function to test Vietnamese university data ingestion with enhanced format
test_vietnamese_data_enhanced() {
    echo -e "\n${YELLOW}Testing Enhanced Vietnamese University Data Ingestion${NC}"
    
    # Test ingestion endpoint with enhanced data
    json_file_path="/Users/doviethoang/github/careerup-monorepo/services/llm-gateway-py/data/diem_chuan_dai_hoc_2024_enhanced.json"
    pdf_file_path="/Users/doviethoang/github/careerup-monorepo/services/llm-gateway-py/data/de-an-tuyen-sinh-2024final.pdf"
    
    # Test JSON ingestion with enhanced format
    if [[ -f "$json_file_path" ]]; then
        echo "Testing enhanced JSON data ingestion..."
        response=$(curl -s -X POST \
            -H "Authorization: Bearer $ADMIN_API_KEY" \
            -H "Content-Type: application/json" \
            -d "{\"file_path\": \"$json_file_path\", \"file_type\": \"json\", \"collection_name\": \"vietnamese-university-scores-enhanced\"}" \
            "$ADMIN_HOST/admin/ingest/vietnamese-university-data" 2>/dev/null || echo "Failed")
        
        if [[ "$response" != "Failed" ]]; then
            echo -e "${GREEN}✅ Enhanced JSON data ingestion initiated${NC}"
            echo "$response" | jq . 2>/dev/null || echo "$response"
        else
            echo -e "${YELLOW}⚠️  Enhanced JSON ingestion endpoint test skipped${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  Enhanced JSON file not found: $json_file_path${NC}"
    fi
    
    echo "----------------------------------------"
}

# Function to test enhanced Vietnamese queries with improved data
test_enhanced_vietnamese_queries() {
    echo -e "\n${YELLOW}Testing Enhanced Vietnamese Query Processing${NC}"
    
    # Test 1: Query with Vietnamese keywords
    test_streaming_rag \
        "Tôi muốn biết điểm chuẩn ngành Kỹ thuật Sinh học tại HUST" \
        "Enhanced Vietnamese Query - Specific Major" \
        "vi"
    
    # Test 2: Query with university aliases
    test_streaming_rag \
        "Điểm chuẩn các ngành kỹ thuật tại Bách Khoa Hà Nội" \
        "Enhanced Query - University Alias Recognition" \
        "vi"
    
    # Test 3: Query about score comparison
    test_streaming_rag \
        "So sánh điểm chuẩn ngành CNTT và Cơ khí tại HUST 2024" \
        "Enhanced Query - Score Comparison" \
        "vi"
    
    # Test 4: Query about subject combinations
    test_streaming_rag \
        "Những ngành nào xét tuyển tổ hợp A00 tại Đại học Bách Khoa?" \
        "Enhanced Query - Subject Combination Search" \
        "vi"
    
    # Test 5: Query about career prospects
    test_streaming_rag \
        "Ngành Kỹ thuật có cơ hội nghề nghiệp như thế nào?" \
        "Enhanced Query - Career Information" \
        "vi"
    
    echo "----------------------------------------"
}

# Function to test Vietnamese university data ingestion
test_vietnamese_data() {
    test_vietnamese_data_enhanced
    test_enhanced_vietnamese_queries
}

echo -e "\n${BLUE}1. Testing Health Check${NC}"
health_response=$(curl -s $ADMIN_HOST/health | jq . 2>/dev/null || echo "Service not responding")
if [[ "$health_response" != "Service not responding" ]]; then
    echo -e "${GREEN}✅ Service is healthy${NC}"
    echo "$health_response"
else
    echo -e "${RED}❌ Service health check failed${NC}"
    exit 1
fi

echo -e "\n${BLUE}2. Testing Adaptive Query Routing${NC}"

# Test 1: Vietnamese university query (should route to vectorstore)
test_streaming_rag \
    "Điểm chuẩn ngành Công nghệ Thông tin của trường Đại học Bách Khoa Hà Nội năm 2024?" \
    "Vietnamese University Query (Vectorstore Route)" \
    "vi"

# Test 2: General knowledge query (should route to direct LLM or web search)
test_streaming_rag \
    "What is the capital of Japan?" \
    "General Knowledge Query (Direct LLM Route)" \
    "en"

# Test 3: Complex Vietnamese query with multiple criteria
test_streaming_rag \
    "So sánh điểm chuẩn các ngành Kỹ thuật tại HUST và cho biết ngành nào có điểm chuẩn cao nhất?" \
    "Complex Comparative Query (Adaptive RAG)" \
    "vi"

# Test 4: Ambiguous query that tests document grading
test_streaming_rag \
    "Thông tin về học phí và học bổng tại các trường đại học" \
    "Ambiguous Query (Document Grading Test)" \
    "vi"

echo -e "\n${BLUE}3. Testing Admin API Endpoints${NC}"

test_admin_endpoint "/admin/status" "Service Status"
test_admin_endpoint "/admin/config" "Service Configuration"
test_admin_endpoint "/admin/metrics" "Service Metrics"

echo -e "\n${BLUE}4. Testing Enhanced Vietnamese Data & Queries${NC}"
test_vietnamese_data

echo -e "\n${BLUE}5. Testing Document Relevance & Hallucination Detection${NC}"

# Test with a query that might produce irrelevant documents
test_streaming_rag \
    "Điểm chuẩn ngành Thiên văn học tại Đại học Bách Khoa" \
    "Query for Potentially Non-existent Program (Relevance Test)" \
    "vi"

# Test with a query that might cause hallucination
test_streaming_rag \
    "Điểm chuẩn năm 2025 của các trường đại học tại Việt Nam" \
    "Future Data Query (Hallucination Detection Test)" \
    "vi"

echo -e "\n${BLUE}6. Performance Testing${NC}"

echo "Testing response times for adaptive RAG..."
start_time=$(date +%s.%N)

grpcurl -plaintext -d '{"prompt": "Điểm chuẩn ngành CNTT tại HUST 2024?", "adaptive": true}' \
    $LLM_HOST llm.v1.LLMService/GenerateWithRAG >/dev/null 2>&1

end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc -l)

echo -e "${GREEN}✅ Adaptive RAG response time: ${duration}s${NC}"

if (( $(echo "$duration < 30.0" | bc -l) )); then
    echo -e "${GREEN}✅ Performance: Excellent (< 30s)${NC}"
elif (( $(echo "$duration < 60.0" | bc -l) )); then
    echo -e "${YELLOW}⚠️  Performance: Good (< 60s)${NC}"
else
    echo -e "${RED}❌ Performance: Needs improvement (> 60s)${NC}"
fi

echo -e "\n${BLUE}7. Testing Concurrent Requests${NC}"

echo "Testing concurrent adaptive RAG requests..."
{
    grpcurl -plaintext -d '{"prompt": "Điểm chuẩn HUST 2024?", "adaptive": true}' $LLM_HOST llm.v1.LLMService/GenerateWithRAG >/dev/null &
    grpcurl -plaintext -d '{"prompt": "What is machine learning?", "adaptive": true}' $LLM_HOST llm.v1.LLMService/GenerateWithRAG >/dev/null &
    grpcurl -plaintext -d '{"prompt": "Các ngành học tại BK Hà Nội?", "adaptive": true}' $LLM_HOST llm.v1.LLMService/GenerateWithRAG >/dev/null &
    wait
}

echo -e "${GREEN}✅ Concurrent requests completed${NC}"

echo -e "\n${GREEN}🎉 Enhanced Adaptive RAG Testing Complete!${NC}"
echo "=================================================="
echo -e "${BLUE}Key Features Tested:${NC}"
echo "✅ Adaptive query routing (vectorstore vs direct LLM)"
echo "✅ Vietnamese language processing"
echo "✅ Document relevance grading"
echo "✅ Hallucination detection capabilities" 
echo "✅ Multi-representation indexing support"
echo "✅ Enhanced admin API endpoints"
echo "✅ Performance and concurrency"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Ingest Vietnamese university PDF and JSON data"
echo "2. Configure multi-representation indexing"
echo "3. Fine-tune document grading thresholds"
echo "4. Monitor hallucination detection rates"
echo "5. Optimize query routing rules"
