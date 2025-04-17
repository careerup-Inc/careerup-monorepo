package service

import (
	"context"
	"errors"

	"github.com/careerup-Inc/careerup-monorepo/services/avatar-service/internal/client"
	"github.com/careerup-Inc/careerup-monorepo/services/avatar-service/internal/model"
)

type AvatarService struct {
	vroidClient *client.VRoidClient
}

func NewAvatarService(vroidClient *client.VRoidClient) *AvatarService {
	return &AvatarService{
		vroidClient: vroidClient,
	}
}

func (s *AvatarService) GenerateAvatar(ctx context.Context, req *model.AvatarGenerationRequest) (*model.Avatar, error) {
	if req.Style == "" {
		return nil, errors.New("style is required")
	}
	if req.Features == nil {
		return nil, errors.New("features are required")
	}

	return s.vroidClient.GenerateAvatar(ctx, req)
}

func (s *AvatarService) GetAvatar(ctx context.Context, id string) (*model.Avatar, error) {
	if id == "" {
		return nil, errors.New("avatar ID is required")
	}

	return s.vroidClient.GetAvatar(ctx, id)
}

func (s *AvatarService) UpdateAvatar(ctx context.Context, id string, req *model.AvatarUpdateRequest) (*model.Avatar, error) {
	if id == "" {
		return nil, errors.New("avatar ID is required")
	}

	return s.vroidClient.UpdateAvatar(ctx, id, req)
}

func (s *AvatarService) DeleteAvatar(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("avatar ID is required")
	}

	return s.vroidClient.DeleteAvatar(ctx, id)
}
