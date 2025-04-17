package client

import (
	"context"
	"fmt"
	"time"

	"github.com/careerup-Inc/careerup-monorepo/services/avatar-service/internal/model"
)

type MockVRoidClient struct {
	avatars map[string]*model.Avatar
}

func NewMockVRoidClient() *MockVRoidClient {
	return &MockVRoidClient{
		avatars: make(map[string]*model.Avatar),
	}
}

func (c *MockVRoidClient) GenerateAvatar(ctx context.Context, req *model.AvatarGenerationRequest) (*model.Avatar, error) {
	if req.Style == "" {
		return nil, fmt.Errorf("style is required")
	}
	if req.Features == nil {
		return nil, fmt.Errorf("features are required")
	}

	// Generate a mock avatar
	avatar := &model.Avatar{
		ID:        fmt.Sprintf("mock-%d", time.Now().UnixNano()),
		Style:     req.Style,
		Features:  req.Features,
		ImageURL:  "https://example.com/mock-avatar.png",
		Status:    "ready",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store the avatar
	c.avatars[avatar.ID] = avatar

	return avatar, nil
}

func (c *MockVRoidClient) GetAvatar(ctx context.Context, id string) (*model.Avatar, error) {
	avatar, exists := c.avatars[id]
	if !exists {
		return nil, fmt.Errorf("avatar not found")
	}
	return avatar, nil
}

func (c *MockVRoidClient) UpdateAvatar(ctx context.Context, id string, req *model.AvatarUpdateRequest) (*model.Avatar, error) {
	avatar, exists := c.avatars[id]
	if !exists {
		return nil, fmt.Errorf("avatar not found")
	}

	// Update fields if provided
	if req.Style != "" {
		avatar.Style = req.Style
	}
	if req.Features != nil {
		avatar.Features = req.Features
	}
	avatar.UpdatedAt = time.Now()

	return avatar, nil
}

func (c *MockVRoidClient) DeleteAvatar(ctx context.Context, id string) error {
	if _, exists := c.avatars[id]; !exists {
		return fmt.Errorf("avatar not found")
	}
	delete(c.avatars, id)
	return nil
}
