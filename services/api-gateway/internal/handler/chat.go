package handler

import (
	"bufio"
	"encoding/json"
	"log"

	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	pb "github.com/careerup-Inc/careerup-monorepo/services/api-gateway/proto/careerup/chat"
	"github.com/gofiber/fiber/v2"
)

type ChatHandler struct {
	chatClient *client.ChatClient
}

func NewChatHandler(chatClient *client.ChatClient) *ChatHandler {
	return &ChatHandler{
		chatClient: chatClient,
	}
}

func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	var message pb.ChatMessage
	if err := c.BodyParser(&message); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	resp, err := h.chatClient.SendMessage(c.Context(), &message)
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send message",
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *ChatHandler) StreamMessages(c *fiber.Ctx) error {
	stream, err := h.chatClient.StreamMessages(c.Context(), c.Params("user_id"))
	if err != nil {
		log.Printf("Error creating message stream: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create message stream",
		})
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for {
			message, err := stream.Recv()
			if err != nil {
				log.Printf("Error receiving message: %v", err)
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			if _, err := w.WriteString("data: " + string(data) + "\n\n"); err != nil {
				log.Printf("Error writing to stream: %v", err)
				return
			}

			if err := w.Flush(); err != nil {
				log.Printf("Error flushing stream: %v", err)
				return
			}
		}
	})

	return nil
}
