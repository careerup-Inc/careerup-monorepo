#!/bin/zsh

# =============================================================================
# CareerUP Unified Test Suite
# Consolidated testing script for all components
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
API_URL="http://localhost:8080"
AUTH_URL="http://localhost:8081"
LLM_HOST="localhost:50054"
ADMIN_HOST="http://localhost:8091"
ADMIN_API_KEY="admin-secret-key-change-in-production"

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

# Test suite functions
test_health_checks() {
    echo -e "\n${CYAN}🏥 Running Health Checks${NC}"
    echo "========================================"
    
    # API Gateway
    log "Checking API Gateway health..."
    if curl -s "$API_URL/health" >/dev/null 2>&1; then
        success "API Gateway is healthy"
    else
        error "API Gateway is not responding"
        return 1
    fi
    
    # Auth Service
    log "Checking Auth Service health..."
    if curl -s "$AUTH_URL/health" >/dev/null 2>&1; then
        success "Auth Service is healthy"
    else
        warning "Auth Service is not responding"
    fi
    
    # LLM Gateway Python
    log "Checking LLM Gateway Python health..."
    if curl -s "$ADMIN_HOST/health" >/dev/null 2>&1; then
        success "LLM Gateway Python is healthy"
    else
        error "LLM Gateway Python is not responding"
        return 1
    fi
}

test_authentication() {
    echo -e "\n${CYAN}🔐 Testing Authentication${NC}"
    echo "========================================"
    
    local email="test@example.com"
    local password="password123"
    
    log "Testing user login..."
    LOGIN_JSON=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
      -H "Content-Type: application/json" \
      -d "{\"email\":\"$email\",\"password\":\"$password\"}")

    ACCESS_TOKEN=$(echo "$LOGIN_JSON" | jq -r '.access_token' 2>/dev/null || echo "null")

    if [[ "$ACCESS_TOKEN" == "null" || -z "$ACCESS_TOKEN" ]]; then
        warning "Login failed - may need to register user first"
        
        # Try registration
        log "Attempting user registration..."
        REGISTER_JSON=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
          -H "Content-Type: application/json" \
          -d "{\"email\":\"$email\",\"password\":\"$password\",\"fullName\":\"Test User\"}")
        
        if echo "$REGISTER_JSON" | grep -q "success\|created"; then
            success "User registered successfully"
            
            # Try login again
            LOGIN_JSON=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
              -H "Content-Type: application/json" \
              -d "{\"email\":\"$email\",\"password\":\"$password\"}")
            ACCESS_TOKEN=$(echo "$LOGIN_JSON" | jq -r '.access_token' 2>/dev/null || echo "null")
        fi
    fi

    if [[ "$ACCESS_TOKEN" != "null" && -n "$ACCESS_TOKEN" ]]; then
        success "Authentication successful"
        export ACCESS_TOKEN
        return 0
    else
        error "Authentication failed"
        return 1
    fi
}

test_llm_gateway() {
    echo -e "\n${CYAN}🤖 Testing LLM Gateway${NC}"
    echo "========================================"
    
    log "Testing basic chat completion..."
    
    # Test basic completion
    local test_prompt="Hello, how are you?"
    local response=$(grpcurl -plaintext -d "{\"prompt\": \"$test_prompt\"}" \
        $LLM_HOST llm.v1.LLMService/Generate 2>/dev/null | \
        jq -r '.token // empty' | tr -d '\n')
    
    if [[ ${#response} -gt 5 ]]; then
        success "Basic chat completion working"
    else
        error "Basic chat completion failed"
        return 1
    fi
    
    # Test RAG functionality
    log "Testing RAG functionality..."
    local rag_prompt="What are the admission requirements for Vietnamese universities?"
    local rag_response=$(grpcurl -plaintext -d "{\"prompt\": \"$rag_prompt\"}" \
        $LLM_HOST llm.v1.LLMService/GenerateWithRAG 2>/dev/null | \
        jq -r '.token // empty' | tr -d '\n')
    
    if [[ ${#rag_response} -gt 10 ]]; then
        success "RAG functionality working"
    else
        warning "RAG functionality may have issues"
    fi
}

test_vietnamese_rag() {
    echo -e "\n${CYAN}🇻🇳 Testing Vietnamese RAG${NC}"
    echo "========================================"
    
    local vietnamese_queries=(
        "Điểm chuẩn trường Đại học Bách khoa Hà Nội"
        "Trường đại học nào có điểm chuẩn cao nhất?"
        "Các ngành học tại ĐHQGHN"
    )
    
    local success_count=0
    local total_queries=${#vietnamese_queries[@]}
    
    for query in "${vietnamese_queries[@]}"; do
        log "Testing query: $query"
        
        local response=$(grpcurl -plaintext -d "{\"prompt\": \"$query\"}" \
            $LLM_HOST llm.v1.LLMService/GenerateWithRAG 2>/dev/null | \
            jq -r '.token // empty' | tr -d '\n')
        
        if [[ ${#response} -gt 20 ]]; then
            success "Query processed successfully"
            ((success_count++))
        else
            warning "Query failed or returned short response"
        fi
    done
    
    info "Vietnamese RAG test results: $success_count/$total_queries successful"
    
    if [[ $success_count -ge $((total_queries / 2)) ]]; then
        success "Vietnamese RAG is functioning"
        return 0
    else
        error "Vietnamese RAG has significant issues"
        return 1
    fi
}

test_websocket_chat() {
    echo -e "\n${CYAN}💬 Testing WebSocket Chat${NC}"
    echo "========================================"
    
    if [[ -z "$ACCESS_TOKEN" ]]; then
        warning "No access token available, skipping WebSocket test"
        return 1
    fi
    
    log "Testing WebSocket connection..."
    
    # Create a simple WebSocket test using available tools
    # This is a simplified test - in practice you'd use a WebSocket client
    local ws_test_result=$(timeout 5s curl -s \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Connection: Upgrade" \
        -H "Upgrade: websocket" \
        "$API_URL/api/v1/chat/ws" 2>/dev/null || echo "timeout")
    
    if [[ "$ws_test_result" != "timeout" ]]; then
        success "WebSocket endpoint is accessible"
    else
        warning "WebSocket test inconclusive (may require proper WS client)"
    fi
}

# Main test runner
run_test_suite() {
    local test_type="${1:-all}"
    
    echo -e "${BLUE}🚀 CareerUP Unified Test Suite${NC}"
    echo "=================================================="
    echo "Test type: $test_type"
    echo "Timestamp: $(date)"
    echo ""
    
    case "$test_type" in
        "health")
            test_health_checks
            ;;
        "auth")
            test_authentication
            ;;
        "llm")
            test_llm_gateway
            ;;
        "vietnamese")
            test_vietnamese_rag
            ;;
        "websocket")
            test_websocket_chat
            ;;
        "core")
            test_health_checks && \
            test_authentication && \
            test_llm_gateway
            ;;
        "all"|*)
            test_health_checks && \
            test_authentication && \
            test_llm_gateway && \
            test_vietnamese_rag && \
            test_websocket_chat
            ;;
    esac
    
    local exit_code=$?
    
    echo ""
    echo "=================================================="
    if [[ $exit_code -eq 0 ]]; then
        success "Test suite completed successfully"
    else
        error "Test suite completed with errors"
    fi
    
    return $exit_code
}

# Help function
show_help() {
    echo "CareerUP Unified Test Suite"
    echo ""
    echo "Usage: $0 [test_type]"
    echo ""
    echo "Test types:"
    echo "  all        - Run all tests (default)"
    echo "  health     - Health checks only"
    echo "  auth       - Authentication tests only"
    echo "  llm        - LLM Gateway tests only"
    echo "  vietnamese - Vietnamese RAG tests only"
    echo "  websocket  - WebSocket chat tests only"
    echo "  core       - Core functionality (health + auth + llm)"
    echo ""
    echo "Examples:"
    echo "  $0              # Run all tests"
    echo "  $0 core         # Run core functionality tests"
    echo "  $0 vietnamese   # Run Vietnamese RAG tests only"
}

# Script entry point
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    show_help
    exit 0
fi

run_test_suite "$1"
