package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
)

// Cache for validated tokens to reduce calls to auth-core
var tokenCache = cache.New(5*time.Minute, 10*time.Minute)

func AuthMiddleware(authClient *client.AuthClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create context with timeout for the gRPC call
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}

		// Check if the Authorization header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Check if token is in cache
		if cachedUser, found := tokenCache.Get(tokenString); found {
			// Set user in context
			c.Locals("user", cachedUser)
			return c.Next()
		}

		// Use gRPC client to validate token against auth service
		user, err := authClient.ValidateToken(ctx, tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Store in cache
		tokenCache.Set(tokenString, user, cache.DefaultExpiration)

		// Add user information to the context
		c.Locals("user", user)

		return c.Next()
	}
}
