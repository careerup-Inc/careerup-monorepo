package server

import (
	"io"
	"log"
	"strings"
	"time"

	"os"

	chatpb "github.com/careerup-Inc/careerup-monorepo/proto/careerup/v1"
	"github.com/careerup-Inc/careerup-monorepo/services/chat-gateway/internal/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	// TODO: Add client for llm-gateway when implemented
)

// GrpcServer implements the ConversationServiceServer interface.
type GrpcServer struct {
	chatpb.UnimplementedConversationServiceServer
	llmClient *client.LLMGatewayClient
}

// NewGrpcServer creates a new instance of the gRPC server.
func NewGrpcServer() *GrpcServer {
	// TODO: Get base URL from config/env
	llmGatewayAddr := os.Getenv("LLM_SERVICE_ADDR")
	if llmGatewayAddr == "" {
		llmGatewayAddr = "http://llm-gateway:9090"
		log.Printf("Warning: LLM_SERVICE_ADDR not set, using default %s", llmGatewayAddr)
	} else {
		if !strings.HasPrefix(llmGatewayAddr, "http://") && !strings.HasPrefix(llmGatewayAddr, "https://") {
			llmGatewayAddr = "http://" + llmGatewayAddr
		}
	}

	return &GrpcServer{
		// Initialize llmGatewayClient
		llmClient: client.NewLLMGatewayClient(llmGatewayAddr),
	}
}

// Stream handles the bidirectional communication for chat.
func (s *GrpcServer) Stream(stream chatpb.ConversationService_StreamServer) error {
	log.Println("Stream started")
	ctx := stream.Context()

	// Channel to signal when llm response is ready to be sent
	llmResponseChan := make(chan *client.CompletionResponse)
	errChan := make(chan error, 1) // Buffered error channel

	// Goroutine to handle receiving messages from the client (api-gateway)
	// and potentially forwarding them to the llm-gateway
	go func() {
		defer close(errChan)
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				log.Println("Stream closed by client (EOF)")
				return
			}
			if err != nil {
				st, ok := status.FromError(err)
				if ok && st.Code() == codes.Canceled {
					log.Println("Stream cancelled by client")
				} else {
					log.Printf("Error receiving from stream: %v", err)
					errChan <- err
				}
				return
			}

			log.Printf("Received request: conv_id=%s, text='%s'", req.ConversationId, req.Text)

			// --- TODO: Call llm-gateway ---
			// 1. Create a request for llm-gateway based on req.Text and req.ConversationId
			// 2. Call the llm-gateway client method
			// 3. Handle the response (which might itself be a stream or a single response)
			// 4. For MVP, we'll just simulate a response below in the sending part.
			// --- End TODO ---
			// Use a separate context for the HTTP call if needed, or reuse stream context
			go func(convID, text string) {
				llmResp, err := s.llmClient.GetCompletion(ctx, convID, text)
				if err != nil {
					log.Printf("Error calling llm-gateway: %v", err)
					// Decide how to handle LLM errors - maybe send an error token back?
					// For now, just log it. Could send to errChan if needed.
					return
				}
				log.Printf("LLM Gateway response: %s", llmResp.Completion)
				llmResponseChan <- llmResp
			}(req.ConversationId, req.Text)

			// Select allows checking context cancellation while waiting
			select {
			case <-ctx.Done():
				log.Println("Context cancelled, stopping receive loop.")
				return
			default:
				// Continue processing
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context done, closing stream from server side. Error: %v", ctx.Err())
			return ctx.Err()
		case llmResp := <-llmResponseChan:
			// Received a response from the llm-gateway via the channel
			// TODO: Implement token streaming if llm-gateway supports it.
			// For now, send the whole completion as one token message.
			resp := &chatpb.StreamResponse{
				Response: &chatpb.StreamResponse_Token{Token: llmResp.Completion},
			}
			if err := stream.Send(resp); err != nil {
				log.Printf("Error sending response to stream: %v", err)
				return err
			}
			log.Printf("Sent LLM response to stream.")
		case err := <-errChan:
			log.Printf("Exiting stream due to receive error: %v", err)
			return err // Propagate the error
		case <-time.After(30 * time.Second):
			log.Println("Stream timed out due to inactivity.")
			return status.Error(codes.DeadlineExceeded, "Stream timed out")
		}
	}
}
