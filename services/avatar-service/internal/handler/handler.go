package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	// TODO: Add service dependencies
}

func NewHandler() *Handler {
	return &Handler{}
}

type GenerateAvatarRequest struct {
	Style    string   `json:"style" binding:"required"`
	Features []string `json:"features" binding:"required"`
}

func (h *Handler) GenerateAvatar(c *gin.Context) {
	var req GenerateAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement avatar generation using VRoid Studio API
	c.JSON(http.StatusOK, gin.H{"message": "Avatar generation started"})
}

func (h *Handler) GetAvatar(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar ID is required"})
		return
	}

	// TODO: Implement avatar retrieval
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *Handler) UpdateAvatar(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar ID is required"})
		return
	}

	// TODO: Implement avatar update
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *Handler) DeleteAvatar(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar ID is required"})
		return
	}

	// TODO: Implement avatar deletion
	c.JSON(http.StatusOK, gin.H{"message": "Avatar deleted"})
}
