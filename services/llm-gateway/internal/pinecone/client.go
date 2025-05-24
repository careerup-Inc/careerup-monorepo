// Package pinecone provides a client for the Pinecone vector database
package pinecone

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

// Client represents a Pinecone client
type Client struct {
	pc       *pinecone.Client
	indexMap map[string]*pinecone.IndexConnection // Map collection names to index connections
}

// Collection represents a Pinecone index
type Collection struct {
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Document represents a document to be upserted to Pinecone
type Document struct {
	ID        string            `json:"id"`
	Content   string            `json:"content"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Embedding []float32         `json:"embedding,omitempty"`
}

// CreateCollectionRequest is used to create a new index
type CreateCollectionRequest struct {
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// AddDocumentsRequest is used to upsert documents to an index
type AddDocumentsRequest struct {
	Documents []Document `json:"documents"`
}

// AddDocumentsResponse is the response from upserting documents
type AddDocumentsResponse struct {
	Success bool     `json:"success"`
	IDs     []string `json:"ids"`
}

// QueryRequest is used to query documents from an index
type QueryRequest struct {
	QueryEmbeddings [][]float32 `json:"query_embeddings"`
	NResults        int         `json:"n_results,omitempty"`
	Include         []string    `json:"include,omitempty"`
}

// QueryResult represents a single query result
type QueryResult struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Distance float32           `json:"distance,omitempty"`
	Score    float32           `json:"score,omitempty"`
}

// QueryResponse is the response from querying documents
type QueryResponse struct {
	Results [][]QueryResult `json:"results"`
}

// New creates a new Pinecone client
func New() (*Client, error) {
	apiKey := os.Getenv("PINECONE_API_KEY")
	if apiKey == "" {
		log.Fatal("PINECONE_API_KEY environment variable not set")
	}

	return NewWithAPIKey(apiKey)
}

// NewWithAPIKey creates a new Pinecone client with the provided API key
func NewWithAPIKey(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	// Initialize the actual Pinecone client
	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pinecone client: %w", err)
	}

	return &Client{
		pc:       pc,
		indexMap: make(map[string]*pinecone.IndexConnection),
	}, nil
}

// CreateCollection creates a new index in Pinecone (using collection naming convention)
func (c *Client) CreateCollection(req CreateCollectionRequest) (*Collection, error) {
	// Add nil check at the beginning
	if c == nil {
		return nil, fmt.Errorf("Pinecone client is nil")
	}
	if c.pc == nil {
		return nil, fmt.Errorf("Pinecone API client is nil")
	}

	// Check if index already exists
	if _, exists := c.indexMap[req.Name]; exists {
		return &Collection{Name: req.Name}, nil
	}

	ctx := context.Background()

	// Use 1536 dimensions for OpenAI text-embedding-ada-002
	dimension := int32(1536)
	metric := pinecone.Cosine

	// Check if index already exists
	indexDesc, err := c.pc.DescribeIndex(ctx, req.Name)
	if err == nil {
		// Index already exists, connect to it
		indexConn, err := c.pc.Index(pinecone.NewIndexConnParams{Host: indexDesc.Host})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to existing index: %w", err)
		}
		c.indexMap[req.Name] = indexConn

		return &Collection{
			Name:     req.Name,
			Metadata: req.Metadata,
		}, nil
	}

	// Create new serverless index
	_, err = c.pc.CreateServerlessIndex(ctx, &pinecone.CreateServerlessIndexRequest{
		Name:      req.Name,
		Dimension: &dimension,
		Metric:    &metric,
		Cloud:     pinecone.Aws,
		Region:    "us-east-1", // Default region, can be configurable
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	// Wait for index to be ready
	for {
		indexDesc, err := c.pc.DescribeIndex(ctx, req.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to describe index: %w", err)
		}

		if indexDesc.Status.Ready {
			break
		}

		log.Printf("Waiting for index %s to be ready...", req.Name)
		time.Sleep(5 * time.Second)
	}

	// Connect to the index
	indexConn, err := c.pc.Index(pinecone.NewIndexConnParams{Host: indexDesc.Host})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to index: %w", err)
	}

	c.indexMap[req.Name] = indexConn

	return &Collection{
		Name:     req.Name,
		Metadata: req.Metadata,
	}, nil
}

// ListCollections lists all indexes in Pinecone (using collection naming convention)
func (c *Client) ListCollections() ([]Collection, error) {
	ctx := context.Background()

	indexes, err := c.pc.ListIndexes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}

	var collections []Collection
	for _, index := range indexes {
		collections = append(collections, Collection{
			Name: index.Name,
		})
	}

	return collections, nil
}

// AddDocuments upserts documents to an index
func (c *Client) AddDocuments(collectionName string, req AddDocumentsRequest) (*AddDocumentsResponse, error) {
	ctx := context.Background()

	index, exists := c.indexMap[collectionName]
	if !exists {
		// Try to connect to the index
		indexDesc, err := c.pc.DescribeIndex(ctx, collectionName)
		if err != nil {
			return nil, fmt.Errorf("index %s not found: %w", collectionName, err)
		}

		index, err = c.pc.Index(pinecone.NewIndexConnParams{Host: indexDesc.Host})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to index %s: %w", collectionName, err)
		}
		c.indexMap[collectionName] = index
	}

	// Convert documents to Pinecone vectors
	var vectors []*pinecone.Vector
	var processedIDs []string

	for _, doc := range req.Documents {
		if len(doc.Embedding) == 0 {
			return nil, fmt.Errorf("document %s is missing embedding", doc.ID)
		}

		// Add content to metadata using structpb
		metadataMap := make(map[string]interface{})
		for k, v := range doc.Metadata {
			metadataMap[k] = v
		}
		metadataMap["content"] = doc.Content

		metadata, err := structpb.NewStruct(metadataMap)
		if err != nil {
			return nil, fmt.Errorf("failed to create metadata for document %s: %w", doc.ID, err)
		}

		vector := &pinecone.Vector{
			Id:       doc.ID,
			Values:   &doc.Embedding,
			Metadata: metadata,
		}

		vectors = append(vectors, vector)
		processedIDs = append(processedIDs, doc.ID)
	}

	// Upsert vectors
	_, err := index.UpsertVectors(ctx, vectors)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert vectors: %w", err)
	}

	return &AddDocumentsResponse{
		Success: true,
		IDs:     processedIDs,
	}, nil
}

// QueryDocuments queries documents from an index using vector similarity
func (c *Client) QueryDocuments(collectionName string, req QueryRequest) (*QueryResponse, error) {
	ctx := context.Background()

	index, exists := c.indexMap[collectionName]
	if !exists {
		// Try to connect to the index
		indexDesc, err := c.pc.DescribeIndex(ctx, collectionName)
		if err != nil {
			return nil, fmt.Errorf("index %s not found: %w", collectionName, err)
		}

		index, err = c.pc.Index(pinecone.NewIndexConnParams{Host: indexDesc.Host})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to index %s: %w", collectionName, err)
		}
		c.indexMap[collectionName] = index
	}

	// Pinecone supports only one query at a time, so we'll process the first embedding
	if len(req.QueryEmbeddings) == 0 {
		return &QueryResponse{Results: [][]QueryResult{}}, nil
	}

	queryVector := req.QueryEmbeddings[0]
	topK := uint32(req.NResults)
	if topK == 0 {
		topK = 5 // Default
	}

	includeMetadata := false
	includeValues := false
	for _, include := range req.Include {
		if include == "metadatas" || include == "metadata" {
			includeMetadata = true
		}
		if include == "values" {
			includeValues = true
		}
	}

	// Query the index
	queryResp, err := index.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:          queryVector,
		TopK:            topK,
		IncludeMetadata: includeMetadata,
		IncludeValues:   includeValues,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query vectors: %w", err)
	}

	// Convert results
	var results []QueryResult
	for _, match := range queryResp.Matches {
		result := QueryResult{
			ID:    match.Vector.Id,
			Score: match.Score,
		}

		// Extract content from metadata
		if match.Vector.Metadata != nil {
			metadata := make(map[string]string)
			metadataStruct := match.Vector.Metadata.AsMap()

			for k, v := range metadataStruct {
				if k == "content" {
					if content, ok := v.(string); ok {
						result.Content = content
					}
				} else {
					if strVal, ok := v.(string); ok {
						metadata[k] = strVal
					} else {
						metadata[k] = fmt.Sprintf("%v", v)
					}
				}
			}
			result.Metadata = metadata
		}

		results = append(results, result)
	}

	return &QueryResponse{
		Results: [][]QueryResult{results},
	}, nil
}

// DeleteCollection deletes an index from Pinecone
func (c *Client) DeleteCollection(collectionName string) error {
	ctx := context.Background()

	err := c.pc.DeleteIndex(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}

	// Remove from our map
	delete(c.indexMap, collectionName)

	return nil
}

// GetOrCreateIndex ensures an index exists and returns it
func (c *Client) GetOrCreateIndex(indexName string) (*pinecone.IndexConnection, error) {
	// Check if we already have it
	if index, exists := c.indexMap[indexName]; exists {
		return index, nil
	}

	// Try to connect to existing index
	indexDesc, err := c.pc.DescribeIndex(context.Background(), indexName)
	if err == nil {
		index, err := c.pc.Index(pinecone.NewIndexConnParams{Host: indexDesc.Host})
		if err == nil {
			c.indexMap[indexName] = index
			return index, nil
		}
	}

	// Create the index if it doesn't exist
	_, err = c.CreateCollection(CreateCollectionRequest{
		Name: indexName,
	})
	if err != nil {
		return nil, err
	}

	return c.indexMap[indexName], nil
}
