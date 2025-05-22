package service

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// OpenAIService handles interactions with the OpenAI API via LangChainGo.
type OpenAIService struct {
	llm llms.Model // Use the LangChainGo LLM interface
}

// NewOpenAIService creates a new service instance using LangChainGo.
func NewOpenAIService() (*OpenAIService, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// Create a new LangChainGo OpenAI client
	llm, err := openai.New(openai.WithToken(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create LangChainGo OpenAI client: %w", err)
	}

	return &OpenAIService{llm: llm}, nil
}

// GetCompletion sends a prompt using LangChainGo and returns the completion.
func (s *OpenAIService) GetCompletion(ctx context.Context, prompt string) (string, error) {
	completion, err := llms.GenerateFromSinglePrompt(ctx, s.llm, prompt)
	if err != nil {
		log.Printf("LangChainGo OpenAI GenerateFromSinglePrompt error: %v", err)
		return "", fmt.Errorf("failed to get completion from LangChainGo OpenAI: %w", err)
	}
	return completion, nil
}
