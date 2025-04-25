#!/bin/bash

# Test registration
echo "Testing registration..."
curl -X POST "http://localhost:8081/api/v1/auth/register?email=test@example.com&password=password123&firstName=Test&lastName=User"

# Test login
echo -e "\n\nTesting login..."
TOKEN=$(curl -s -X POST "http://localhost:8081/api/v1/auth/login?email=test@example.com&password=password123")
echo "Token: $TOKEN"

# Test token validation
echo -e "\n\nTesting token validation..."
curl -X POST "http://localhost:8081/api/v1/auth/validate?token=$TOKEN"

# Test get current user
echo -e "\n\nTesting get current user..."
curl -X GET "http://localhost:8081/api/v1/auth/me?email=test@example.com"

# Test update user
echo -e "\n\nTesting update user..."
curl -X PUT "http://localhost:8081/api/v1/auth/me?email=test@example.com&firstName=Updated&lastName=Name&hometown=New%20York&interests=AI,Machine%20Learning" 