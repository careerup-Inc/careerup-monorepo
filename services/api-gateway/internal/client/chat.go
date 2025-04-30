package client

import (
	"context"
	"fmt"
	"log" // Added for logging

	// Import the correct v1 proto package
	chatpb "github.com/careerup-Inc/careerup-monorepo/proto/careerup/v1"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type ChatClientInterface interface {
	UpgradeToWebSocket(ctx *fiber.Ctx, user *User) error
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

func (c *ChatClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Remove SendMessage and StreamMessages methods as they are replaced by the bidirectional stream logic

// UpgradeToWebSocket needs to be completely rewritten to use the bidirectional Stream RPC
func (c *ChatClient) UpgradeToWebSocket(ctx *fiber.Ctx, user *User) error { // Pass the authenticated user
	// Upgrade HTTP connection to WebSocket
	return websocket.New(func(conn *websocket.Conn) {
		defer conn.Close()
		log.Printf("WebSocket connection established for user: %s", user.ID)

		streamCtx, cancelStream := context.WithCancel(context.Background())
		defer cancelStream() // Ensure cancellation on exit

		// --- Establish gRPC Stream ---
		stream, err := c.client.Stream(streamCtx)
		if err != nil {
			log.Printf("Error establishing gRPC stream for user %s: %v", user.ID, err)
			conn.WriteJSON(chatpb.WebSocketMessage{Type: "error", Payload: &chatpb.WebSocketMessage_AssistantToken{AssistantToken: &chatpb.AssistantToken{Token: "Failed to connect to chat service"}}})
			return
		}
		log.Printf("gRPC stream established for user: %s", user.ID)

		// --- Goroutine: Read from WebSocket, Send to gRPC Stream ---
		go func() {
			defer cancelStream() // Cancel stream context if this goroutine exits
			for {
				var wsMsg chatpb.WebSocketMessage
				// Read message from WebSocket client
				if err := conn.ReadJSON(&wsMsg); err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("WebSocket read error for user %s: %v", user.ID, err)
					} else {
						log.Printf("WebSocket closed for user %s", user.ID)
					}
					return
				}

				// Process only user messages
				if wsMsg.Type == "user_msg" {
					userMsg := wsMsg.GetUserMessage()
					if userMsg != nil {
						log.Printf("Received user_msg from %s: conv=%s, text=%s", user.ID, userMsg.ConversationId, userMsg.Text)
						// Prepare gRPC request
						streamReq := &chatpb.StreamRequest{
							ConversationId: userMsg.ConversationId,
							Text:           userMsg.Text,
							// Add user_id if needed by the backend service
						}
						// Send to gRPC stream
						if err := stream.Send(streamReq); err != nil {
							log.Printf("Error sending to gRPC stream for user %s: %v", user.ID, err)
							// Optionally notify the client
							// conn.WriteJSON(...)
							return // Exit goroutine on send error
						}
						log.Printf("Sent message to gRPC stream for user %s", user.ID)
					}
				} else {
					log.Printf("Received non-user_msg type '%s' from user %s", wsMsg.Type, user.ID)
				}
			}
		}()

		// --- Goroutine: Receive from gRPC Stream, Write to WebSocket ---
		for {
			// Receive response from gRPC stream
			resp, err := stream.Recv()
			if err != nil {
				// Check for context cancellation vs other errors
				grpcStatus, ok := status.FromError(err)
				if ok && (grpcStatus.Code() == codes.Canceled || grpcStatus.Code() == codes.Unavailable) {
					log.Printf("gRPC stream closed or unavailable for user %s: %v", user.ID, err)
				} else {
					log.Printf("Error receiving from gRPC stream for user %s: %v", user.ID, err)
					// Send error to client if connection is still open
					conn.WriteJSON(chatpb.WebSocketMessage{Type: "error", Payload: &chatpb.WebSocketMessage_AssistantToken{AssistantToken: &chatpb.AssistantToken{Token: "Chat service error"}}})
				}
				return // Exit loop (and handler) on receive error/close
			}

			log.Printf("Received response from gRPC stream for user %s", user.ID)

			// Prepare WebSocket message based on gRPC response
			var wsResp chatpb.WebSocketMessage
			if token := resp.GetToken(); token != "" {
				wsResp.Type = "assistant_token"
				wsResp.Payload = &chatpb.WebSocketMessage_AssistantToken{
					AssistantToken: &chatpb.AssistantToken{Token: token},
				}
			} else if avatarURL := resp.GetAvatarUrl(); avatarURL != "" {
				wsResp.Type = "avatar_url"
				wsResp.Payload = &chatpb.WebSocketMessage_AvatarUrl{
					AvatarUrl: &chatpb.AvatarUrl{Url: avatarURL},
				}
			} else {
				log.Printf("Received empty/unknown response from gRPC stream for user %s", user.ID)
				continue
			}

			// Write message back to WebSocket client
			if err := conn.WriteJSON(&wsResp); err != nil {
				log.Printf("WebSocket write error for user %s: %v", user.ID, err)
				return // Exit loop on write error
			}
			log.Printf("Sent '%s' to WebSocket for user %s", wsResp.Type, user.ID)
		}
	})(ctx)
}
