package handler

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email" example:"user@example.com"`
	Password  string `json:"password" binding:"required,min=8" example:"password123"`
	FirstName string `json:"first_name" binding:"required" example:"John"`
	LastName  string `json:"last_name" binding:"required" example:"Doe"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type UpdateUserRequest struct {
	Token     string   `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	FirstName string   `json:"first_name" example:"John"`
	LastName  string   `json:"last_name" example:"Doe"`
	Hometown  string   `json:"hometown" example:"New York"`
	Interests []string `json:"interests" example:"['AI', 'Machine Learning']"`
}

type ValidateTokenRequest struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"your_refresh_token_here"`
}

// ClientMessage defines the structure for messages received from the WebSocket client
type ClientMessage struct {
	Type           string `json:"type"` // e.g., "user_msg"
	ConversationID string `json:"conversation_id"`
	Text           string `json:"text"`
}

// ServerMessage defines the structure for messages sent to the WebSocket client
type ServerMessage struct {
	Type         string `json:"type"`            // e.g., "assistant_token", "avatar_url", "error"
	Token        string `json:"token,omitempty"` // For type="assistant_token"
	URL          string `json:"url,omitempty"`   // For type="avatar_url"
	ErrorMessage string `json:"error,omitempty"` // For type="error"
}
