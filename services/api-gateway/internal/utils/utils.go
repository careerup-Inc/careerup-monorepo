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
