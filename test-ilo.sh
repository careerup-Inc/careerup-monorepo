#!/bin/zsh

API_URL="http://localhost:8080"
EMAIL="test@example.com"
PASSWORD="password123"

# 0. Login and get access token
LOGIN_JSON=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"'$EMAIL'","password":"'$PASSWORD'"}')
ACCESS_TOKEN=$(echo "$LOGIN_JSON" | jq -r '.access_token')

if [[ "$ACCESS_TOKEN" == "null" || -z "$ACCESS_TOKEN" ]]; then
  echo "Login failed! Response: $LOGIN_JSON"
  exit 1
fi

# 1. Fetch ILO questions
QUESTIONS_JSON=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" "$API_URL/api/v1/ilo/test")

echo "QUESTIONS_JSON:"
echo "$QUESTIONS_JSON" | jq

# 2. Build answers array directly with jq
ANSWERS_JSON=$(echo "$QUESTIONS_JSON" | jq '[.questions[] | {question_id: .id, question_number: .question_number, selected_option: (1 + (.question_number % 4))}]')

# Create sample result data with domain information
RESULT_DATA='{
  "domains": {
    "LANG": {"raw_score": 36, "percent": 75.0, "level": "Strong", "rank": 2},
    "LOGIC": {"raw_score": 42, "percent": 87.5, "level": "Very Strong", "rank": 1},
    "DESIGN": {"raw_score": 30, "percent": 62.5, "level": "Average", "rank": 3},
    "PEOPLE": {"raw_score": 24, "percent": 50.0, "level": "Average", "rank": 4},
    "MECH": {"raw_score": 18, "percent": 37.5, "level": "Weak", "rank": 5}
  },
  "top_domains": ["LOGIC", "LANG", "DESIGN"],
  "suggested_careers": ["Software Engineer", "Data Scientist", "Business Analyst"]
}'

echo "Submitting answers:"
echo "$ANSWERS_JSON" | jq

# 3. Submit answers - SIMPLIFIED VERSION
RESULT=$(curl -s -X POST "$API_URL/api/v1/ilo/result" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "answers": '"$ANSWERS_JSON"'
  }')

# Clean any control characters before processing with jq
CLEAN_RESULT=$(echo "$RESULT" | tr -d '\000-\037')

echo "Raw result:"
echo "$CLEAN_RESULT"

echo "Submit Result (parsed):"
echo "$CLEAN_RESULT" | jq 2>/dev/null || echo "Could not parse JSON response"

# Extract the result ID, with added robustness
RESULT_ID=$(echo "$CLEAN_RESULT" | grep -o '"id":"[^"]*"' | head -1 | cut -d':' -f2 | tr -d '"')

if [[ -z "$RESULT_ID" ]]; then
  echo "Failed to get result ID"
  exit 1
fi

echo "Success! Result ID: $RESULT_ID"

echo -e "\nWaiting 2 seconds before retrieving the result..."
sleep 2

# 5. Get the saved result by ID
echo -e "\nRetrieving result by ID ($RESULT_ID):"
GET_RESULT=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" "$API_URL/api/v1/ilo/result/$RESULT_ID")
echo "$GET_RESULT" | tr -d '\000-\037' | jq 2>/dev/null || echo "Failed to parse result JSON"

# 6. Get all results for the user
echo -e "\nRetrieving all results for the user:"
ALL_RESULTS=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" "$API_URL/api/v1/ilo/results")

# Clean any control characters before processing with jq
CLEAN_ALL_RESULTS=$(echo "$ALL_RESULTS" | tr -d '\000-\037')

# Try to parse the clean result
echo "$CLEAN_ALL_RESULTS" | jq 2>/dev/null || echo "Raw response: $CLEAN_ALL_RESULTS"