package middleware

import (
	"context"
	"time"

	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
)

// CORS middleware
func CORS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}

		return c.Next()
	}
}

// RateLimitInMemory middleware
func RateLimitInMemory() fiber.Handler {
	limiter := rate.NewLimiter(rate.Every(time.Second), 10) // 10 requests per second

	return func(c *fiber.Ctx) error {
		if !limiter.Allow() {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded",
			})
		}
		return c.Next()
	}
}

// Auth middleware
func Auth(authClient *client.AuthClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create context with timeout for the gRPC call
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		token := c.Get("Authorization")
		if token == "" {
			return c.Next()
		}

		user, err := authClient.ValidateToken(ctx, token)
		if err != nil {
			return c.Next()
		}

		c.Locals("user_id", user.ID)
		return c.Next()
	}
}

// RequireAuth middleware
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}
		return c.Next()
	}
}
