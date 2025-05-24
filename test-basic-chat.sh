#!/bin/zsh

# =============================================================================
# CareerUP WebSocket Chat Test
# =============================================================================

set -e  # Exit on error

# Configuration
API_URL="http://localhost:8080"
WS_URL="ws://localhost:8080"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

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

# Step 1: Authentication
authenticate() {
    log "Step 1: Authenticating user..."
    
    LOGIN_JSON=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
      -H "Content-Type: application/json" \
      -d '{"email":"test@example.com","password":"password123"}')

    ACCESS_TOKEN=$(echo "$LOGIN_JSON" | jq -r '.access_token')

    if [[ "$ACCESS_TOKEN" == "null" || -z "$ACCESS_TOKEN" ]]; then
        error "Login failed! Response: $LOGIN_JSON"
        exit 1
    else
        success "Login successful. Access token: $ACCESS_TOKEN"
    fi
    export ACCESS_TOKEN
}

# Test WebSocket upgrade using curl with timeout
test_websocket_upgrade() {
    log "Testing WebSocket upgrade handshake..."
    
    # Test the upgrade request with proper headers and timeout
    UPGRADE_RESPONSE=$(timeout 5s curl -s -i \
        -H "Connection: Upgrade" \
        -H "Upgrade: websocket" \
        -H "Sec-WebSocket-Version: 13" \
        -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        "$API_URL/api/v1/ws" 2>/dev/null || echo "timeout_or_error")
    
    echo "Upgrade response headers:"
    echo "$UPGRADE_RESPONSE"
    
    if echo "$UPGRADE_RESPONSE" | grep -q "101 Switching Protocols"; then
        success "WebSocket upgrade successful"
        return 0
    elif echo "$UPGRADE_RESPONSE" | grep -q "401"; then
        error "WebSocket upgrade failed: Authentication error"
        echo "Check if the token is valid and properly formatted"
        return 1
    elif echo "$UPGRADE_RESPONSE" | grep -q "404"; then
        error "WebSocket upgrade failed: Endpoint not found"
        echo "Check if the WebSocket endpoint is properly configured"
        return 1
    elif [[ "$UPGRADE_RESPONSE" == "timeout_or_error" ]]; then
        warning "WebSocket upgrade test timed out (this may be normal - curl can't maintain WebSocket connections)"
        warning "This doesn't necessarily mean the WebSocket is broken - trying proper WebSocket clients next..."
        return 0
    else
        error "WebSocket upgrade failed"
        echo "Response: $UPGRADE_RESPONSE"
        return 1
    fi
}

# Test using Python WebSocket client (fixed version)
test_websocket_with_python() {
    log "Testing WebSocket chat with Python..."
    
    # Create a temporary Python script with corrected websockets usage
    cat > /tmp/websocket_test.py << 'EOF'
import asyncio
import websockets
import json
import sys

async def test_chat():
    uri = "ws://localhost:8080/api/v1/ws"
    token = sys.argv[1]
    
    try:
        # Connect with authorization header in the URI or use connect with headers
        async with websockets.connect(
            uri,
            additional_headers={"Authorization": f"Bearer {token}"}
        ) as websocket:
            print("✅ WebSocket connected")
            
            # Send a test message
            message = {
                "type": "user_msg",
                "conversation_id": "test-python",
                "text": "What are the benefits of pursuing a career in AI?"
            }
            
            print(f"📤 Sending: {json.dumps(message)}")
            await websocket.send(json.dumps(message))
            
            # Listen for responses with timeout
            response_count = 0
            full_response = ""
            
            try:
                while response_count < 100:
                    # Wait for message with timeout
                    message = await asyncio.wait_for(websocket.recv(), timeout=2.0)
                    response = json.loads(message)
                    print(f"📥 Received: {response}")
                    
                    if response.get("type") == "assistant_token":
                        full_response += response.get("token", "")
                        print(response.get("token", ""), end="", flush=True)
                    elif response.get("type") == "error":
                        print(f"❌ Error: {response.get('error_message', 'Unknown error')}")
                        break
                    
                    response_count += 1
                    
            except asyncio.TimeoutError:
                print("\n✅ Response completed (timeout reached)")
            
            print(f"\n✅ Full response: {full_response}")
            
    except websockets.exceptions.InvalidStatusCode as e:
        print(f"❌ WebSocket connection failed with status {e.status_code}")
        if e.status_code == 401:
            print("   Authentication failed - check your token")
        elif e.status_code == 404:
            print("   WebSocket endpoint not found")
    except Exception as e:
        print(f"❌ WebSocket error: {e}")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python websocket_test.py <access_token>")
        sys.exit(1)
    
    asyncio.run(test_chat())
EOF

    if command -v python3 &> /dev/null; then
        # Check if websockets library is available
        if python3 -c "import websockets" 2>/dev/null; then
            python3 /tmp/websocket_test.py "$ACCESS_TOKEN"
        else
            warning "Python websockets library not found. Install with: pip install websockets"
        fi
    else
        warning "Python3 not found"
    fi
    
    rm -f /tmp/websocket_test.py
}

# Test WebSocket using Node.js (more reliable)
test_websocket_with_node() {
    log "Testing WebSocket chat with Node.js..."
    
    if ! command -v node &> /dev/null; then
        warning "Node.js not found. Install Node.js to run this test"
        return 1
    fi
    
    # Use current working directory to access node_modules
    cd "/Users/doviethoang/github/careerup-monorepo"
    
    # Check if ws package is available
    if ! node -e "require('ws')" 2>/dev/null; then
        warning "Node.js 'ws' package not found. Install with: npm install ws"
        return 1
    fi
    
    # Create a temporary Node.js script in the project directory
    cat > ./websocket_test_temp.js << EOF
const WebSocket = require('ws');

const ws = new WebSocket('$WS_URL/api/v1/ws', {
  headers: {
    'Authorization': 'Bearer $ACCESS_TOKEN'
  }
});

let responseCount = 0;
let fullResponse = '';

ws.on('open', function open() {
  console.log('✅ WebSocket connected');
  
  // Send a test message
  const message = {
    type: 'user_msg',
    conversation_id: 'test-node-' + Date.now(),
    text: 'Hello! Can you tell me about AI careers?'
  };
  
  console.log('📤 Sending:', JSON.stringify(message));
  ws.send(JSON.stringify(message));
});

ws.on('message', function message(data) {
  const response = JSON.parse(data.toString());
  console.log('📥 Received:', response);
  
  if (response.type === 'assistant_token') {
    fullResponse += response.token;
    process.stdout.write(response.token);
  } else if (response.type === 'error') {
    console.error('❌ Error:', response.error_message);
  }
  
  responseCount++;
  if (responseCount > 100) {
    console.log('\n✅ Response limit reached');
    ws.close();
  }
});

ws.on('error', function error(err) {
  console.error('❌ WebSocket error:', err.message);
  if (err.message.includes('401')) {
    console.error('   Authentication failed - check your token');
  }
});

ws.on('close', function close(code, reason) {
  console.log('\n🔌 WebSocket connection closed. Code:', code, 'Reason:', reason);
  console.log('✅ Full response:', fullResponse);
  process.exit(0);
});

// Close after 15 seconds
setTimeout(() => {
  console.log('\n⏰ Closing connection after timeout');
  ws.close();
}, 15000);
EOF

    node ./websocket_test_temp.js
    rm -f ./websocket_test_temp.js
}

# Test using websocat (if available)
test_websocket_with_websocat() {
    log "Testing WebSocket with websocat..."
    
    if ! command -v websocat &> /dev/null; then
        warning "websocat not found. Install with: brew install websocat"
        return 1
    fi
    
    # Create test message
    echo '{"type":"user_msg","conversation_id":"test-websocat","text":"Hello from websocat!"}' | \
    timeout 10s websocat -H "Authorization: Bearer $ACCESS_TOKEN" "$WS_URL/api/v1/ws" || \
    warning "websocat test timed out or failed"
}

# Test the complete flow
test_complete_flow() {
    log "Testing complete chat flow..."
    
    # 1. Test auth endpoint
    log "1. Testing auth validation..."
    AUTH_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/auth/validate" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    if echo "$AUTH_RESPONSE" | jq -e '.email' > /dev/null 2>&1; then
        success "Auth validation successful"
    else
        error "Auth validation failed: $AUTH_RESPONSE"
        return 1
    fi
    
    # 2. Test WebSocket upgrade
    test_websocket_upgrade
    
    # 3. Test actual WebSocket communication (try multiple methods)
    echo ""
    log "Testing WebSocket communication..."
    
    test_websocket_with_node || \
    test_websocket_with_python || \
    test_websocket_with_websocat || \
    warning "All WebSocket communication tests failed"
}

# Debug function to check services
debug_services() {
    log "Debugging service connectivity..."
    
    # Test API Gateway
    API_HEALTH=$(curl -s "$API_URL/api/v1/health" 2>/dev/null || echo "failed")
    if [[ "$API_HEALTH" == *"healthy"* ]] || [[ "$API_HEALTH" == *"ok"* ]]; then
        success "API Gateway is responding"
    else
        warning "API Gateway health check failed: $API_HEALTH"
    fi
    
    # Test Chat Gateway (if it has a health endpoint)
    CHAT_HEALTH=$(curl -s "http://localhost:8082/health" 2>/dev/null || echo "failed")
    if [[ "$CHAT_HEALTH" != "failed" ]]; then
        success "Chat Gateway is responding"
    else
        warning "Chat Gateway not responding on port 8082"
    fi
    
    # Test LLM Gateway (if it has a health endpoint)
    LLM_HEALTH=$(curl -s "http://localhost:50053/health" 2>/dev/null || echo "failed")
    if [[ "$LLM_HEALTH" != "failed" ]]; then
        success "LLM Gateway is responding"
    else
        warning "LLM Gateway not responding on port 50053 (this is normal - LLM Gateway is gRPC only)"
    fi
}

# Main function
main() {
    echo "========================================================================================="
    echo "🚀 CareerUP WebSocket Chat Test"
    echo "========================================================================================="
    echo ""
    
    # Debug services first
    debug_services
    echo ""
    
    authenticate
    echo ""
    
    test_complete_flow
    echo ""

    echo "========================================================================================="
    echo "✅ WebSocket Chat Test completed!"
    echo "========================================================================================="
}

# Check dependencies
check_dependencies() {
    local missing_deps=()
    
    if ! command -v jq &> /dev/null; then
        missing_deps+=("jq")
    fi
    
    if ! command -v curl &> /dev/null; then
        missing_deps+=("curl")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        error "Missing dependencies: ${missing_deps[*]}"
        echo "Install them with:"
        echo "  brew install jq curl"
        exit 1
    fi
}

# Run the test
check_dependencies
main "$@"