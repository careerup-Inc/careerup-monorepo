package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"

	pbllm "github.com/careerup-Inc/careerup-monorepo/proto/llm/v1"
	"github.com/careerup-Inc/careerup-monorepo/services/llm-gateway/internal/pinecone"
)

// QueryRoute represents the routing decision for a query
type QueryRoute string

const (
	RouteVectorStore QueryRoute = "vectorstore"
	RouteWebSearch   QueryRoute = "web_search"
	RoutePureLLM     QueryRoute = "pure_llm"
)

// RAGState represents the state in our RAG workflow
type RAGState struct {
	Question   string
	Documents  []schema.Document
	Generation string
	Route      QueryRoute
	Iteration  int
	MaxRetries int
}

// WebSearchResult represents a web search result
type WebSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type RAGConfig struct {
	ChunkSize        int
	ChunkOverlap     int
	RetrievalTopK    int
	Temperature      float64
	MaxTokens        int
	MaxRetries       int
	WebSearchEnabled bool
	WebSearchAPIKey  string
	WebSearchBaseURL string
}

// DefaultRAGConfig returns default RAG configuration
func DefaultRAGConfig() RAGConfig {
	return RAGConfig{
		ChunkSize:        1000,
		ChunkOverlap:     200,
		RetrievalTopK:    5,
		Temperature:      0.7,
		MaxTokens:        1000,
		MaxRetries:       3,
		WebSearchEnabled: true,
		WebSearchAPIKey:  os.Getenv("TAVILY_API_KEY"),
		WebSearchBaseURL: "https://api.tavily.com/search",
	}
}

// LLMServiceImpl implements the LLMService gRPC interface.
type LLMServiceImpl struct {
	pbllm.UnimplementedLLMServiceServer
	llm            llms.Model                      // LangChainGo LLM interface
	embedder       embeddings.Embedder             // Embeddings model
	pineconeClient *pinecone.Client                // Pinecone client
	collections    map[string]*pinecone.Collection // Collections metadata
	ragConfig      RAGConfig                       // RAG configuration
}

// NewLLMService creates a new LLMService implementation using langchaingo.
func NewLLMService() (*LLMServiceImpl, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// Check for Pinecone API key
	pineconeAPIKey := os.Getenv("PINECONE_API_KEY")
	if pineconeAPIKey == "" {
		return nil, fmt.Errorf("PINECONE_API_KEY environment variable not set")
	}

	// Initialize langchaingo OpenAI client
	llm, err := openai.New(
		openai.WithModel("gpt-4o"),
		openai.WithToken(apiKey),
	)
	if err != nil {
		log.Printf("Failed to initialize langchaingo OpenAI client: %v", err)
		return nil, err
	}

	// Choose embedding model based on environment variable
	embeddingModel := os.Getenv("EMBEDDING_MODEL")
	if embeddingModel == "" {
		embeddingModel = "llama"
	}

	var embedder embeddings.Embedder

	// Use OpenAI embeddings
	log.Printf("Initializing OpenAI embeddings...")
	embedder, err = embeddings.NewEmbedder(llm)
	if err != nil {
		log.Printf("Failed to initialize OpenAI embeddings: %v", err)
		return nil, fmt.Errorf("failed to initialize embeddings: %v", err)
	}

	vectorStores := make(map[string]*pinecone.Collection)

	// Initialize Pinecone client
	pineconeClient, err := pinecone.NewWithAPIKey(pineconeAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Pinecone client: %v", err)
	}

	return &LLMServiceImpl{
		llm:            llm,
		embedder:       embedder,
		pineconeClient: pineconeClient,
		collections:    vectorStores,
		ragConfig:      DefaultRAGConfig(),
	}, nil
}

// retrieveDocuments retrieves relevant documents from the vector store
func (s *LLMServiceImpl) retrieveDocuments(ctx context.Context, query string, collection string) ([]schema.Document, error) {
	// Get collection metadata
	_, ok := s.collections[collection]
	if !ok {
		log.Printf("Collection %s not found, returning empty documents", collection)
		return []schema.Document{}, nil
	}

	// Generate embedding for the query
	queryEmbedding, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		log.Printf("Failed to generate query embedding: %v", err)
		return []schema.Document{}, nil
	}

	// Convert embedding to float32 slice
	queryEmbeddingFloat32 := make([]float32, len(queryEmbedding))
	for i, v := range queryEmbedding {
		queryEmbeddingFloat32[i] = float32(v)
	}

	// Query Pinecone for similar documents
	queryReq := pinecone.QueryRequest{
		QueryEmbeddings: [][]float32{queryEmbeddingFloat32},
		NResults:        5, // Get top 5 most similar documents
		Include:         []string{"metadatas"},
	}

	resp, err := s.pineconeClient.QueryDocuments(collection, queryReq)
	if err != nil {
		log.Printf("Failed to query Pinecone: %v", err)
		return []schema.Document{}, nil
	}

	// Convert Pinecone results to schema.Document
	var documents []schema.Document
	if len(resp.Results) > 0 {
		results := resp.Results[0] // First query results
		for _, result := range results {
			// Convert metadata map[string]string to map[string]any
			metadata := make(map[string]any)
			for k, v := range result.Metadata {
				metadata[k] = v
			}

			doc := schema.Document{
				PageContent: result.Content,
				Metadata:    metadata,
			}
			documents = append(documents, doc)
		}
	}

	log.Printf("Retrieved %d documents for query: %s", len(documents), query)
	return documents, nil
}

// gradeDocumentRelevance checks if a document is relevant to the query
func (s *LLMServiceImpl) gradeDocumentRelevance(ctx context.Context, doc schema.Document, query string) (bool, error) {
	prompt := fmt.Sprintf(`You are a grader assessing relevance of a retrieved document to a user question.
If the document contains keyword(s) or semantic meaning related to the question, grade it as relevant.
Give a binary score 'yes' or 'no' to indicate whether the document is relevant to the question.

Retrieved document:
%s

User question: %s

Relevant (yes/no):`, doc.PageContent, query)

	response, err := llms.GenerateFromSinglePrompt(ctx, s.llm, prompt,
		llms.WithTemperature(0),
		llms.WithMaxTokens(10),
	)
	if err != nil {
		return false, err
	}

	// Simple check for positive response
	answer := strings.ToLower(strings.TrimSpace(response))
	return strings.Contains(answer, "yes"), nil
}

// checkHallucination verifies if the response is grounded in the documents
func (s *LLMServiceImpl) checkHallucination(ctx context.Context, response string, documents []schema.Document) (bool, error) {
	// Combine documents content
	var docsContent strings.Builder
	for _, doc := range documents {
		docsContent.WriteString(doc.PageContent)
		docsContent.WriteString("\n\n")
	}

	prompt := fmt.Sprintf(`You are a grader assessing whether an LLM generation is grounded in / supported by a set of retrieved facts.
Give a binary score 'yes' or 'no'. 'Yes' means that the answer is grounded in / supported by the set of facts.

Set of facts:
%s

LLM generation: %s

Grounded (yes/no):`, docsContent.String(), response)

	hallucinationCheck, err := llms.GenerateFromSinglePrompt(ctx, s.llm, prompt,
		llms.WithTemperature(0),
		llms.WithMaxTokens(10),
	)
	if err != nil {
		return false, err
	}

	// Check if grounded (not hallucinating)
	answer := strings.ToLower(strings.TrimSpace(hallucinationCheck))
	return strings.Contains(answer, "yes"), nil
}

// GenerateStream handles the streaming request from chat-gateway
func (s *LLMServiceImpl) GenerateStream(req *pbllm.GenerateStreamRequest, stream pbllm.LLMService_GenerateStreamServer) error {
	log.Printf("LLM GenerateStream request received: UserID=%s, ConvID=%s", req.UserId, req.ConversationId)

	ctx, cancel := context.WithTimeout(stream.Context(), 120*time.Second)
	defer cancel()

	// Prepare options for langchaingo streaming call
	options := []llms.CallOption{
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			token := string(chunk)
			grpcRes := &pbllm.GenerateStreamResponse{Token: token}
			if err := stream.Send(grpcRes); err != nil {
				log.Printf("gRPC stream send error: %v", err)
				return err
			}
			return nil
		}),
		llms.WithTemperature(0.7),
		llms.WithTopP(0.95),
		llms.WithPresencePenalty(0.3),
		llms.WithFrequencyPenalty(0.1),
		// llms.WithMaxTokens(1024),
	}

	log.Println("Calling langchaingo LLM GenerateContent...")

	_, err := llms.GenerateFromSinglePrompt(ctx, s.llm, req.Prompt, options...)
	if err != nil {
		return err
	}

	log.Printf("LLM GenerateStream completed successfully for UserID=%s, ConvID=%s", req.UserId, req.ConversationId)
	return nil
}

// GenerateWithRAG handles RAG-augmented streaming requests with enhanced features
func (s *LLMServiceImpl) GenerateWithRAG(
	req *pbllm.GenerateWithRAGRequest,
	stream pbllm.LLMService_GenerateWithRAGServer,
) error {
	log.Printf("LLM GenerateWithRAG request received: UserID=%s, ConvID=%s, Collection=%s, Adaptive=%v",
		req.UserId, req.ConversationId, req.RagCollection, req.Adaptive)

	ctx, cancel := context.WithTimeout(stream.Context(), 120*time.Second)
	defer cancel()

	// Use default collection if not specified
	collection := req.RagCollection
	if collection == "" {
		// TODO: work on this together with pinecone later
		collection = "university-scores-2"
	}

	// Initialize RAG state for enhanced processing
	state := &RAGState{
		Question:   req.Prompt,
		Documents:  []schema.Document{},
		Generation: "",
		Route:      RouteVectorStore,
		Iteration:  0,
		MaxRetries: s.ragConfig.MaxRetries,
	}

	// Step 1: Route query to appropriate data source if adaptive mode enabled
	var docs []schema.Document
	var err error

	if req.Adaptive {
		state.Route = s.routeQuery(ctx, req.Prompt)

		switch state.Route {
		case RouteWebSearch:
			docs, err = s.webSearch(ctx, req.Prompt)
			if err != nil {
				log.Printf("Web search failed, falling back to vectorstore: %v", err)
				state.Route = RouteVectorStore
				docs, err = s.retrieveDocuments(ctx, req.Prompt, collection)
			}
		case RouteVectorStore:
			docs, err = s.retrieveDocuments(ctx, req.Prompt, collection)
			// If no relevant documents found, try web search fallback
			if err == nil && len(docs) == 0 && s.ragConfig.WebSearchEnabled {
				log.Printf("No documents found in vectorstore, trying web search")
				webDocs, webErr := s.webSearch(ctx, req.Prompt)
				if webErr == nil && len(webDocs) > 0 {
					docs = webDocs
					state.Route = RouteWebSearch
				}
			}
		default:
			docs, err = s.retrieveDocuments(ctx, req.Prompt, collection)
		}
	} else {
		// Non-adaptive mode: use traditional RAG
		docs, err = s.retrieveDocuments(ctx, req.Prompt, collection)
	}

	if err != nil {
		log.Printf("Failed to retrieve documents: %v", err)
		docs = []schema.Document{}
	}

	// Step 2: Filter documents if adaptive mode is enabled
	relevantDocs := docs
	if req.Adaptive && len(docs) > 0 {
		relevantDocs = []schema.Document{}
		for _, doc := range docs {
			isRelevant, err := s.gradeDocumentRelevance(ctx, doc, req.Prompt)
			if err != nil {
				log.Printf("Failed to grade document relevance: %v", err)
				// Include document if grading fails
				relevantDocs = append(relevantDocs, doc)
			} else if isRelevant {
				relevantDocs = append(relevantDocs, doc)
			}
		}

		// If no relevant documents and we haven't tried web search yet, try it
		if len(relevantDocs) == 0 && state.Route == RouteVectorStore && s.ragConfig.WebSearchEnabled {
			log.Printf("No relevant documents found, trying web search fallback")
			webDocs, webErr := s.webSearch(ctx, req.Prompt)
			if webErr == nil {
				relevantDocs = webDocs
				state.Route = RouteWebSearch
			}
		}

		log.Printf("Filtered %d relevant documents from %d retrieved", len(relevantDocs), len(docs))
	}

	state.Documents = relevantDocs

	// Step 3: Build augmented prompt with context
	var contextBuilder strings.Builder
	if len(relevantDocs) > 0 {
		contextBuilder.WriteString("\n\nContext from retrieved sources:\n")
		for i, doc := range relevantDocs {
			source := "knowledge base"
			if sourceVal, ok := doc.Metadata["source"]; ok {
				if s, ok := sourceVal.(string); ok {
					source = s
				}
			}
			contextBuilder.WriteString(fmt.Sprintf("\n[Source %d - %s]:\n%s\n", i+1, source, doc.PageContent))
		}
	}

	// Enhanced prompt based on context availability
	var ragPrompt string
	if len(relevantDocs) > 0 {
		ragPrompt = fmt.Sprintf(`You are an AI assistant helping with career guidance and educational content.
Use the following retrieved context to answer the question accurately and helpfully.
If the context doesn't contain enough information, say so clearly.
Keep your answer concise but comprehensive.

Question: %s%s

Answer:`, req.Prompt, contextBuilder.String())
	} else {
		ragPrompt = fmt.Sprintf(`You are an AI assistant helping with career guidance and educational content.
I don't have specific context available for this question, so I'll provide a general response based on my knowledge.

Question: %s

Answer:`, req.Prompt)
	}

	// Step 4: Generate response with streaming
	var fullResponse strings.Builder
	options := []llms.CallOption{
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			token := string(chunk)
			fullResponse.WriteString(token)
			grpcRes := &pbllm.GenerateWithRAGResponse{Token: token}
			if err := stream.Send(grpcRes); err != nil {
				log.Printf("gRPC stream send error: %v", err)
				return err
			}
			return nil
		}),
		llms.WithTemperature(s.ragConfig.Temperature),
		llms.WithMaxTokens(s.ragConfig.MaxTokens),
	}

	log.Println("Calling langchaingo LLM GenerateContent with enhanced RAG context...")

	_, err = llms.GenerateFromSinglePrompt(ctx, s.llm, ragPrompt, options...)
	if err != nil {
		return fmt.Errorf("failed to generate response: %v", err)
	}

	// Step 5: Check for hallucinations if adaptive mode is enabled
	if req.Adaptive && len(relevantDocs) > 0 {
		isGrounded, err := s.checkHallucination(ctx, fullResponse.String(), relevantDocs)
		if err != nil {
			log.Printf("Failed to check hallucination: %v", err)
		} else if !isGrounded && state.Iteration < state.MaxRetries {
			log.Printf("Response may contain hallucinations, regenerating (attempt %d/%d)", state.Iteration+1, state.MaxRetries)

			// Clear the response and regenerate with stricter parameters
			fullResponse.Reset()
			strictOptions := []llms.CallOption{
				llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
					token := string(chunk)
					grpcRes := &pbllm.GenerateWithRAGResponse{Token: token}
					return stream.Send(grpcRes)
				}),
				llms.WithTemperature(0.3), // Lower temperature for more focused response
				llms.WithMaxTokens(s.ragConfig.MaxTokens),
			}

			_, err = llms.GenerateFromSinglePrompt(ctx, s.llm, ragPrompt, strictOptions...)
			if err != nil {
				return fmt.Errorf("failed to regenerate response: %v", err)
			}
		}
	}

	log.Printf("Enhanced RAG completed successfully for UserID=%s, ConvID=%s", req.UserId, req.ConversationId)
	return nil
}

// webSearch performs a web search using Tavily API
func (s *LLMServiceImpl) webSearch(ctx context.Context, query string) ([]schema.Document, error) {
	if !s.ragConfig.WebSearchEnabled || s.ragConfig.WebSearchAPIKey == "" {
		log.Printf("Web search disabled or API key missing")
		return []schema.Document{}, nil
	}

	// Prepare search request
	searchData := map[string]interface{}{
		"query":          query,
		"search_depth":   "basic",
		"include_raw":    false,
		"max_results":    3,
		"include_images": false,
	}

	requestBody, err := json.Marshal(searchData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", s.ragConfig.WebSearchBaseURL, strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", s.ragConfig.WebSearchAPIKey)

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned status: %d", resp.StatusCode)
	}

	// Parse response
	var searchResponse struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %v", err)
	}

	// Convert to schema.Document
	documents := make([]schema.Document, len(searchResponse.Results))
	for i, result := range searchResponse.Results {
		documents[i] = schema.Document{
			PageContent: result.Content,
			Metadata: map[string]interface{}{
				"title":  result.Title,
				"url":    result.URL,
				"source": "web_search",
			},
		}
	}

	log.Printf("Web search returned %d results for query: %s", len(documents), query)
	return documents, nil
}

// routeQuery determines the best data source for the query
func (s *LLMServiceImpl) routeQuery(ctx context.Context, query string) QueryRoute {
	// Use LLM to intelligently route the query
	routingPrompt := fmt.Sprintf(`You are an expert at routing a user question to a vectorstore or web search.
The vectorstore contains documents related to CareerUP platform, career guidance, university admissions, academic scores, and educational content.
Use the vectorstore for questions on these topics. For current events, news, or general knowledge questions, use web search.

Question: %s

Route to: vectorstore or web_search`, query)

	response, err := llms.GenerateFromSinglePrompt(ctx, s.llm, routingPrompt,
		llms.WithTemperature(0),
		llms.WithMaxTokens(20),
	)
	if err != nil {
		log.Printf("Failed to route query, defaulting to vectorstore: %v", err)
		return RouteVectorStore
	}

	answer := strings.ToLower(strings.TrimSpace(response))
	if strings.Contains(answer, "web_search") {
		log.Printf("Routing query to web search: %s", query)
		return RouteWebSearch
	}

	log.Printf("Routing query to vectorstore: %s", query)
	return RouteVectorStore
}

// Enhanced GenerateWithRAG with intelligent routing and state machine
func (s *LLMServiceImpl) GenerateWithRAGEnhanced(
	req *pbllm.GenerateWithRAGRequest,
	stream pbllm.LLMService_GenerateWithRAGServer,
) error {
	log.Printf("Enhanced RAG request: UserID=%s, Query=%s", req.UserId, req.Prompt)

	ctx, cancel := context.WithTimeout(stream.Context(), 120*time.Second)
	defer cancel()

	// Initialize RAG state
	state := &RAGState{
		Question:   req.Prompt,
		Documents:  []schema.Document{},
		Generation: "",
		Route:      RouteVectorStore,
		Iteration:  0,
		MaxRetries: s.ragConfig.MaxRetries,
	}

	// Step 1: Route query to appropriate data source
	if req.Adaptive {
		state.Route = s.routeQuery(ctx, req.Prompt)
	}

	// Step 2: Retrieve documents based on routing decision
	var docs []schema.Document
	var err error

	collection := req.RagCollection
	if collection == "" {
		collection = "university-scores-2"
	}

	switch state.Route {
	case RouteWebSearch:
		docs, err = s.webSearch(ctx, req.Prompt)
		if err != nil {
			log.Printf("Web search failed, falling back to vectorstore: %v", err)
			state.Route = RouteVectorStore
			docs, err = s.retrieveDocuments(ctx, req.Prompt, collection)
		}
	case RouteVectorStore:
		docs, err = s.retrieveDocuments(ctx, req.Prompt, collection)
		// If no relevant documents found and adaptive mode enabled, try web search
		if err == nil && len(docs) == 0 && req.Adaptive && s.ragConfig.WebSearchEnabled {
			log.Printf("No documents found in vectorstore, trying web search")
			webDocs, webErr := s.webSearch(ctx, req.Prompt)
			if webErr == nil && len(webDocs) > 0 {
				docs = webDocs
				state.Route = RouteWebSearch
			}
		}
	default:
		docs, err = s.retrieveDocuments(ctx, req.Prompt, collection)
	}

	if err != nil {
		log.Printf("Failed to retrieve documents: %v", err)
		docs = []schema.Document{}
	}

	state.Documents = docs

	// Step 3: Filter relevant documents if adaptive mode
	relevantDocs := docs
	if req.Adaptive && len(docs) > 0 {
		relevantDocs = []schema.Document{}
		for _, doc := range docs {
			isRelevant, err := s.gradeDocumentRelevance(ctx, doc, req.Prompt)
			if err != nil {
				log.Printf("Failed to grade document relevance: %v", err)
				relevantDocs = append(relevantDocs, doc)
			} else if isRelevant {
				relevantDocs = append(relevantDocs, doc)
			}
		}

		// If no relevant documents and we haven't tried web search yet, try it
		if len(relevantDocs) == 0 && state.Route == RouteVectorStore && s.ragConfig.WebSearchEnabled {
			log.Printf("No relevant documents found, trying web search fallback")
			webDocs, webErr := s.webSearch(ctx, req.Prompt)
			if webErr == nil {
				relevantDocs = webDocs
				state.Route = RouteWebSearch
			}
		}

		log.Printf("Filtered %d relevant documents from %d retrieved", len(relevantDocs), len(docs))
		state.Documents = relevantDocs
	}

	// Step 4: Generate response with context
	return s.generateStreamingResponse(ctx, req, state, stream)
}

// generateStreamingResponse handles the actual response generation with state management
func (s *LLMServiceImpl) generateStreamingResponse(
	ctx context.Context,
	req *pbllm.GenerateWithRAGRequest,
	state *RAGState,
	stream pbllm.LLMService_GenerateWithRAGServer,
) error {
	// Build context from documents
	var contextBuilder strings.Builder
	if len(state.Documents) > 0 {
		contextBuilder.WriteString("\n\nContext from retrieved sources:\n")
		for i, doc := range state.Documents {
			source := "knowledge base"
			if sourceVal, ok := doc.Metadata["source"]; ok {
				if s, ok := sourceVal.(string); ok {
					source = s
				}
			}
			contextBuilder.WriteString(fmt.Sprintf("\n[Source %d - %s]:\n%s\n", i+1, source, doc.PageContent))
		}
	}

	// Enhanced prompt based on context availability
	var ragPrompt string
	if len(state.Documents) > 0 {
		ragPrompt = fmt.Sprintf(`You are an AI assistant helping with career guidance and educational content.
Use the following retrieved context to answer the question accurately and helpfully.
If the context doesn't contain enough information, say so clearly.
Keep your answer concise but comprehensive.

Question: %s%s

Answer:`, req.Prompt, contextBuilder.String())
	} else {
		ragPrompt = fmt.Sprintf(`You are an AI assistant helping with career guidance and educational content.
I don't have specific context available for this question, so I'll provide a general response based on my knowledge.

Question: %s

Answer:`, req.Prompt)
	}

	// Generate response with streaming
	var fullResponse strings.Builder
	options := []llms.CallOption{
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			token := string(chunk)
			fullResponse.WriteString(token)
			grpcRes := &pbllm.GenerateWithRAGResponse{Token: token}
			if err := stream.Send(grpcRes); err != nil {
				log.Printf("gRPC stream send error: %v", err)
				return err
			}
			return nil
		}),
		llms.WithTemperature(s.ragConfig.Temperature),
		llms.WithMaxTokens(s.ragConfig.MaxTokens),
	}

	_, err := llms.GenerateFromSinglePrompt(ctx, s.llm, ragPrompt, options...)
	if err != nil {
		return fmt.Errorf("failed to generate response: %v", err)
	}

	// Check for hallucinations if adaptive mode and we have context
	if req.Adaptive && len(state.Documents) > 0 {
		isGrounded, err := s.checkHallucination(ctx, fullResponse.String(), state.Documents)
		if err != nil {
			log.Printf("Failed to check hallucination: %v", err)
		} else if !isGrounded && state.Iteration < state.MaxRetries {
			log.Printf("Response may contain hallucinations, regenerating (attempt %d/%d)", state.Iteration+1, state.MaxRetries)
			state.Iteration++

			// Clear response and regenerate with stricter parameters
			fullResponse.Reset()
			strictOptions := []llms.CallOption{
				llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
					token := string(chunk)
					grpcRes := &pbllm.GenerateWithRAGResponse{Token: token}
					return stream.Send(grpcRes)
				}),
				llms.WithTemperature(0.3), // Lower temperature
				llms.WithMaxTokens(s.ragConfig.MaxTokens),
			}

			_, err = llms.GenerateFromSinglePrompt(ctx, s.llm, ragPrompt, strictOptions...)
			if err != nil {
				return fmt.Errorf("failed to regenerate response: %v", err)
			}
		}
	}

	state.Generation = fullResponse.String()
	log.Printf("Enhanced RAG completed successfully for UserID=%s", req.UserId)
	return nil
}

// InitializeVectorStore creates a new vector store for the given collection
func (s *LLMServiceImpl) InitializeVectorStore(collection string) error {
	log.Printf("Creating Pinecone index: %s", collection)

	// Add nil checks
	if s.pineconeClient == nil {
		return fmt.Errorf("Pinecone client is not initialized")
	}

	// Check if collection already exists
	if _, exists := s.collections[collection]; exists {
		log.Printf("Collection %s already exists", collection)
		return nil
	}

	// Create collection using our Pinecone client
	req := pinecone.CreateCollectionRequest{
		Name:     collection,
		Metadata: map[string]string{"created_at": time.Now().Format(time.RFC3339)},
	}

	createdCollection, err := s.pineconeClient.CreateCollection(req)
	if err != nil {
		return fmt.Errorf("failed to create Pinecone index: %w", err)
	}

	s.collections[collection] = createdCollection
	log.Printf("Successfully created Pinecone index: %s", collection)
	return nil
}

// ProcessDocumentChunks processes and indexes document chunks into a vector store
func (s *LLMServiceImpl) ProcessDocumentChunks(content string, collection string) error {
	// Create document
	doc := schema.Document{
		PageContent: content,
		Metadata: map[string]interface{}{
			"indexed_at": time.Now().Format(time.RFC3339),
		},
	}

	// Simplified: Use a basic splitter (split by paragraphs)
	chunks := strings.Split(doc.PageContent, "\n\n")

	// Check if collection exists
	_, ok := s.collections[collection]
	if !ok {
		log.Printf("Collection %s not found, documents not indexed", collection)
		return nil
	}

	// Convert chunks to Pinecone documents with embeddings
	var pineconeDocs []pinecone.Document
	for i, chunk := range chunks {
		chunkText := strings.TrimSpace(chunk)
		if chunkText == "" {
			continue // Skip empty chunks
		}

		// Generate embedding for this chunk
		embedding, err := s.embedder.EmbedQuery(context.Background(), chunkText)
		if err != nil {
			log.Printf("Failed to generate embedding for chunk %d: %v", i, err)
			continue
		}

		// Convert embedding to float32 slice
		embeddingFloat32 := make([]float32, len(embedding))
		for j, v := range embedding {
			embeddingFloat32[j] = float32(v)
		}

		pineconeDocs = append(pineconeDocs, pinecone.Document{
			ID:        fmt.Sprintf("%s_%d_%d", collection, time.Now().Unix(), i),
			Content:   chunkText,
			Embedding: embeddingFloat32,
			Metadata: map[string]string{
				"chunk_index": fmt.Sprintf("%d", i),
				"indexed_at":  time.Now().Format(time.RFC3339),
			},
		})
	}

	if len(pineconeDocs) == 0 {
		log.Printf("No valid chunks to index for collection %s", collection)
		return nil
	}

	log.Printf("Generated embeddings for %d chunks in collection %s", len(pineconeDocs), collection)

	// Add documents to Pinecone
	req := pinecone.AddDocumentsRequest{
		Documents: pineconeDocs,
	}

	_, err := s.pineconeClient.AddDocuments(collection, req)
	if err != nil {
		return fmt.Errorf("failed to add documents to Pinecone: %v", err)
	}

	log.Printf("Indexed %d document chunks into collection %s", len(pineconeDocs), collection)
	return nil
}
