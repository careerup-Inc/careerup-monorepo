package handler

import (
	"log"

	"github.com/careerup-Inc/careerup-monorepo/services/llm-gateway/internal/service" // Import the service package
	"github.com/gofiber/fiber/v2"
)

// Define request/response structs matching chat-gateway's client expectations
type CompletionRequest struct {
	ConversationID string `json:"conversation_id"`
	Prompt         string `json:"prompt"`
}

type CompletionResponse struct {
	Completion string `json:"completion"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// HTTPHandler holds dependencies like the OpenAI service.
type HTTPHandler struct {
	openAIService *service.OpenAIService
}

// NewHTTPHandler creates a new handler instance.
func NewHTTPHandler(openAIService *service.OpenAIService) *HTTPHandler {
	return &HTTPHandler{
		openAIService: openAIService,
	}
}

// HandleCompletion handles requests to the /api/v1/completion endpoint.
func (h *HTTPHandler) HandleCompletion(c *fiber.Ctx) error {
	var req CompletionRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body: " + err.Error()})
	}

	if req.Prompt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Prompt cannot be empty"})
	}

	log.Printf("Received completion request: conv_id=%s, prompt='%s'", req.ConversationID, req.Prompt)

	// Call the OpenAI service using GetCompletion
	completion, err := h.openAIService.GetCompletion(c.Context(), req.Prompt)
	if err != nil {
		log.Printf("Error getting completion from OpenAI service: %v", err)
		// Don't expose internal error details directly unless needed
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to get completion"})
	}

	log.Printf("Sending completion response: %s", completion)

	// Return the successful response
	return c.Status(fiber.StatusOK).JSON(CompletionResponse{
		Completion: completion,
	})
}
