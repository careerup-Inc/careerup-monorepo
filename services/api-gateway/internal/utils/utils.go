package utils

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func SendErrorResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error":     message,
		"status":    status,
		"timestamp": time.Now().Unix(),
	})
}

// ExtractTokenFromHeader extracts the Bearer token from the Authorization header
func ExtractTokenFromHeader(c *fiber.Ctx) string {
	header := c.Get("Authorization")
	if len(header) > 7 && header[:7] == "Bearer " {
		return header[7:]
	}
	return ""
}
