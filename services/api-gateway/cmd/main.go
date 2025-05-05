package main

import (
	"log"
	"strconv"

	_ "github.com/careerup-Inc/careerup-monorepo/services/api-gateway/docs"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/config"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/handler"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/middleware"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
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

// @securityDefinitions.BearerAuth
// @type http
// @scheme bearer
// @bearerFormat JWT

func main() {
	log.Println("Starting API Gateway...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	cfg, err := config.LoadConfig("./configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	})

	// Middleware
	app.Use(cors.New())
	app.Use(logger.New())

	// Initialize Redis for rate limiting
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RateLimit.RedisAddr,
	})
	defer redisClient.Close()

	// Add rate limiting if enabled
	if cfg.RateLimit.Enabled {
		app.Use(middleware.RateLimitMiddleware(redisClient, cfg.RateLimit.RequestsPerMinute))
	}

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Serve Swagger JSON
	app.Static("/swagger/doc.json", "./docs/swagger.json")

	// Initialize clients
	authClient, err := client.NewAuthClient(cfg.Auth.ServiceAddr)
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}
	defer authClient.Close()

	chatClient, err := client.NewChatClient(cfg.Chat.ServiceAddr)
	if err != nil {
		log.Fatalf("Failed to create chat client: %v", err)
	}
	defer chatClient.Close()

	// Initialize middlewares with auth client
	authMiddleware := middleware.AuthMiddleware(authClient)

	// Initialize handlers
	mainHandler := handler.NewHandler(authClient, chatClient)

	// Protected routes (Apply middleware before defining groups/routes)
	protectedUser := app.Group("/api/v1/user", authMiddleware)       // Apply middleware to group
	protectedProfile := app.Group("/api/v1/profile", authMiddleware) // Apply middleware to group

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
			auth.Post("/register", mainHandler.HandleRegister)
			auth.Post("/login", mainHandler.HandleLogin)
			auth.Post("/refresh", mainHandler.HandleRefreshToken)
			auth.Get("/validate", mainHandler.HandleValidateToken)
		}

		// User routes (Protected via group middleware)
		// These routes are already prefixed with /api/v1/user by the group
		protectedUser.Get("/me", mainHandler.HandleGetProfile)

		// Profile routes (Protected via group middleware)
		// These routes are already prefixed with /api/v1/profile by the group
		protectedProfile.Put("", mainHandler.HandleUpdateProfile) // Use PUT on the group base path

		// Chat routes with WebSocket support (Unprotected initial upgrade, auth done inside handler)
		api.Get("/ws", mainHandler.HandleWebSocket)
		api.Get("/ws", websocket.New(mainHandler.WebSocketProxy))
	}

	// Start server
	port := cfg.Server.Port
	if port == 0 {
		port = 8080 // Default port
	}

	log.Printf("Server starting on port %d", port)
	if err := app.Listen(":" + strconv.Itoa(port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
