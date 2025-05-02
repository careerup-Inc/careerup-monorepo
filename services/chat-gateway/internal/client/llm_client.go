package client

import (
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // Use insecure for local development

	pbllm "github.com/careerup-Inc/careerup-monorepo/proto/llm/v1" // Corrected import path
)

// LLMClient wraps the gRPC client for the LLM service.
type LLMClient struct {
	grpcClient pbllm.LLMServiceClient
	conn       *grpc.ClientConn // Keep a reference to close it later
}

// NewLLMClient creates a new gRPC client for the LLM service.
func NewLLMClient(llmServiceAddr string) (*LLMClient, error) {
	log.Printf("Attempting to connect to LLM gRPC service at %s", llmServiceAddr)
	// Establish gRPC connection (use insecure credentials for local dev)
	// Add options for retry, timeout, etc. in production
	conn, err := grpc.Dial(
		llmServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),                 // Block until connection is up or fails
		grpc.WithTimeout(10*time.Second), // Connection timeout
	)
	if err != nil {
		log.Printf("Failed to connect to LLM service at %s: %v", llmServiceAddr, err)
		return nil, err
	}

	client := pbllm.NewLLMServiceClient(conn)
	log.Printf("Connected to LLM gRPC service at %s", llmServiceAddr)
	return &LLMClient{grpcClient: client, conn: conn}, nil
}

// GetLLMServiceClient returns the raw gRPC client interface.
func (c *LLMClient) GetLLMServiceClient() pbllm.LLMServiceClient {
	return c.grpcClient
}

// Close closes the underlying gRPC connection.
func (c *LLMClient) Close() error {
	log.Println("Closing connection to LLM gRPC service...")
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
