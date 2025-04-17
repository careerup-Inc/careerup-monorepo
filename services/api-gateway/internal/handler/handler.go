package handler

import (
	"log"
	"net/http"

	v1 "github.com/careerup-Inc/careerup-monorepo/proto/careerup/v1"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// @title CareerUP API
// @version 1.0
// @description This is the CareerUP API server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Configure allowed origins
	},
}

type Handler struct {
	authClient *client.AuthClient
	chatClient *client.ChatClient
}

func NewHandler(authClient *client.AuthClient, chatClient *client.ChatClient) *Handler {
	return &Handler{
		authClient: authClient,
		chatClient: chatClient,
	}
}

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
	FirstName string   `json:"first_name" example:"John"`
	LastName  string   `json:"last_name" example:"Doe"`
	Hometown  string   `json:"hometown" example:"New York"`
	Interests []string `json:"interests" example:"['AI', 'Machine Learning']"`
}

type ValidateTokenRequest struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register Request"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.authClient.Register(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// @Summary Login user
// @Description Login user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Request"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.authClient.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get current user
// @Description Get the current authenticated user's profile
// @Tags user
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} User
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /me [get]
func (h *Handler) GetCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
		return
	}

	user, err := h.authClient.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary Update current user
// @Description Update the current authenticated user's profile
// @Tags user
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body UpdateUserRequest true "Update Request"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /me [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	updateReq := &v1.UpdateUserRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Hometown:  req.Hometown,
		Interests: req.Interests,
	}

	user, err := h.authClient.UpdateUser(c.Request.Context(), userID, updateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary WebSocket chat
// @Description WebSocket endpoint for real-time chat
// @Tags chat
// @Security ApiKeyAuth
// @Success 101 {string} string "Switching Protocols"
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /ws [get]
func (h *Handler) WebSocket(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "user not authenticated"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to upgrade connection"})
		return
	}
	defer conn.Close()

	// Create gRPC stream
	stream, err := h.chatClient.Stream(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create chat stream"})
		return
	}

	// Handle WebSocket messages
	for {
		var msg v1.WebSocketMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("Failed to read WebSocket message: %v", err)
			break
		}

		switch msg.Type {
		case "user_msg":
			if err := stream.Send(&v1.StreamRequest{
				ConversationId: msg.GetUserMessage().ConversationId,
				Text:           msg.GetUserMessage().Text,
			}); err != nil {
				log.Printf("Failed to send message: %v", err)
				break
			}
		}
	}
}

// @Summary Validate token
// @Description Validate an authentication token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ValidateTokenRequest true "Token to validate"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/validate [post]
func (h *Handler) ValidateToken(c *gin.Context) {
	var req ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.authClient.ValidateToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
		return
	}

	c.JSON(http.StatusOK, user)
}
