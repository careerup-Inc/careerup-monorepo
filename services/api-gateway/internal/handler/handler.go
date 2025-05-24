package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	utils "github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/utils"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pbChat "github.com/careerup-Inc/careerup-monorepo/proto/careerup/v1"
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

// @securityDefinitions.BearerAuth
// @type http
// @scheme bearer
// @bearerFormat JWT

type Handler struct {
	authClient client.AuthClientInterface
	chatClient client.ChatClientInterface
	// ILO gRPC client
	IloClient *client.IloClient
	LLMClient *client.LLMClient
	// Auth core service address for REST calls
	authCoreServiceAddr string
}

func NewHandler(authClient client.AuthClientInterface, chatClient client.ChatClientInterface, iloClient *client.IloClient, llmClient *client.LLMClient, authCoreAddr string) *Handler {
	return &Handler{
		authClient:          authClient,
		chatClient:          chatClient,
		IloClient:           iloClient,
		LLMClient:           llmClient,
		authCoreServiceAddr: authCoreAddr,
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
// @Security BearerAuth
// @Success 200 {object} User
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/user/me [get]
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
// @Security BearerAuth
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

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Authorization header is required")
	}
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// Call auth service to update user
	updatedUser, err := h.authClient.UpdateUser(c.Context(), &client.UpdateUserRequest{
		Token:     token,
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
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/validate [get]
func (h *Handler) HandleValidateToken(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Authorization header is required")
	}

	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	user, err := h.authClient.ValidateToken(c.Context(), token)
	if err != nil {
		// Map gRPC errors
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated:
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
// @Security BearerAuth
// @Success 101 {string} string "Switching Protocols"
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ws [get]
func (h *Handler) HandleWebSocket(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authorization header is required"})
		}
		token := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
		user, err := h.authClient.ValidateToken(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}
		c.Locals("userID", user.ID)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// WebSocketProxy handles the persistent WebSocket connection after upgrade.
func (h *Handler) WebSocketProxy(conn *websocket.Conn) {
	defer func() {
		log.Println("Closing WebSocket connection")
		conn.Close()
	}()

	// Retrieve user ID from locals set during the upgrade
	userID := conn.Locals("userID").(string)
	log.Printf("WebSocket connection established for user: %s", userID)

	// --- gRPC Stream Setup ---
	md := metadata.Pairs("user-id", userID)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	// Add cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // Ensure cancellation happens on function exit

	// Establish gRPC stream with chat-gateway
	stream, err := h.chatClient.GetChatServiceClient().Stream(ctx)
	if err != nil {
		log.Printf("Failed to establish gRPC stream with chat-gateway: %v", err)
		_ = conn.WriteJSON(ServerMessage{Type: "error", ErrorMessage: "Failed to connect to chat service"})
		return
	}
	log.Println("gRPC stream established with chat-gateway")

	// Goroutine to read from gRPC stream and write to WebSocket
	go func() {
		defer log.Println("Exiting gRPC read goroutine")
		for {
			res, err := stream.Recv()
			if err != nil {
				// Handle different kinds of errors
				st, ok := status.FromError(err)
				if ok {
					if st.Code() == codes.Canceled {
						log.Println("gRPC stream context cancelled (likely client disconnect)")
					} else {
						log.Printf("gRPC stream receive error: %v, code: %s", err, st.Code())
						// Send error to WebSocket client if connection is still likely open
						_ = conn.WriteJSON(ServerMessage{Type: "error", ErrorMessage: "Chat service connection error"})
					}
				} else if err == io.EOF {
					log.Println("gRPC stream closed by chat-gateway (EOF)")
				} else {
					log.Printf("gRPC stream receive error (non-gRPC): %v", err)
					_ = conn.WriteJSON(ServerMessage{Type: "error", ErrorMessage: "Chat service communication error"})
				}
				cancel() // Cancel context to potentially stop the write loop below
				return   // Exit goroutine
			}

			// Construct message based on gRPC response type
			var msg ServerMessage
			switch res.Type {
			case "assistant_token":
				if tokenContent := res.GetToken(); tokenContent != "" {
					msg = ServerMessage{Type: "assistant_token", Token: tokenContent}
				} else {
					log.Println("Received assistant_token with empty content")
					continue
				}
			case "avatar_url":
				if urlContent := res.GetUrl(); urlContent != "" {
					msg = ServerMessage{Type: "avatar_url", URL: urlContent}
				} else {
					log.Println("Received avatar_url with empty content")
					continue
				}
			case "error":
				if errorContent := res.GetErrorMessage(); errorContent != "" {
					msg = ServerMessage{Type: "error", ErrorMessage: errorContent}
				} else {
					log.Println("Received error with empty content")
					continue
				}
			default:
				log.Printf("Unknown message type from gRPC: %s", res.Type)
				continue // Skip unknown types
			}

			// Write the message to the WebSocket client
			if err := conn.WriteJSON(msg); err != nil {
				log.Printf("WebSocket write error: %v", err)
				// Assume client disconnected, cancel context to close gRPC stream
				cancel()
				return // Exit goroutine
			}
			// log.Printf("Sent message to WebSocket: Type=%s", msg.Type) // Can be noisy
		}
	}()

	// --- WebSocket Read Loop ---
	log.Println("Starting WebSocket read loop")
	for {
		messageType, msgBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error (unexpected close): %v", err)
			} else {
				log.Printf("WebSocket read error: %v", err)
			}
			cancel() // Close gRPC stream
			break    // Exit read loop
		}

		if messageType == websocket.TextMessage {
			// log.Printf("Received message from WebSocket: %s", string(msgBytes))
			var clientMsg ClientMessage
			if err := json.Unmarshal(msgBytes, &clientMsg); err != nil {
				log.Printf("Failed to unmarshal client message: %v", err)
				_ = conn.WriteJSON(ServerMessage{Type: "error", ErrorMessage: "Invalid message format"})
				continue
			}

			// Basic validation
			if clientMsg.Type != "user_msg" || clientMsg.Text == "" {
				log.Printf("Invalid client message type or empty text: Type=%s", clientMsg.Type)
				_ = conn.WriteJSON(ServerMessage{Type: "error", ErrorMessage: "Invalid message type or empty text"})
				continue
			}

			// Send to gRPC stream
			grpcReq := &pbChat.StreamRequest{
				Type:           clientMsg.Type,
				ConversationId: clientMsg.ConversationID,
				Text:           clientMsg.Text,
			}
			if err := stream.Send(grpcReq); err != nil {
				log.Printf("gRPC stream send error: %v", err)
				// Assume gRPC stream is broken, send error and close connection
				_ = conn.WriteJSON(ServerMessage{Type: "error", ErrorMessage: "Failed to send message to chat service"})
				cancel()
				break // Exit read loop
			}
			// log.Printf("Sent message to gRPC: Type=%s", grpcReq.Type)
		} else {
			log.Printf("Received non-text message type: %d", messageType)
		}
	}
	log.Println("Exiting WebSocket read loop")
}

// @Summary Submit ILO test result
// @Description Submit ILO test result for the authenticated user and get analysis
// @Tags ilo
// @Accept json
// @Produce json
// @Param request body IloTestResultRequest true "ILO Test Result Request"
// @Success 201 {object} IloTestResultResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/ilo/result [post]
func (h *Handler) HandleIloTestResult(c *fiber.Ctx) error {
	var req IloTestResultRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	token := utils.ExtractTokenFromHeader(c)
	if token == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Missing or invalid Authorization header")
	}

	// Save ILO result via gRPC to ILO service
	user, err := h.authClient.ValidateToken(c.Context(), token)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid token: "+err.Error())
	}

	// Parse answers from request if they exist
	var answers []client.IloAnswer
	if len(req.Answers) > 0 {
		answers = make([]client.IloAnswer, len(req.Answers))
		for i, ans := range req.Answers {
			answers[i] = client.IloAnswer{
				QuestionID:     ans.QuestionID,
				QuestionNumber: ans.QuestionNumber,
				SelectedOption: ans.SelectedOption,
			}
		}
	}

	result, err := h.IloClient.SubmitILOTestResult(c.Context(), &client.SubmitILOTestResultRequest{
		UserID:        user.ID,
		Answers:       answers,
		RawResultData: req.ResultData,
	})
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to save ILO test result: "+err.Error())
	}

	// Create a rich prompt for LLM analysis with structured data
	// Build an expert‑level prompt so the LLM answers like a seasoned career‑guidance counsellor
	promptLines := []string{
		"You are a certified Vietnamese career counsellor who specialises in interpreting ILO tests for high-school students and parents.",
		"You are a certified career guidance expert with deep knowledge of the Vietnamese ILO (Interest, Learning, Orientation) framework.",
		"You are a friendly, slightly cheeky career-guidance guru who sprinkles gentle humour into professional advice.",
		"Analyse the candidate’s ILO result and produce a report in vietnamese with the following sections:",
		"1. Brief narrative overview of the candidate’s dominant interest profile.",
		"2. Key strengths and potential development areas, illustrated with concrete examples.",
		"3. Three to five career pathways that fit the profile, each followed by a one‑sentence rationale.",
		"4. Actionable next steps for the candidate over the next 3–6 months (courses, extracurriculars, shadowing, mentorship, etc.).",
		"ILO Domain Scores:",
	}

	for _, s := range result.Scores {
		promptLines = append(promptLines, fmt.Sprintf("- %s: %.1f%% (%s)", s.DomainCode, s.Percent, s.Level))
	}

	if len(result.TopDomains) > 0 {
		promptLines = append(promptLines, "",
			"Top domains: "+strings.Join(result.TopDomains, ", "))
	}

	// Personalise advice if basic user context is available
	if user.FirstName != "" {
		promptLines = append(promptLines, "",
			"Candidate first name: "+user.FirstName)
	}

	promptLines = append(promptLines, "",
		"Raw ILO data: "+req.ResultData)

	llmPrompt := strings.Join(promptLines, "\n")

	llmAnalysis, err := h.LLMClient.AnalyzeILOResult(c.Context(), &client.LLMAnalysisRequest{
		Prompt: llmPrompt,
		UserID: user.ID,
	})
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to analyze ILO test result: "+err.Error())
	}

	resp := IloTestResultResponse{
		ID:               result.ID,
		UserID:           result.UserID,
		ResultData:       result.ResultData,
		CreatedAt:        result.CreatedAt,
		Scores:           result.Scores,
		TopDomains:       result.TopDomains,
		SuggestedCareers: result.SuggestedCareers,
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"result":    resp,
		"analysis":  llmAnalysis,
		"copyright": "Thang đo ILO © ILO Vietnam 2020 – sử dụng cho mục đích hướng nghiệp, trích dẫn có ghi nguồn.",
	})
}

// @Summary Get ILO test questions
// @Description Get all questions for the ILO test
// @Tags ilo
// @Produce json
// @Success 200 {object} GetIloTestResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ilo/test [get]
func (h *Handler) HandleGetIloTest(c *fiber.Ctx) error {
	// Call the client to get ILO test questions
	test, err := h.IloClient.GetIloTest(c.Context())
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to get ILO test: "+err.Error())
	}

	// Add copyright information
	response := fiber.Map{
		"questions": test.Questions,
		"domains":   test.Domains,
		"levels":    test.Levels,
		"copyright": "Thang đo ILO © ILO Vietnam 2020 – sử dụng cho mục đích hướng nghiệp, trích dẫn có ghi nguồn.",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// @Summary Get all ILO test results for a user
// @Description Get all ILO test results for the authenticated user
// @Tags ilo
// @Produce json
// @Success 200 {array} IloTestResultResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ilo/results [get]
func (h *Handler) HandleGetIloResults(c *fiber.Ctx) error {
	token := utils.ExtractTokenFromHeader(c)
	if token == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Missing or invalid Authorization header")
	}

	// Validate token and get user ID
	user, err := h.authClient.ValidateToken(c.Context(), token)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid token: "+err.Error())
	}

	// Get all results for this user
	results, err := h.IloClient.GetIloTestResults(c.Context(), user.ID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to get ILO test results: "+err.Error())
	}

	// Create response array
	var respResults []IloTestResultResponse
	for _, result := range results {
		respResults = append(respResults, IloTestResultResponse{
			ID:               result.ID,
			UserID:           result.UserID,
			ResultData:       result.ResultData,
			CreatedAt:        result.CreatedAt,
			Scores:           result.Scores,
			TopDomains:       result.TopDomains,
			SuggestedCareers: result.SuggestedCareers,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"results":   respResults,
		"copyright": "Thang đo ILO © ILO Vietnam 2020 – sử dụng cho mục đích hướng nghiệp, trích dẫn có ghi nguồn.",
	})
}

// @Summary Get a specific ILO test result by ID
// @Description Get a specific ILO test result by ID for the authenticated user
// @Tags ilo
// @Produce json
// @Param id path string true "Result ID"
// @Success 200 {object} IloTestResultResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ilo/result/{id} [get]
func (h *Handler) HandleGetIloResultById(c *fiber.Ctx) error {
	token := utils.ExtractTokenFromHeader(c)
	if token == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Missing or invalid Authorization header")
	}

	// Validate token and get user ID
	user, err := h.authClient.ValidateToken(c.Context(), token)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid token: "+err.Error())
	}

	// Get result ID from path params
	resultID := c.Params("id")
	if resultID == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Missing result ID")
	}

	// Use the client method directly to retrieve the result by ID
	result, err := h.IloClient.GetIloTestResultById(c.Context(), resultID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to get ILO test result: "+err.Error())
	}

	// Verify that this result belongs to the authenticated user
	if result.UserID != user.ID {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "You don't have permission to access this result")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"result":    result,
		"copyright": "Thang đo ILO © ILO Vietnam 2020 – sử dụng cho mục đích hướng nghiệp, trích dẫn có ghi nguồn.",
	})
}
