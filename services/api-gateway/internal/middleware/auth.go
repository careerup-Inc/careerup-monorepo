package middleware

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/patrickmn/go-cache"
)

// Cache for validated tokens to reduce calls to auth-core
var tokenCache = cache.New(5*time.Minute, 10*time.Minute)
var authClientInstance *client.AuthClient

// Initialize the auth client
func init() {
	authServiceAddr := os.Getenv("AUTH_SERVICE_ADDR")
	if authServiceAddr == "" {
		authServiceAddr = "auth-core:8081" // Default value
	}
	authClientInstance = client.NewAuthClient("http://" + authServiceAddr)
}

func AuthMiddleware(authClient *client.AuthClient) fiber.Handler {
	// Get JWT secret from environment variables
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// if not found then panic
		panic("JWT_SECRET environment variable is not set")
	}

	secretBytes := []byte(jwtSecret)

	return func(c *fiber.Ctx) error {
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

		// Basic format validation before calling auth-core
		if !isValidTokenFormat(tokenString) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token format",
			})
		}

		// Parse and validate the token using the environment variable
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg is what you expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretBytes, nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token is invalid",
			})
		}

		user, err := authClientInstance.ValidateToken(tokenString)
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

func isValidTokenFormat(tokenString string) bool {
	parts := strings.Split(tokenString, ".")
	return len(parts) == 3
}
