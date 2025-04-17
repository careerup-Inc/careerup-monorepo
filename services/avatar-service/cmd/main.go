package main

import (
	"log"
	"net/http"

	"github.com/careerup-Inc/careerup-monorepo/services/avatar-service/internal/handler"
	"github.com/careerup-Inc/careerup-monorepo/services/avatar-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize router
	r := gin.Default()

	// Add middleware
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit())

	// Create handler
	h := handler.NewHandler()

	// Routes
	r.POST("/v1/avatar/generate", h.GenerateAvatar)
	r.GET("/v1/avatar/:id", h.GetAvatar)
	r.PUT("/v1/avatar/:id", h.UpdateAvatar)
	r.DELETE("/v1/avatar/:id", h.DeleteAvatar)

	// Start server
	if err := http.ListenAndServe(":8082", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
