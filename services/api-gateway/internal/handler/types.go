package handler

import "github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"

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

// ILO Test Result submission

// IloAnswer represents a single answer in an ILO test
type IloAnswer struct {
	QuestionID     string `json:"question_id"`
	QuestionNumber int32  `json:"question_number"`
	SelectedOption int32  `json:"selected_option"`
}

type IloTestResultRequest struct {
	ResultData string      `json:"result_data" example:"{\"score\":85,\"details\":{...}}"`
	Answers    []IloAnswer `json:"answers,omitempty"`
}

type IloTestResultResponse struct {
	ID               string                  `json:"id"`
	UserID           string                  `json:"user_id"`
	ResultData       string                  `json:"result_data"`
	CreatedAt        string                  `json:"created_at"`
	Scores           []client.IloDomainScore `json:"scores,omitempty"`
	TopDomains       []string                `json:"top_domains,omitempty"`
	SuggestedCareers []string                `json:"suggested_careers,omitempty"`
}

type IloTestResultAnalysisResponse struct {
	Result   IloTestResultResponse `json:"result"`
	Analysis string                `json:"analysis"`
}

// IloDomain represents one of the 5 domains assessed in the ILO test
type IloDomain struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// IloLevel represents the evaluation level for a domain score
type IloLevel struct {
	MinPercent int32  `json:"min_percent"`
	MaxPercent int32  `json:"max_percent"`
	LevelName  string `json:"level_name"`
	Suggestion string `json:"suggestion,omitempty"`
}

// IloTestQuestion represents a single question in the ILO test
type IloTestQuestion struct {
	ID             string   `json:"id"`
	QuestionNumber int32    `json:"question_number"`
	Text           string   `json:"text"`
	DomainCode     string   `json:"domain_code,omitempty"`
	Options        []string `json:"options"`
}

// GetIloTestResponse is the response for ILO test questions
type GetIloTestResponse struct {
	Questions []IloTestQuestion `json:"questions"`
	Domains   []IloDomain       `json:"domains,omitempty"`
	Levels    []IloLevel        `json:"levels,omitempty"`
}
