package model

import "time"

// Avatar represents a VRoid Studio avatar
type Avatar struct {
	ID        string            `json:"id"`
	Style     string            `json:"style"`
	Features  map[string]string `json:"features"`
	ImageURL  string            `json:"image_url"`
	Status    string            `json:"status"` // pending, generating, ready, error
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// AvatarGenerationRequest represents a request to generate a new avatar
type AvatarGenerationRequest struct {
	Style    string            `json:"style"`
	Features map[string]string `json:"features"`
}

// AvatarUpdateRequest represents a request to update an existing avatar
type AvatarUpdateRequest struct {
	Style    string            `json:"style,omitempty"`
	Features map[string]string `json:"features,omitempty"`
}
