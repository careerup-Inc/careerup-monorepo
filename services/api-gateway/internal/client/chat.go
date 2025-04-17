package client

import (
	"context"
	"log"

	v1 "github.com/careerup-Inc/careerup-monorepo/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChatClient struct {
	client v1.ConversationServiceClient
	conn   *grpc.ClientConn
}

func NewChatClient(addr string) (*ChatClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &ChatClient{
		client: v1.NewConversationServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *ChatClient) Close() error {
	return c.conn.Close()
}

func (c *ChatClient) Stream(ctx context.Context) (v1.ConversationService_StreamClient, error) {
	stream, err := c.client.Stream(ctx)
	if err != nil {
		log.Printf("Failed to create chat stream: %v", err)
		return nil, err
	}
	return stream, nil
}
