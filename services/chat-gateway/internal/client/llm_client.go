package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LLMGatewayClient handles communication with the llm-gateway service.
type LLMGatewayClient struct {
	httpClient *http.Client
	baseURL    string
}

// CompletionRequest defines the structure sent to the llm-gateway.
// Adjust this based on what llm-gateway will expect.
type CompletionRequest struct {
	ConversationID string `json:"conversation_id"`
	Prompt         string `json:"prompt"`
	// Add other fields like user_id, model preference, etc. if needed
}

// CompletionResponse defines the structure received from the llm-gateway.
// Adjust this based on what llm-gateway will return.
// For MVP, let's assume it returns the full completion text.
// Later, this might need to support streaming.
type CompletionResponse struct {
	Completion string `json:"completion"`
	// Add other fields like model used, token counts, etc.
}

// NewLLMGatewayClient creates a new client instance.
func NewLLMGatewayClient(baseURL string) *LLMGatewayClient {
	return &LLMGatewayClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // Set a reasonable timeout
		},
		baseURL: baseURL,
	}
}

// GetCompletion sends a prompt to the llm-gateway and gets a completion.
// This is a blocking call for MVP. Streaming would require a different approach.
func (c *LLMGatewayClient) GetCompletion(ctx context.Context, convID, prompt string) (*CompletionResponse, error) {
	requestURL := fmt.Sprintf("%s/api/v1/completion", c.baseURL) // Example endpoint

	reqBody := CompletionRequest{
		ConversationID: convID,
		Prompt:         prompt,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add any other required headers (e.g., internal auth token if needed)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to llm-gateway: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// TODO: Read error body for more details if available
		return nil, fmt.Errorf("llm-gateway returned non-OK status: %s", resp.Status)
	}

	var completionResp CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completionResp); err != nil {
		return nil, fmt.Errorf("failed to decode llm-gateway response: %w", err)
	}

	return &completionResp, nil
}
