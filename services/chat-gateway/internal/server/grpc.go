package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	pbChat "github.com/careerup-Inc/careerup-monorepo/proto/careerup/v1"
	pbllm "github.com/careerup-Inc/careerup-monorepo/proto/llm/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/careerup-Inc/careerup-monorepo/services/chat-gateway/internal/client"
)

// ChatServer implements the ConversationService gRPC interface.
type ChatServer struct {
	pbChat.UnimplementedConversationServiceServer                   // Embed the unimplemented server
	llmClient                                     *client.LLMClient // Use the gRPC client wrapper
	iloClient                                     *client.IloClient // ILO client for user context
}

// NewChatServer creates a new chat server instance.
func NewChatServer(llmClient *client.LLMClient, iloClient *client.IloClient) *ChatServer {
	return &ChatServer{
		llmClient: llmClient,
		iloClient: iloClient,
	}
}

// Stream handles the bidirectional stream between api-gateway and chat-gateway.
func (s *ChatServer) Stream(stream pbChat.ConversationService_StreamServer) error {
	log.Println("Chat stream established with a client (api-gateway)")
	ctx := stream.Context()

	// Extract user-id from incoming context (set by api-gateway)
	md, ok := metadata.FromIncomingContext(ctx)
	userID := "unknown"
	if ok && len(md.Get("user-id")) > 0 {
		userID = md.Get("user-id")[0]
	}
	log.Printf("User ID from metadata: %s", userID)

	// Channel to signal when LLM processing for a message is done
	llmDone := make(chan struct{}, 1) // Buffered channel to avoid blocking sender

	// Goroutine to handle receiving messages from the client (api-gateway)
	// and triggering LLM calls.
	go func() {
		defer close(llmDone) // Ensure channel is closed when this goroutine exits
		for {
			// Check if the client context is cancelled first
			select {
			case <-ctx.Done():
				log.Printf("Client stream context cancelled: %v", ctx.Err())
				return // Exit goroutine if client disconnected
			default:
				// Proceed to receive message
			}

			req, err := stream.Recv()
			if err == io.EOF {
				log.Println("Client (api-gateway) closed the send stream.")
				return // Client closed the connection stream
			}
			if err != nil {
				// Handle specific gRPC errors if needed
				st, ok := status.FromError(err)
				if ok && st.Code() == codes.Canceled {
					log.Println("Client stream cancelled.")
				} else {
					log.Printf("Error receiving message from client stream: %v", err)
				}
				return // Terminate this goroutine on error
			}

			// Validate message type (add more checks as needed)
			if req.Type != "user_msg" || req.Text == "" {
				log.Printf("Received invalid message type or empty text: Type=%s", req.Type)
				errMsg := &pbChat.StreamResponse{
					Type:    "error",
					Content: &pbChat.StreamResponse_ErrorMessage{ErrorMessage: "Invalid message format"},
				}
				if sendErr := stream.Send(errMsg); sendErr != nil {
					log.Printf("Failed to send error message back to api-gateway: %v", sendErr)
					return // Assume connection is broken
				}
				continue // Wait for next valid message
			}

			log.Printf("Received user_msg from api-gateway: ConvID=%s", req.ConversationId)

			// Fetch latest ILO test result for user (if available)
			var iloContext string
			if s.iloClient != nil && userID != "unknown" {
				result, err := s.iloClient.GetLatestIloTestResult(ctx, userID)
				if err != nil {
					log.Printf("Failed to fetch ILO test result for user %s: %v", userID, err)
				} else if result != nil {
					// Compose a context string from ILO result (top domains, scores, suggestions)
					iloContext = "User ILO profile: "
					if len(result.TopDomains) > 0 {
						iloContext += "Top domains: " + strings.Join(result.TopDomains, ", ") + ". "
					}
					if len(result.SuggestedCareers) > 0 {
						iloContext += "Suggested careers: " + strings.Join(result.SuggestedCareers, ", ") + ". "
					}
					if len(result.Scores) > 0 {
						scoreStrs := make([]string, 0, len(result.Scores))
						for _, s := range result.Scores {
							scoreStrs = append(scoreStrs, s.DomainCode+":"+fmt.Sprintf("%.0f%%", s.Percent))
						}
						iloContext += "Domain scores: " + strings.Join(scoreStrs, ", ") + ". "
					}
				}
			}

			// --- Trigger LLM Streaming Call with RAG ---
			llmReq := &pbllm.GenerateWithRAGRequest{
				Prompt:         iloContext + req.Text,
				UserId:         userID,
				ConversationId: req.ConversationId,
				RagCollection:  "academy", // Use "academy" or latest collection for RAG
				Adaptive:       true,      // Enable adaptive RAG
			}

			llmCtx, llmCancel := context.WithTimeout(ctx, 60*time.Second)
			log.Println("Calling LLMService.GenerateWithRAG...")
			llmStream, err := s.llmClient.GetLLMServiceClient().GenerateWithRAG(llmCtx, llmReq)
			if err != nil {
				log.Printf("Failed to start LLM RAG stream: %v", err)
				llmCancel()
				errMsg := &pbChat.StreamResponse{
					Type:    "error",
					Content: &pbChat.StreamResponse_ErrorMessage{ErrorMessage: "Failed to connect to LLM RAG service"},
				}
				if sendErr := stream.Send(errMsg); sendErr != nil {
					log.Printf("Failed to send error message back to api-gateway: %v", sendErr)
					return // Assume connection is broken
				}
				continue
			}

			log.Println("LLM RAG stream started, receiving tokens...")
			var llmReceiveErr error
			for {
				llmRes, err := llmStream.Recv()
				if err == io.EOF {
					log.Println("LLM RAG stream ended.")
					break
				}
				if err != nil {
					st, ok := status.FromError(err)
					if ok && st.Code() == codes.Canceled {
						log.Println("LLM RAG stream context cancelled.")
					} else {
						log.Printf("Error receiving from LLM RAG stream: %v", err)
						llmReceiveErr = err
					}
					break
				}
				chatRes := &pbChat.StreamResponse{
					Type:    "assistant_token",
					Content: &pbChat.StreamResponse_Token{Token: llmRes.Token},
				}
				if err := stream.Send(chatRes); err != nil {
					log.Printf("Error sending token to api-gateway stream: %v", err)
					llmCancel()
					return
				}
			}
			llmCancel()
			if llmReceiveErr != nil {
				errMsg := &pbChat.StreamResponse{
					Type:    "error",
					Content: &pbChat.StreamResponse_ErrorMessage{ErrorMessage: "Error receiving response from LLM RAG"},
				}
				if sendErr := stream.Send(errMsg); sendErr != nil {
					log.Printf("Failed to send LLM error message back to api-gateway: %v", sendErr)
					return
				}
			}
			// --- End LLM RAG Streaming Call ---

			// TODO: Add Avatar Service call here if needed, send avatar_url message
			// Example:
			// avatarURL := getAvatarURL(req.ConversationId, ...) // Call avatar service
			// avatarMsg := &pbChat.StreamResponse{
			// 	Type: "avatar_url",
			// 	Content: &pbChat.StreamResponse_Url{Url: avatarURL},
			// }
			// if err := stream.Send(avatarMsg); err != nil { ... }

			// Signal that processing for this message is done (optional, might be useful for flow control)
			// llmDone <- struct{}{}
		}
	}()

	// Keep the main stream handler alive. It will exit when:
	// 1. The client context (ctx) is Done (client disconnected).
	// 2. The receiving goroutine exits (due to client closing stream or error).
	select {
	case <-ctx.Done():
		log.Printf("Chat stream context done (client disconnected): %v", ctx.Err())
	case <-llmDone:
		log.Println("Chat stream processing goroutine finished.")
	}

	return ctx.Err() // Return the context error, if any
}
