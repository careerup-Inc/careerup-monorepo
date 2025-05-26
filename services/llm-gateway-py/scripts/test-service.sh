#!/bin/bash

# Test script for LLM Gateway Python service
set -e

echo "🧪 Testing LLM Gateway Python service..."

# Configuration
GRPC_PORT=${GRPC_PORT:-50054}
HTTP_PORT=${HTTP_PORT:-8091}
ADMIN_API_KEY=${ADMIN_API_KEY:-dev-admin-key}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test functions
test_health_check() {
    echo -e "${BLUE}🏥 Testing health check endpoint...${NC}"
    
    response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$HTTP_PORT/health)
    
    if [ "$response" = "200" ]; then
        echo -e "${GREEN}✅ Health check passed${NC}"
        return 0
    else
        echo -e "${RED}❌ Health check failed (HTTP $response)${NC}"
        return 1
    fi
}

test_admin_config() {
    echo -e "${BLUE}⚙️  Testing admin configuration endpoint...${NC}"
    
    response=$(curl -s -H "Authorization: Bearer $ADMIN_API_KEY" \
                    http://localhost:$HTTP_PORT/admin/config)
    
    if echo "$response" | jq . >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Admin config endpoint working${NC}"
        echo "Response: $(echo "$response" | jq -r '.service_name')"
        return 0
    else
        echo -e "${RED}❌ Admin config endpoint failed${NC}"
        echo "Response: $response"
        return 1
    fi
}

test_metrics() {
    echo -e "${BLUE}📊 Testing metrics endpoint...${NC}"
    
    response=$(curl -s -H "Authorization: Bearer $ADMIN_API_KEY" \
                    http://localhost:$HTTP_PORT/admin/metrics)
    
    if echo "$response" | jq . >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Metrics endpoint working${NC}"
        
        # Extract some key metrics
        total_requests=$(echo "$response" | jq -r '.current_stats.total_requests // 0')
        echo "Total requests: $total_requests"
        return 0
    else
        echo -e "${RED}❌ Metrics endpoint failed${NC}"
        echo "Response: $response"
        return 1
    fi
}

test_grpc_reflection() {
    echo -e "${BLUE}🔍 Testing gRPC reflection...${NC}"
    
    # Check if grpcurl is available
    if ! command -v grpcurl &> /dev/null; then
        echo -e "${YELLOW}⚠️  grpcurl not found, skipping gRPC reflection test${NC}"
        echo "Install grpcurl: brew install grpcurl"
        return 0
    fi
    
    services=$(grpcurl -plaintext localhost:$GRPC_PORT list 2>/dev/null)
    
    if echo "$services" | grep -q "llm.v1.LLMService"; then
        echo -e "${GREEN}✅ gRPC reflection working${NC}"
        echo "Available services:"
        echo "$services" | sed 's/^/  /'
        return 0
    else
        echo -e "${RED}❌ gRPC reflection failed${NC}"
        echo "Available services: $services"
        return 1
    fi
}

test_llm_service() {
    echo -e "${BLUE}🤖 Testing LLM service via admin API...${NC}"
    
    test_query='{
        "query": "Hello, this is a test query",
        "context": "",
        "use_rag": false,
        "language": "en"
    }'
    
    response=$(curl -s -X POST \
                    -H "Authorization: Bearer $ADMIN_API_KEY" \
                    -H "Content-Type: application/json" \
                    -d "$test_query" \
                    http://localhost:$HTTP_PORT/admin/test)
    
    if echo "$response" | jq . >/dev/null 2>&1; then
        request_id=$(echo "$response" | jq -r '.request_id // "unknown"')
        query_type=$(echo "$response" | jq -r '.query_type // "unknown"')
        
        echo -e "${GREEN}✅ LLM service test passed${NC}"
        echo "Request ID: $request_id"
        echo "Query Type: $query_type"
        return 0
    else
        echo -e "${RED}❌ LLM service test failed${NC}"
        echo "Response: $response"
        return 1
    fi
}

test_vietnamese_support() {
    echo -e "${BLUE}🇻🇳 Testing Vietnamese language support...${NC}"
    
    test_query='{
        "query": "Xin chào, đây là câu hỏi bằng tiếng Việt",
        "context": "",
        "use_rag": false,
        "language": "vi"
    }'
    
    response=$(curl -s -X POST \
                    -H "Authorization: Bearer $ADMIN_API_KEY" \
                    -H "Content-Type: application/json" \
                    -d "$test_query" \
                    http://localhost:$HTTP_PORT/admin/test)
    
    if echo "$response" | jq . >/dev/null 2>&1; then
        language=$(echo "$response" | jq -r '.language // "unknown"')
        
        if [ "$language" = "vietnamese" ]; then
            echo -e "${GREEN}✅ Vietnamese language detection working${NC}"
            echo "Detected language: $language"
            return 0
        else
            echo -e "${YELLOW}⚠️  Vietnamese detection may not be working properly${NC}"
            echo "Detected language: $language"
            return 1
        fi
    else
        echo -e "${RED}❌ Vietnamese test failed${NC}"
        echo "Response: $response"
        return 1
    fi
}

wait_for_service() {
    echo -e "${BLUE}⏳ Waiting for service to be ready...${NC}"
    
    max_attempts=30
    attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:$HTTP_PORT/health >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Service is ready${NC}"
            return 0
        fi
        
        echo "Attempt $attempt/$max_attempts - waiting..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo -e "${RED}❌ Service failed to start within $(($max_attempts * 2)) seconds${NC}"
    return 1
}

# Main test execution
main() {
    echo "🎯 Target service:"
    echo "   gRPC: localhost:$GRPC_PORT"
    echo "   HTTP: localhost:$HTTP_PORT"
    echo ""
    
    # Wait for service to be ready
    if ! wait_for_service; then
        echo -e "${RED}❌ Cannot proceed with tests - service not ready${NC}"
        exit 1
    fi
    
    echo ""
    echo "🧪 Running test suite..."
    echo ""
    
    # Run tests
    failed_tests=0
    
    test_health_check || failed_tests=$((failed_tests + 1))
    echo ""
    
    test_admin_config || failed_tests=$((failed_tests + 1))
    echo ""
    
    test_metrics || failed_tests=$((failed_tests + 1))
    echo ""
    
    test_grpc_reflection || failed_tests=$((failed_tests + 1))
    echo ""
    
    test_llm_service || failed_tests=$((failed_tests + 1))
    echo ""
    
    test_vietnamese_support || failed_tests=$((failed_tests + 1))
    echo ""
    
    # Summary
    echo "📋 Test Summary:"
    echo "=================="
    
    if [ $failed_tests -eq 0 ]; then
        echo -e "${GREEN}✅ All tests passed!${NC}"
        echo ""
        echo "🎉 LLM Gateway Python service is working correctly"
        exit 0
    else
        echo -e "${RED}❌ $failed_tests test(s) failed${NC}"
        echo ""
        echo "🔧 Please check the service logs and configuration"
        exit 1
    fi
}

# Check dependencies
if ! command -v curl &> /dev/null; then
    echo -e "${RED}❌ curl is required for testing${NC}"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo -e "${RED}❌ jq is required for JSON parsing${NC}"
    echo "Install jq: brew install jq"
    exit 1
fi

# Run main function
main "$@"
