package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	pbllm "github.com/careerup-Inc/careerup-monorepo/proto/llm/v1"
)

// LLMServiceImpl implements the LLMService gRPC interface.
type LLMServiceImpl struct {
	pbllm.UnimplementedLLMServiceServer
	llm llms.Model // Use langchaingo LLM interface
}

// NewLLMService creates a new LLMService implementation using langchaingo.
func NewLLMService() (*LLMServiceImpl, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Return an error instead of just logging a warning
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// Initialize langchaingo OpenAI client
	llm, err := openai.New(openai.WithModel("gpt-3.5-turbo"), openai.WithToken(apiKey)) // Use string model name
	if err != nil {
		log.Printf("Failed to initialize langchaingo OpenAI client: %v", err)
		return nil, err
	}

	return &LLMServiceImpl{llm: llm}, nil
}

// GenerateStream handles the streaming request from chat-gateway using langchaingo's GenerateContent.
func (s *LLMServiceImpl) GenerateStream(req *pbllm.GenerateRequest, stream pbllm.LLMService_GenerateStreamServer) error {
	log.Printf("LLM GenerateStream request received: UserID=%s, ConvID=%s", req.UserId, req.ConversationId)

	ctx, cancel := context.WithTimeout(stream.Context(), 120*time.Second)
	defer cancel()

	// Prepare options for langchaingo streaming call
	options := []llms.CallOption{
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			token := string(chunk)
			// log.Printf("Sending token chunk: %s", token) // Can be noisy
			grpcRes := &pbllm.GenerateResponse{Token: token}
			if err := stream.Send(grpcRes); err != nil {
				log.Printf("gRPC stream send error: %v", err)
				return err
			}
			return nil // Indicate success for this chunk
		}),
		// Add other options like temperature, max tokens if needed
		// llms.WithTemperature(0.7),
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
