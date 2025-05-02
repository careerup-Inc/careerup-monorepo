package client

import (
	"fmt"
	// Added for logging
	// Import the correct v1 proto package
	chatpb "github.com/careerup-Inc/careerup-monorepo/proto/careerup/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChatClientInterface interface {
	// GetChatServiceClient returns the raw gRPC client for the ConversationService.
	GetChatServiceClient() chatpb.ConversationServiceClient
	Close() error
}

type ChatClient struct {
	conn *grpc.ClientConn
	// Use the ConversationServiceClient from the v1 proto
	client chatpb.ConversationServiceClient
}

// NewChatClient needs to initialize the ConversationServiceClient
func NewChatClient(addr string) (*ChatClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to chat service at %s: %w", addr, err)
	}
	client := chatpb.NewConversationServiceClient(conn)
	return &ChatClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetChatServiceClient implements the ChatClientInterface.
func (c *ChatClient) GetChatServiceClient() chatpb.ConversationServiceClient {
	return c.client
}

func (c *ChatClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
