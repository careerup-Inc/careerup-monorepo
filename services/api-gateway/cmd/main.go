package main

import (
	"log"
	"net/http"

	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/handler"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize clients
	authClient, err := client.NewAuthClient("localhost:50051")
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}
	defer authClient.Close()

	chatClient, err := client.NewChatClient("localhost:50052")
	if err != nil {
		log.Fatalf("Failed to create chat client: %v", err)
	}
	defer chatClient.Close()

	// Create handler
	h := handler.NewHandler(authClient, chatClient)

	// Initialize router
	r := gin.Default()

	// Add middleware
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit())
	r.Use(middleware.Auth(authClient))

	// Public routes
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)

	// Protected routes
	auth := r.Group("/")
	auth.Use(middleware.RequireAuth())
	{
		auth.GET("/me", h.GetCurrentUser)
		auth.PUT("/me", h.UpdateUser)
		auth.GET("/ws", h.WebSocket)
	}

	// Start server
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
