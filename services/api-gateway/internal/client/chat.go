package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/proto/careerup/chat"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChatClient struct {
	conn   *grpc.ClientConn
	client chat.ChatServiceClient
}

func NewChatClient(addr string) (*ChatClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to chat service: %v", err)
	}

	return &ChatClient{
		conn:   conn,
		client: chat.NewChatServiceClient(conn),
	}, nil
}

func (c *ChatClient) Close() error {
	return c.conn.Close()
}

func (c *ChatClient) SendMessage(ctx context.Context, message *chat.ChatMessage) (*chat.SendMessageResponse, error) {
	return c.client.SendMessage(ctx, &chat.SendMessageRequest{
		Message: message,
	})
}

func (c *ChatClient) StreamMessages(ctx context.Context, userID string) (chat.ChatService_StreamMessagesClient, error) {
	return c.client.StreamMessages(ctx, &chat.StreamMessagesRequest{
		UserId: userID,
	})
}

func (c *ChatClient) UpgradeToWebSocket(ctx *fiber.Ctx) error {
	// Get the user ID from the context
	userID := ctx.Locals("user_id")
	if userID == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "user ID not found")
	}

	// Create a context for the WebSocket connection
	wsCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start streaming messages from chat-gateway
	stream, err := c.StreamMessages(wsCtx, userID.(string))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to stream messages: %v", err))
	}

	// Upgrade to WebSocket
	ws := websocket.New(func(conn *websocket.Conn) {
		// Handle incoming WebSocket messages
		go func() {
			defer cancel()
			for {
				msgType, msg, err := conn.ReadMessage()
				if err != nil {
					break
				}

				// Only handle text messages
				if msgType != websocket.TextMessage {
					continue
				}

				// Send message to chat-gateway
				_, err = c.SendMessage(wsCtx, &chat.ChatMessage{
					Content:  string(msg),
					SenderId: userID.(string),
				})
				if err != nil {
					break
				}
			}
		}()

		// Stream messages from chat-gateway to WebSocket
		go func() {
			defer cancel()
			for {
				msg, err := stream.Recv()
				if err != nil {
					break
				}

				// Marshal message to JSON
				jsonMsg, err := json.Marshal(msg)
				if err != nil {
					break
				}

				if err := conn.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
					break
				}
			}
		}()

		// Wait for context cancellation
		<-wsCtx.Done()
	})

	return ws(ctx)
}
