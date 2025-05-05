package handler

import (
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	"github.com/gofiber/fiber/v2"
)

func UpdateProfile(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header is required",
		})
	}
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	var req client.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	// Set token from header
	req.Token = token

	// Use gRPC client from context or global (update as needed)
	handlerIface := c.Locals("handler")
	var authClient client.AuthClientInterface
	if handlerIface != nil {
		h, ok := handlerIface.(interface {
			GetAuthClient() client.AuthClientInterface
		})
		if ok {
			authClient = h.GetAuthClient()
		}
	}
	if authClient == nil {
		// fallback: get from global or return error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Auth client not available",
		})
	}

	updatedUser, err := authClient.UpdateUser(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update profile: " + err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(updatedUser)
}
