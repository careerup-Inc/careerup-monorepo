package handler

import (
	"log"
	"strings"

	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	utils "github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/utils"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

type Handler struct {
	authClient client.AuthClientInterface
	chatClient client.ChatClientInterface
}

func NewHandler(authClient *client.AuthClient, chatClient *client.ChatClient) *Handler {
	return &Handler{
		authClient: authClient,
		chatClient: chatClient,
	}
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
// @Router /api/v1/auth/register [post]
func (h *Handler) HandleRegister(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// Call auth service to register user
	user, err := h.authClient.Register(c.Context(), &client.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})

	if err != nil {
		// Map gRPC errors
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid registration data: "+st.Message())
			case codes.AlreadyExists:
				return utils.SendErrorResponse(c, fiber.StatusConflict, "User already exists: "+st.Message())
			case codes.Unavailable:
				return utils.SendErrorResponse(c, fiber.StatusServiceUnavailable, "Auth service unavailable: "+st.Message())
			default:
				log.Printf("Unhandled gRPC error during registration: %v", err)
				return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Registration failed: "+st.Message())
			}
		}
		// Handle non-gRPC errors
		log.Printf("Non-gRPC error during registration: %v", err)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Registration failed: "+err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(user)
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
// @Router /api/v1/auth/login [post]
func (h *Handler) HandleLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	loginResp, err := h.authClient.Login(c.Context(), &client.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		// Map gRPC errors
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid login data: "+st.Message())
			case codes.Unauthenticated:
				return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid credentials: "+st.Message())
			case codes.NotFound: // Assuming NotFound might mean user doesn't exist
				return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User not found: "+st.Message())
			case codes.Unavailable:
				return utils.SendErrorResponse(c, fiber.StatusServiceUnavailable, "Auth service unavailable: "+st.Message())
			default:
				log.Printf("Unhandled gRPC error during login: %v", err)
				return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Login failed: "+st.Message())
			}
		}
		// Handle non-gRPC errors (like Fiber errors if client returns them)
		if fiberErr, ok := err.(*fiber.Error); ok {
			return utils.SendErrorResponse(c, fiberErr.Code, fiberErr.Message)
		}
		log.Printf("Non-gRPC error during login: %v", err)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Login failed: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(loginResp)
}

// @Summary Refresh authentication token
// @Description Provides new access and refresh tokens using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} client.TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *Handler) HandleRefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if req.RefreshToken == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "refresh_token is required")
	}

	tokens, err := h.authClient.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		// Map gRPC errors
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated: // Treat invalid/expired refresh token as Unauthenticated
				return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid or expired refresh token: "+st.Message())
			case codes.InvalidArgument:
				return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format: "+st.Message())
			case codes.Unavailable:
				return utils.SendErrorResponse(c, fiber.StatusServiceUnavailable, "Auth service unavailable: "+st.Message())
			default:
				log.Printf("Unhandled gRPC error during refresh token: %v", err)
				// Default to Unauthorized for safety with refresh tokens
				return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Token refresh failed: "+st.Message())
			}
		}
		// Handle non-gRPC errors
		log.Printf("Non-gRPC error during refresh token: %v", err)
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Token refresh failed: "+err.Error()) // Default to 401
	}

	// Return the new token response (contains new access_token, refresh_token, expires_in)
	return c.Status(fiber.StatusOK).JSON(tokens)
}

// @Summary Get current user
// @Description Get the current authenticated user's profile
// @Tags user
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} User
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/profile [get]
func (h *Handler) HandleGetProfile(c *fiber.Ctx) error {
	userLocal := c.Locals("user")
	if userLocal == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User not found in context (middleware issue?)")
	}
	user, ok := userLocal.(*client.User)
	if !ok || user == nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Invalid user data in context")
	}
	return c.Status(fiber.StatusOK).JSON(user)
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
// @Router /api/v1/profile [put]
func (h *Handler) HandleUpdateProfile(c *fiber.Ctx) error {
	// Get user from context (set by auth middleware)
	userLocal := c.Locals("user")
	if userLocal == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User not found in context")
	}
	authUser, ok := userLocal.(*client.User)
	if !ok || authUser == nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Invalid user data in context")
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// Call auth service to update user
	updatedUser, err := h.authClient.UpdateUser(c.Context(), &client.UpdateUserRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Hometown:  req.Hometown,
		Interests: req.Interests,
	})

	if err != nil {
		// Map gRPC errors
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid update data: "+st.Message())
			case codes.NotFound: // User to update not found (shouldn't happen if token is valid)
				return utils.SendErrorResponse(c, fiber.StatusNotFound, "User not found for update: "+st.Message())
			case codes.Unauthenticated: // Should be caught by middleware, but handle defensively
				return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Authentication required: "+st.Message())
			case codes.Unavailable:
				return utils.SendErrorResponse(c, fiber.StatusServiceUnavailable, "Auth service unavailable: "+st.Message())
			default:
				log.Printf("Unhandled gRPC error during UpdateProfile: %v", err)
				return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update profile: "+st.Message())
			}
		}
		// Handle non-gRPC errors
		log.Printf("Non-gRPC error during UpdateProfile: %v", err)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update profile: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(updatedUser)
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
// @Router /api/v1/auth/validate [post]
func (h *Handler) HandleValidateToken(c *fiber.Ctx) error {
	var req ValidateTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	user, err := h.authClient.ValidateToken(c.Context(), req.Token)
	if err != nil {
		// Map gRPC errors
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated: // Invalid/expired token
				return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid or expired token: "+st.Message())
			case codes.InvalidArgument:
				return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid token format: "+st.Message())
			case codes.Unavailable:
				return utils.SendErrorResponse(c, fiber.StatusServiceUnavailable, "Auth service unavailable: "+st.Message())
			default:
				log.Printf("Unhandled gRPC error during token validation: %v", err)
				return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Token validation failed: "+st.Message())
			}
		}
		// Handle non-gRPC errors
		log.Printf("Non-gRPC error during token validation: %v", err)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Token validation failed: "+err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// @Summary WebSocket chat
// @Description WebSocket endpoint for real-time chat
// @Tags chat
// @Security ApiKeyAuth
// @Success 101 {string} string "Switching Protocols"
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ws [get]
func (h *Handler) HandleWebSocket(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization") // Get header directly
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Authorization header with Bearer token is required")
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate token
	user, err := h.authClient.ValidateToken(c.Context(), token)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated:
				return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid or expired token for WebSocket: "+st.Message())
			case codes.InvalidArgument:
				return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid token format for WebSocket: "+st.Message())
			case codes.Unavailable:
				return utils.SendErrorResponse(c, fiber.StatusServiceUnavailable, "Auth service unavailable for WebSocket: "+st.Message())
			default:
				log.Printf("Unhandled gRPC error during WebSocket auth: %v", err)
				return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "WebSocket authentication failed: "+st.Message())
			}
		}
		log.Printf("Non-gRPC error during WebSocket auth: %v", err)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "WebSocket authentication failed: "+err.Error())
	}

	// Upgrade to WebSocket connection
	if err := h.chatClient.UpgradeToWebSocket(c, user); err != nil {
		// UpgradeToWebSocket might return an error if the upgrade itself fails
		// or if the initial connection to the backend chat service fails.
		log.Printf("WebSocket upgrade/connection failed for user %s: %v", user.ID, err)
		// Don't return JSON here if the upgrade failed, the connection might be hijacked.
		// The error might have already been logged within UpgradeToWebSocket.
		// Fiber's websocket.New handles sending the 101 internally on success.
		// If UpgradeToWebSocket returns an error *before* the upgrade, we could send JSON.
		// Let's assume UpgradeToWebSocket handles logging and potential initial error responses.
		return err // Propagate the error if necessary
	}

	return nil
}
