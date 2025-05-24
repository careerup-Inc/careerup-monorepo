#!/bin/bash

# Setup script for RAG (Retrieval-Augmented Generation) system
# This script creates collections and ingests documents into the vector database

set -e

# Configuration
LLM_ADMIN_BASE_URL="http://localhost:8090"
COLLECTION_NAME="academy"

echo "🚀 Setting up RAG system for CareerUP..."
echo "Admin URL: $LLM_ADMIN_BASE_URL"
echo "Collection: $COLLECTION_NAME"

# Function to check if LLM gateway admin is ready
check_admin_ready() {
    echo "⏳ Checking if LLM Gateway admin server is ready..."
    max_attempts=30
    attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$LLM_ADMIN_BASE_URL/health" > /dev/null 2>&1; then
            echo "✅ LLM Gateway admin server is ready!"
            return 0
        fi
        echo "   Attempt $attempt/$max_attempts: Admin server not ready, waiting..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "❌ LLM Gateway admin server not ready after $max_attempts attempts"
    return 1
}

# Function to create collection
create_collection() {
    echo "📚 Creating collection '$COLLECTION_NAME'..."
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$LLM_ADMIN_BASE_URL/admin/collections" \
        -H "Content-Type: application/json" \
        -d "{
            \"collection_name\": \"$COLLECTION_NAME\",
            \"metadata\": {
                \"description\": \"CareerUP educational content and career guidance\",
                \"created_by\": \"setup-script\",
                \"version\": \"1.0\"
            }
        }")
    
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 201 ]; then
        echo "✅ Collection '$COLLECTION_NAME' created successfully!"
        echo "   Response: $response_body"
    else
        echo "⚠️  Collection creation response (HTTP $http_code): $response_body"
        # Don't fail if collection already exists
        if echo "$response_body" | grep -q "already exists"; then
            echo "   Collection already exists, continuing..."
        fi
    fi
}

# Function to ingest a document
ingest_document() {
    local doc_id="$1"
    local content="$2"
    local title="$3"
    
    echo "📄 Ingesting document: $title"
    
    response=$(curl -s -w "\n%{http_code}" -X POST "$LLM_ADMIN_BASE_URL/admin/ingest-document" \
        -H "Content-Type: application/json" \
        -d "{
            \"document_id\": \"$doc_id\",
            \"collection\": \"$COLLECTION_NAME\",
            \"content\": $(echo "$content" | jq -R .),
            \"metadata\": {
                \"title\": \"$title\",
                \"type\": \"educational_content\",
                \"ingested_at\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"
            }
        }")
    
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 201 ]; then
        echo "   ✅ Document ingested successfully!"
        chunks=$(echo "$response_body" | jq -r '.chunks_created // 0')
        echo "   📊 Created $chunks chunks"
    else
        echo "   ❌ Failed to ingest document (HTTP $http_code): $response_body"
        return 1
    fi
}

# Function to list collections
list_collections() {
    echo "📋 Listing all collections..."
    
    response=$(curl -s -w "\n%{http_code}" -X GET "$LLM_ADMIN_BASE_URL/admin/collections")
    
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" -eq 200 ]; then
        echo "✅ Collections retrieved successfully!"
        echo "$response_body" | jq -r '.collections[]? | "   📚 \(.name) (\(.document_count // 0) documents)"'
    else
        echo "❌ Failed to list collections (HTTP $http_code): $response_body"
    fi
}

# Sample educational content for CareerUP
ingest_sample_content() {
    echo "📚 Ingesting sample educational content..."
    
    # Career guidance content
    ingest_document "career_guidance_001" \
        "Career planning is essential for professional success. Start by identifying your interests, skills, and values. Research different career paths and industries that align with your goals. Network with professionals in your field of interest. Consider pursuing relevant education, certifications, or training programs. Set short-term and long-term career goals and regularly review your progress." \
        "Career Planning Fundamentals"
    
    # University admissions guide
    ingest_document "university_guide_001" \
        "University admissions in Vietnam are highly competitive. Students typically take the National High School Examination (Thi THPT Quoc Gia) which covers Mathematics, Literature, and Foreign Language as core subjects, plus specialized subjects based on their chosen field. Top universities like Vietnam National University, Hanoi University of Science and Technology, and Ho Chi Minh City University of Technology have high admission scores. Preparation should start early with consistent study habits and practice tests." \
        "Vietnamese University Admission Guide"
    
    # Study tips and academic success
    ingest_document "study_tips_001" \
        "Effective studying requires good time management and study techniques. Use the Pomodoro Technique for focused study sessions. Create a distraction-free study environment. Practice active recall by testing yourself regularly. Form study groups to discuss concepts with peers. Take regular breaks and maintain a healthy sleep schedule. Organize your notes and materials systematically." \
        "Effective Study Techniques"
    
    # Technology career paths
    ingest_document "tech_careers_001" \
        "Technology careers offer diverse opportunities including software development, data science, cybersecurity, AI/ML engineering, and product management. Key programming languages include Python, JavaScript, Java, and C++. Cloud platforms like AWS, Azure, and Google Cloud are increasingly important. Continuous learning is essential in tech through online courses, coding bootcamps, and hands-on projects. Building a portfolio of projects demonstrates practical skills to employers." \
        "Technology Career Paths"
    
    # Soft skills development
    ingest_document "soft_skills_001" \
        "Soft skills are crucial for career success and include communication, teamwork, problem-solving, leadership, and emotional intelligence. Develop communication skills through practice and feedback. Build teamwork abilities by collaborating on group projects. Enhance problem-solving by approaching challenges systematically. Practice leadership by taking initiative and guiding others. Improve emotional intelligence by understanding your emotions and those of others." \
        "Essential Soft Skills for Career Success"
    
    # Interview preparation
    ingest_document "interview_prep_001" \
        "Job interview preparation involves researching the company, practicing common questions, and preparing thoughtful questions to ask. Use the STAR method (Situation, Task, Action, Result) to structure behavioral question responses. Prepare specific examples that demonstrate your skills and achievements. Dress professionally and arrive early. Follow up with a thank-you email within 24 hours of the interview." \
        "Job Interview Preparation Guide"
    
    echo "✅ Sample content ingestion completed!"
}

# Main execution
main() {
    echo "🎯 Starting RAG system setup..."
    
    # Check if admin server is ready
    if ! check_admin_ready; then
        echo "❌ Cannot proceed without admin server. Please ensure LLM Gateway is running."
        echo "💡 Try: docker-compose up llm-gateway"
        exit 1
    fi
    
    # Create collection
    create_collection
    
    # Ingest sample content
    ingest_sample_content
    
    # List final state
    echo ""
    list_collections
    
    echo ""
    echo "🎉 RAG system setup completed successfully!"
    echo "📚 Collection '$COLLECTION_NAME' is ready for use"
    echo "🔍 You can now test RAG-enabled chat at: http://localhost:8080"
    echo ""
    echo "🛠️  Available admin endpoints:"
    echo "   📋 List collections: GET $LLM_ADMIN_BASE_URL/admin/collections"
    echo "   📄 Ingest document: POST $LLM_ADMIN_BASE_URL/admin/ingest-document"
    echo "   ❤️  Health check: GET $LLM_ADMIN_BASE_URL/health"
}

# Check if required tools are available
if ! command -v curl &> /dev/null; then
    echo "❌ curl is required but not installed."
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "❌ jq is required but not installed."
    echo "💡 Install with: brew install jq (macOS) or sudo apt-get install jq (Ubuntu)"
    exit 1
fi

# Run main function
main "$@"
