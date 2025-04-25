package main

import (
	"log"
	"os"

	"github.com/arsmn/fiber-swagger/v2"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/docs"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/handler"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
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

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	app := fiber.New()

	// Middleware
	app.Use(cors.New())
	app.Use(logger.New())

	// Swagger
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/swagger/doc.json",
		DeepLinking: true,
		Title:       "CareerUP API",
	}))

	// Serve Swagger JSON
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		return c.JSON(docs.SwaggerInfo)
	})

	// Initialize auth client
	authClient := client.NewAuthClient("http://auth-core:8081")

	// Initialize middlewares with auth client
	authMiddleware := middleware.Auth(authClient)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authClient)

	// Use auth middleware for protected routes
	app.Use("/api/protected/*", authMiddleware)

	// Routes
	api := app.Group("/api/v1")
	{
		// Health check
		api.Get("/health", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"status": "ok",
			})
		})

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.Post("/register", authHandler.Register)
			auth.Post("/login", authHandler.Login)
			auth.Post("/refresh", authHandler.RefreshToken)
			auth.Get("/validate", authHandler.ValidateToken)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
