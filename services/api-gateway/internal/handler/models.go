package handler

// User represents a user in the system
type User struct {
	ID        string   `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email     string   `json:"email" example:"user@example.com"`
	FirstName string   `json:"first_name" example:"John"`
	LastName  string   `json:"last_name" example:"Doe"`
	Hometown  string   `json:"hometown" example:"New York"`
	Interests []string `json:"interests" example:"['AI', 'Machine Learning']"`
}

// LoginResponse represents the response from a login request
type LoginResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User         User   `json:"user"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}
