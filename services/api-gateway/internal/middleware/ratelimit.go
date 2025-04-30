package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimitMiddleware creates a Redis-backed rate limiter for Fiber
func RateLimitMiddleware(client *redis.Client, requestsPerMinute int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := "rate_limit:" + ip

		// Get current count
		count, err := client.Get(c.Context(), key).Int()
		if err != nil && err != redis.Nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Check if limit exceeded
		if count >= requestsPerMinute {
			c.Set("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))
			c.Set("Retry-After", "60")

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Rate limit exceeded",
				"retry_after": 60,
			})
		}

		// Increment counter
		pipe := client.Pipeline()
		pipe.Incr(c.Context(), key)
		pipe.Expire(c.Context(), key, time.Minute)
		_, err = pipe.Exec(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Add rate limit headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(requestsPerMinute-count-1))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))

		return c.Next()
	}
}
