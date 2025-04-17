package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/careerup-Inc/careerup-monorepo/services/avatar-service/internal/model"
)

const (
	vroidAPIBaseURL = "https://hub.vroid.com/api/v1"
)

type VRoidClientInterface interface {
	GenerateAvatar(ctx context.Context, req *model.AvatarGenerationRequest) (*model.Avatar, error)
	GetAvatar(ctx context.Context, id string) (*model.Avatar, error)
	UpdateAvatar(ctx context.Context, id string, req *model.AvatarUpdateRequest) (*model.Avatar, error)
	DeleteAvatar(ctx context.Context, id string) error
}

type VRoidClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewVRoidClient(apiKey string) *VRoidClient {
	return &VRoidClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateAvatar starts the avatar generation process
func (c *VRoidClient) GenerateAvatar(ctx context.Context, req *model.AvatarGenerationRequest) (*model.Avatar, error) {
	// Create the request body
	requestBody := map[string]interface{}{
		"style":    req.Style,
		"features": req.Features,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", vroidAPIBaseURL+"/avatars", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	var response struct {
		ID        string            `json:"id"`
		Style     string            `json:"style"`
		Features  map[string]string `json:"features"`
		ImageURL  string            `json:"image_url"`
		Status    string            `json:"status"`
		CreatedAt string            `json:"created_at"`
		UpdatedAt string            `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	createdAt, _ := time.Parse(time.RFC3339, response.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, response.UpdatedAt)

	return &model.Avatar{
		ID:        response.ID,
		Style:     response.Style,
		Features:  response.Features,
		ImageURL:  response.ImageURL,
		Status:    response.Status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// GetAvatar retrieves an avatar by ID
func (c *VRoidClient) GetAvatar(ctx context.Context, id string) (*model.Avatar, error) {
	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", vroidAPIBaseURL+"/avatars/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	var response struct {
		ID        string            `json:"id"`
		Style     string            `json:"style"`
		Features  map[string]string `json:"features"`
		ImageURL  string            `json:"image_url"`
		Status    string            `json:"status"`
		CreatedAt string            `json:"created_at"`
		UpdatedAt string            `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	createdAt, _ := time.Parse(time.RFC3339, response.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, response.UpdatedAt)

	return &model.Avatar{
		ID:        response.ID,
		Style:     response.Style,
		Features:  response.Features,
		ImageURL:  response.ImageURL,
		Status:    response.Status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// UpdateAvatar updates an existing avatar
func (c *VRoidClient) UpdateAvatar(ctx context.Context, id string, req *model.AvatarUpdateRequest) (*model.Avatar, error) {
	// Create the request body
	requestBody := map[string]interface{}{
		"style":    req.Style,
		"features": req.Features,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "PUT", vroidAPIBaseURL+"/avatars/"+id, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	var response struct {
		ID        string            `json:"id"`
		Style     string            `json:"style"`
		Features  map[string]string `json:"features"`
		ImageURL  string            `json:"image_url"`
		Status    string            `json:"status"`
		CreatedAt string            `json:"created_at"`
		UpdatedAt string            `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	createdAt, _ := time.Parse(time.RFC3339, response.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, response.UpdatedAt)

	return &model.Avatar{
		ID:        response.ID,
		Style:     response.Style,
		Features:  response.Features,
		ImageURL:  response.ImageURL,
		Status:    response.Status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// DeleteAvatar deletes an avatar
func (c *VRoidClient) DeleteAvatar(ctx context.Context, id string) error {
	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", vroidAPIBaseURL+"/avatars/"+id, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
