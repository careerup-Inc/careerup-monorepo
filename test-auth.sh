#!/bin/zsh

# =============================================================================
# CareerUP Authentication Service Test
# =============================================================================

set -e  # Exit on error

# Configuration
API_URL="http://localhost:8080"  # Updated to match other scripts
AUTH_URL="http://localhost:8081"  # Keep separate for auth-core direct testing
EMAIL="test@example.com" # Replace if account already exist
PASSWORD="password123"

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
    return 1
}

# Helper function to clean JSON response
clean_json() {
    echo "$1" | tr -d '\000-\037'
}

# Test user registration
test_registration() {
    log "Testing user registration..."
    
    # Create registration payload
    REGISTER_PAYLOAD='{
        "email": "'$EMAIL'",
        "password": "'$PASSWORD'",
        "firstName": "Test",
        "lastName": "User"
    }'
    
    REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "$REGISTER_PAYLOAD")
    
    CLEAN_REGISTER=$(clean_json "$REGISTER_RESPONSE")
    
    # Check if registration was successful
    if echo "$CLEAN_REGISTER" | jq -e '.email' > /dev/null 2>&1; then
        success "Registration successful"
        echo "User details: $(echo "$CLEAN_REGISTER" | jq -r '.firstName + " " + .lastName + " (" + .email + ")"')"
    else
        # Registration might fail if user already exists - check error message
        if echo "$CLEAN_REGISTER" | grep -q "already exists\|already registered"; then
            warning "User already exists - continuing with login test"
        else
            error "Registration failed: $CLEAN_REGISTER"
        fi
    fi
    
    echo ""
}

# Test user login
test_login() {
    log "Testing user login..."
    
    LOGIN_PAYLOAD='{
        "email": "'$EMAIL'",
        "password": "'$PASSWORD'"
    }'
    
    LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "$LOGIN_PAYLOAD")
    
    CLEAN_LOGIN=$(clean_json "$LOGIN_RESPONSE")
    
    # Extract tokens
    ACCESS_TOKEN=$(echo "$CLEAN_LOGIN" | jq -r '.access_token')
    REFRESH_TOKEN=$(echo "$CLEAN_LOGIN" | jq -r '.refresh_token')
    EXPIRES_IN=$(echo "$CLEAN_LOGIN" | jq -r '.expires_in')
    
    if [[ "$ACCESS_TOKEN" != "null" && -n "$ACCESS_TOKEN" ]]; then
        success "Login successful"
        echo "Access token: ${ACCESS_TOKEN:0:20}..."
        echo "Refresh token: ${REFRESH_TOKEN:0:20}..."
        echo "Expires in: $EXPIRES_IN seconds"
        
        # Export for use in other tests
        export ACCESS_TOKEN REFRESH_TOKEN
    else
        error "Login failed: $CLEAN_LOGIN"
    fi
    
    echo ""
}

# Test token validation
test_token_validation() {
    log "Testing token validation..."
    
    if [[ -z "$ACCESS_TOKEN" ]]; then
        error "No access token available for validation"
        return 1
    fi
    
    VALIDATE_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/auth/validate" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    CLEAN_VALIDATE=$(clean_json "$VALIDATE_RESPONSE")
    
    if echo "$CLEAN_VALIDATE" | jq -e '.email' > /dev/null 2>&1; then
        success "Token validation successful"
        USER_INFO=$(echo "$CLEAN_VALIDATE" | jq -r '.firstName + " " + .lastName + " (" + .email + ")"')
        echo "Validated user: $USER_INFO"
        echo "Active status: $(echo "$CLEAN_VALIDATE" | jq -r '.isActive')"
    else
        error "Token validation failed: $CLEAN_VALIDATE"
    fi
    
    echo ""
}

# Test refresh token
test_refresh_token() {
    log "Testing refresh token..."
    
    if [[ -z "$REFRESH_TOKEN" ]]; then
        error "No refresh token available"
        return 1
    fi
    
    REFRESH_PAYLOAD='{
        "refresh_token": "'$REFRESH_TOKEN'"
    }'
    
    REFRESH_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/refresh" \
        -H "Content-Type: application/json" \
        -d "$REFRESH_PAYLOAD")
    
    CLEAN_REFRESH=$(clean_json "$REFRESH_RESPONSE")
    
    NEW_ACCESS_TOKEN=$(echo "$CLEAN_REFRESH" | jq -r '.access_token')
    
    if [[ "$NEW_ACCESS_TOKEN" != "null" && -n "$NEW_ACCESS_TOKEN" ]]; then
        success "Token refresh successful"
        echo "New access token: ${NEW_ACCESS_TOKEN:0:20}..."
        export ACCESS_TOKEN="$NEW_ACCESS_TOKEN"
    else
        error "Token refresh failed: $CLEAN_REFRESH"
    fi
    
    echo ""
}

# Test getting current user profile
test_get_profile() {
    log "Testing get user profile..."
    
    if [[ -z "$ACCESS_TOKEN" ]]; then
        error "No access token available"
        return 1
    fi
    
    PROFILE_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/user/me" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    CLEAN_PROFILE=$(clean_json "$PROFILE_RESPONSE")
    
    if echo "$CLEAN_PROFILE" | jq -e '.email' > /dev/null 2>&1; then
        success "Profile retrieval successful"
        echo "Profile: $(echo "$CLEAN_PROFILE" | jq)"
    else
        error "Profile retrieval failed: $CLEAN_PROFILE"
    fi
    
    echo ""
}

# Test updating user profile
test_update_profile() {
    log "Testing update user profile..."
    
    if [[ -z "$ACCESS_TOKEN" ]]; then
        error "No access token available"
        return 1
    fi
    
    UPDATE_PAYLOAD='{
        "firstName": "Updated",
        "lastName": "Name",
        "hometown": "New York",
        "interests": ["AI", "Machine Learning", "Career Development"]
    }'
    
    UPDATE_RESPONSE=$(curl -s -X PUT "$API_URL/api/v1/profile" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d "$UPDATE_PAYLOAD")
    
    CLEAN_UPDATE=$(clean_json "$UPDATE_RESPONSE")
    
    if echo "$CLEAN_UPDATE" | jq -e '.email' > /dev/null 2>&1; then
        success "Profile update successful"
        echo "Updated profile: $(echo "$CLEAN_UPDATE" | jq -r '.firstName + " " + .lastName')"
        echo "Hometown: $(echo "$CLEAN_UPDATE" | jq -r '.hometown // "Not set"')"
        echo "Interests: $(echo "$CLEAN_UPDATE" | jq -r '.interests // [] | join(", ")')"
    else
        error "Profile update failed: $CLEAN_UPDATE"
    fi
    
    echo ""
}

# Test invalid token scenarios
test_invalid_scenarios() {
    log "Testing invalid token scenarios..."
    
    # Test with invalid token
    INVALID_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/auth/validate" \
        -H "Authorization: Bearer invalid_token_12345")
    
    if echo "$INVALID_RESPONSE" | grep -q "401\|Unauthorized\|Invalid"; then
        success "Invalid token properly rejected"
    else
        warning "Invalid token test may not be working as expected"
    fi
    
    # Test with no token
    NO_TOKEN_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/auth/validate")
    
    if echo "$NO_TOKEN_RESPONSE" | grep -q "401\|Unauthorized"; then
        success "Missing token properly rejected"
    else
        warning "Missing token test may not be working as expected"
    fi
    
    echo ""
}

# Test direct auth-core service (if different from API gateway)
test_direct_auth_core() {
    log "Testing direct auth-core service..."
    
    if [[ "$AUTH_URL" != "$API_URL" ]]; then
        # Test direct connection to auth-core
        DIRECT_RESPONSE=$(curl -s -X GET "$AUTH_URL/actuator/health" 2>/dev/null || echo "failed")
        
        if echo "$DIRECT_RESPONSE" | grep -q "UP\|status.*up"; then
            success "Auth-core service is healthy"
        else
            warning "Direct auth-core connection failed or service unhealthy"
        fi
    else
        success "Using unified API endpoint"
    fi
    
    echo ""
}

# Main function
main() {
    echo "========================================================================================="
    echo "🔐 CareerUP Authentication Service Test"
    echo "========================================================================================="
    echo ""
    
    # Test service health first
    test_direct_auth_core
    
    # Core authentication flow
    test_registration
    test_login
    test_token_validation
    test_refresh_token
    
    # Profile management
    test_get_profile
    test_update_profile
    
    # Security tests
    test_invalid_scenarios
    
    echo "========================================================================================="
    echo "✅ Authentication tests completed!"
    echo "========================================================================================="
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    error "jq is required but not installed. Install it with: brew install jq"
    exit 1
fi

# Run the main function
main "$@"