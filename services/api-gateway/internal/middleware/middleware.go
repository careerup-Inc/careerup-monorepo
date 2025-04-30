package middleware

import (
	"time"

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
