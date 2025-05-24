package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tmc/langchaingo/schema"

	pbllm "github.com/careerup-Inc/careerup-monorepo/proto/llm/v1"
	"github.com/careerup-Inc/careerup-monorepo/services/llm-gateway/internal/pinecone"
)

// IngestDocument ingests a document into the specified collection
func (s *LLMServiceImpl) IngestDocument(ctx context.Context, req *pbllm.IngestDocumentRequest) (*pbllm.IngestDocumentResponse, error) {
	log.Printf("Ingesting document into collection: %s", req.GetCollection())

	// Validate request
	if req.GetContent() == "" {
		return &pbllm.IngestDocumentResponse{
			Success: false,
			Message: "Document content cannot be empty",
		}, nil
	}

	collection := req.GetCollection()
	if collection == "" {
		collection = "default"
	}

	// Generate document ID if not provided
	documentId := req.GetDocumentId()
	if documentId == "" {
		documentId = fmt.Sprintf("doc_%d", time.Now().Unix())
	}

	// Ensure vector store exists for the collection
	if err := s.InitializeVectorStore(collection); err != nil {
		return &pbllm.IngestDocumentResponse{
			DocumentId: documentId,
			Success:    false,
			Message:    fmt.Sprintf("Failed to initialize vector store: %v", err),
		}, nil
	}

	// Create document with metadata
	doc := schema.Document{
		PageContent: req.GetContent(),
		Metadata: map[string]interface{}{
			"document_id": documentId,
			"indexed_at":  time.Now().Format(time.RFC3339),
		},
	}

	// Add custom metadata from request
	for k, v := range req.GetMetadata() {
		doc.Metadata[k] = v
	}

	// Split document into chunks (reusing existing logic)
	chunks := strings.Split(doc.PageContent, "\n\n")
	splitDocs := make([]schema.Document, 0, len(chunks))

	for i, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			continue // Skip empty chunks
		}

		chunkDoc := schema.Document{
			PageContent: strings.TrimSpace(chunk),
			Metadata: map[string]interface{}{
				"document_id": documentId,
				"chunk_index": i,
				"indexed_at":  time.Now().Format(time.RFC3339),
			},
		}

		// Add custom metadata to each chunk
		for k, v := range req.GetMetadata() {
			chunkDoc.Metadata[k] = v
		}

		splitDocs = append(splitDocs, chunkDoc)
	}

	// Add to Pinecone collection
	if _, exists := s.collections[collection]; !exists {
		return &pbllm.IngestDocumentResponse{
			DocumentId: documentId,
			Success:    false,
			Message:    fmt.Sprintf("Collection %s does not exist", collection),
		}, nil
	}

	// Convert to Pinecone documents with embeddings
	var pineconeDocs []pinecone.Document
	for i, doc := range splitDocs {
		// Generate embedding for this document
		embedding, err := s.embedder.EmbedQuery(ctx, doc.PageContent)
		if err != nil {
			log.Printf("Failed to generate embedding for document chunk %d: %v", i, err)
			continue
		}

		// Convert embedding to float32 slice
		embeddingFloat32 := make([]float32, len(embedding))
		for j, v := range embedding {
			embeddingFloat32[j] = float32(v)
		}

		pineconeDocs = append(pineconeDocs, pinecone.Document{
			ID:        fmt.Sprintf("%s_%d", documentId, i),
			Content:   doc.PageContent,
			Embedding: embeddingFloat32,
			Metadata: map[string]string{
				"document_id": documentId,
				"chunk_index": fmt.Sprintf("%d", i),
				"indexed_at":  time.Now().Format(time.RFC3339),
			},
		})
	}

	// Add documents to Pinecone
	addReq := pinecone.AddDocumentsRequest{
		Documents: pineconeDocs,
	}

	_, err := s.pineconeClient.AddDocuments(collection, addReq)
	if err != nil {
		return &pbllm.IngestDocumentResponse{
			DocumentId: documentId,
			Success:    false,
			Message:    fmt.Sprintf("Failed to add document to Pinecone: %v", err),
		}, nil
	}

	log.Printf("Successfully ingested document %s into collection %s with %d chunks",
		documentId, collection, len(pineconeDocs))

	return &pbllm.IngestDocumentResponse{
		DocumentId:    documentId,
		Success:       true,
		Message:       "Document successfully ingested",
		ChunksCreated: int32(len(pineconeDocs)),
	}, nil
}

// CreateCollection creates a new vector store collection
func (s *LLMServiceImpl) CreateCollection(ctx context.Context, req *pbllm.CreateCollectionRequest) (*pbllm.CreateCollectionResponse, error) {
	collectionName := req.GetCollectionName()
	log.Printf("Creating collection: %s", collectionName)

	if collectionName == "" {
		return &pbllm.CreateCollectionResponse{
			Success: false,
			Message: "Collection name cannot be empty",
		}, nil
	}

	// Check if collection already exists
	if _, exists := s.collections[collectionName]; exists {
		return &pbllm.CreateCollectionResponse{
			Success:        false,
			Message:        "Collection already exists",
			CollectionName: collectionName,
		}, nil
	}

	// Initialize the vector store for the new collection
	if err := s.InitializeVectorStore(collectionName); err != nil {
		return &pbllm.CreateCollectionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to create collection: %v", err),
		}, nil
	}

	log.Printf("Successfully created collection: %s", collectionName)

	return &pbllm.CreateCollectionResponse{
		Success:        true,
		Message:        "Collection created successfully",
		CollectionName: collectionName,
	}, nil
}

// ListCollections lists all available collections
func (s *LLMServiceImpl) ListCollections(ctx context.Context, req *pbllm.ListCollectionsRequest) (*pbllm.ListCollectionsResponse, error) {
	log.Printf("Listing collections")

	// Get collections from Pinecone
	pineconeCollections, err := s.pineconeClient.ListCollections()
	if err != nil {
		log.Printf("Failed to list Pinecone collections: %v", err)
		// Fall back to local collections if Pinecone fails
		var collections []*pbllm.CollectionInfo
		for name := range s.collections {
			collections = append(collections, &pbllm.CollectionInfo{
				Name: name,
			})
		}
		return &pbllm.ListCollectionsResponse{
			Collections: collections,
		}, nil
	}

	var collections []*pbllm.CollectionInfo
	for _, collection := range pineconeCollections {
		collections = append(collections, &pbllm.CollectionInfo{
			Name: collection.Name,
		})
	}

	return &pbllm.ListCollectionsResponse{
		Collections: collections,
	}, nil
}

// DeleteCollection deletes a collection and all its documents
func (s *LLMServiceImpl) DeleteCollection(ctx context.Context, req *pbllm.DeleteCollectionRequest) (*pbllm.DeleteCollectionResponse, error) {
	collectionName := req.GetCollectionName()
	log.Printf("Deleting collection: %s", collectionName)

	if collectionName == "" {
		return &pbllm.DeleteCollectionResponse{
			Success: false,
			Message: "Collection name cannot be empty",
		}, nil
	}

	// Check if collection exists
	if _, exists := s.collections[collectionName]; !exists {
		return &pbllm.DeleteCollectionResponse{
			Success: false,
			Message: "Collection does not exist",
		}, nil
	}

	// Delete from Pinecone
	err := s.pineconeClient.DeleteCollection(collectionName)
	if err != nil {
		log.Printf("Failed to delete Pinecone index: %v", err)
		return &pbllm.DeleteCollectionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to delete Pinecone index: %v", err),
		}, nil
	}

	// Remove from our local map
	delete(s.collections, collectionName)

	log.Printf("Successfully deleted collection: %s", collectionName)

	return &pbllm.DeleteCollectionResponse{
		Success: true,
		Message: "Collection deleted successfully",
	}, nil
}
