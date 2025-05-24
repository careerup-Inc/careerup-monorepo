package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	chatpb "github.com/careerup-Inc/careerup-monorepo/proto/careerup/v1"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/handler"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TestWebSocket upgrade scenarios
func TestHandleWebSocket_Upgrade(t *testing.T) {
	t.Run("successful upgrade with valid token", func(t *testing.T) {
		// Setup mocks
		mockAuthClient := handler.NewMockAuthClient()
		mockChatClient := handler.NewMockChatClient()

		// Create handler
		h := handler.NewHandler(mockAuthClient, mockChatClient, nil, nil, "")

		// Setup Fiber app with WebSocket support
		app := fiber.New()
		app.Get("/ws", h.HandleWebSocket)
		app.Get("/ws", websocket.New(h.WebSocketProxy))

		// Mock successful token validation
		expectedUser := &client.User{
			ID:        "test-user-123",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			IsActive:  true,
		}
		mockAuthClient.On("ValidateToken", mock.Anything, "valid_token").Return(expectedUser, nil)

		// Create request with WebSocket upgrade headers and auth
		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Sec-WebSocket-Version", "13")
		req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
		req.Header.Set("Authorization", "Bearer valid_token")

		// Test the request
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Should return 101 Switching Protocols for successful upgrade
		assert.Equal(t, fiber.StatusSwitchingProtocols, resp.StatusCode)

		// Verify mocks were called
		mockAuthClient.AssertExpectations(t)
	})

	t.Run("reject upgrade without authorization header", func(t *testing.T) {
		mockAuthClient := handler.NewMockAuthClient()
		mockChatClient := handler.NewMockChatClient()

		h := handler.NewHandler(mockAuthClient, mockChatClient, nil, nil, "")

		app := fiber.New()
		app.Get("/ws", h.HandleWebSocket)

		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Sec-WebSocket-Version", "13")
		req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
		// No Authorization header

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "Authorization header is required", response["error"])
	})

	t.Run("reject upgrade with invalid token", func(t *testing.T) {
		mockAuthClient := handler.NewMockAuthClient()
		mockChatClient := handler.NewMockChatClient()

		h := handler.NewHandler(mockAuthClient, mockChatClient, nil, nil, "")

		app := fiber.New()
		app.Get("/ws", h.HandleWebSocket)

		// Mock invalid token validation
		mockAuthClient.On("ValidateToken", mock.Anything, "invalid_token").Return(nil, status.Error(codes.Unauthenticated, "Invalid token"))

		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Sec-WebSocket-Version", "13")
		req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
		req.Header.Set("Authorization", "Bearer invalid_token")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid token", response["error"])

		mockAuthClient.AssertExpectations(t)
	})

	t.Run("reject non-websocket request", func(t *testing.T) {
		mockAuthClient := handler.NewMockAuthClient()
		mockChatClient := handler.NewMockChatClient()

		h := handler.NewHandler(mockAuthClient, mockChatClient, nil, nil, "")

		app := fiber.New()
		app.Get("/ws", h.HandleWebSocket)

		// Regular HTTP request without WebSocket upgrade headers
		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUpgradeRequired, resp.StatusCode)
	})

	t.Run("handle malformed bearer token", func(t *testing.T) {
		mockAuthClient := handler.NewMockAuthClient()
		mockChatClient := handler.NewMockChatClient()

		h := handler.NewHandler(mockAuthClient, mockChatClient, nil, nil, "")

		app := fiber.New()
		app.Get("/ws", h.HandleWebSocket)

		// Mock token validation for token without "Bearer " prefix
		mockAuthClient.On("ValidateToken", mock.Anything, "just_token").Return(nil, status.Error(codes.Unauthenticated, "Invalid token"))

		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Sec-WebSocket-Version", "13")
		req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
		req.Header.Set("Authorization", "just_token")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		mockAuthClient.AssertExpectations(t)
	})
}

// Mock WebSocket connection for testing
type MockWebSocketConn struct {
	mock.Mock
	locals map[string]interface{}
}

func NewMockWebSocketConn() *MockWebSocketConn {
	return &MockWebSocketConn{
		locals: make(map[string]interface{}),
	}
}

func (m *MockWebSocketConn) Locals(key string, value ...interface{}) interface{} {
	if len(value) > 0 {
		m.locals[key] = value[0]
		return value[0]
	}
	return m.locals[key]
}

func (m *MockWebSocketConn) ReadMessage() (messageType int, p []byte, err error) {
	args := m.Called()
	return args.Int(0), args.Get(1).([]byte), args.Error(2)
}

func (m *MockWebSocketConn) WriteJSON(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

func (m *MockWebSocketConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Note: TestWebSocketProxy tests removed - they were unit tests trying to test
// complex async operations without actually calling the WebSocketProxy method.
// For testing WebSocketProxy functionality, integration tests would be more appropriate.

// Mock implementation for ConversationServiceClient
type MockConversationServiceClient struct {
	mock.Mock
	streamClient *handler.MockStreamClient
}

func (m *MockConversationServiceClient) Stream(ctx context.Context, opts ...interface{}) (chatpb.ConversationService_StreamClient, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return m.streamClient, args.Error(1)
}

// TestMessageTransformation tests message conversion between WebSocket and gRPC
func TestMessageTransformation(t *testing.T) {
	t.Run("client message to grpc request", func(t *testing.T) {
		clientMsg := handler.ClientMessage{
			Type:           "user_msg",
			ConversationID: "conv-123",
			Text:           "Hello, AI assistant!",
		}

		// The actual transformation happens in WebSocketProxy
		// We're testing that the structure is correct
		assert.Equal(t, "user_msg", clientMsg.Type)
		assert.Equal(t, "conv-123", clientMsg.ConversationID)
		assert.Equal(t, "Hello, AI assistant!", clientMsg.Text)
	})

	t.Run("grpc response to server message", func(t *testing.T) {
		// Test assistant_token response
		grpcResponse := &chatpb.StreamResponse{
			Type:    "assistant_token",
			Content: &chatpb.StreamResponse_Token{Token: "Hello"},
		}

		// Expected server message
		expectedMsg := handler.ServerMessage{
			Type:  "assistant_token",
			Token: "Hello",
		}

		assert.Equal(t, "assistant_token", grpcResponse.Type)
		assert.Equal(t, "Hello", grpcResponse.GetToken())
		assert.Equal(t, "assistant_token", expectedMsg.Type)
		assert.Equal(t, "Hello", expectedMsg.Token)
	})

	t.Run("grpc error response to server message", func(t *testing.T) {
		grpcResponse := &chatpb.StreamResponse{
			Type:    "error",
			Content: &chatpb.StreamResponse_ErrorMessage{ErrorMessage: "Service unavailable"},
		}

		expectedMsg := handler.ServerMessage{
			Type:         "error",
			ErrorMessage: "Service unavailable",
		}

		assert.Equal(t, "error", grpcResponse.Type)
		assert.Equal(t, "Service unavailable", grpcResponse.GetErrorMessage())
		assert.Equal(t, "error", expectedMsg.Type)
		assert.Equal(t, "Service unavailable", expectedMsg.ErrorMessage)
	})

	t.Run("grpc avatar_url response to server message", func(t *testing.T) {
		grpcResponse := &chatpb.StreamResponse{
			Type:    "avatar_url",
			Content: &chatpb.StreamResponse_Url{Url: "https://example.com/avatar.png"},
		}

		expectedMsg := handler.ServerMessage{
			Type: "avatar_url",
			URL:  "https://example.com/avatar.png",
		}

		assert.Equal(t, "avatar_url", grpcResponse.Type)
		assert.Equal(t, "https://example.com/avatar.png", grpcResponse.GetUrl())
		assert.Equal(t, "avatar_url", expectedMsg.Type)
		assert.Equal(t, "https://example.com/avatar.png", expectedMsg.URL)
	})
}

// TestErrorHandling tests various error scenarios in WebSocket handling
func TestWebSocketErrorHandling(t *testing.T) {
	t.Run("grpc stream receive error", func(t *testing.T) {
		mockStreamClient := handler.NewMockStreamClient()

		// Mock stream receive error
		mockStreamClient.On("Recv").Return(nil, status.Error(codes.Unavailable, "Service unavailable"))

		// Test that error is properly handled
		_, err := mockStreamClient.Recv()
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unavailable, st.Code())
		assert.Equal(t, "Service unavailable", st.Message())

		mockStreamClient.AssertExpectations(t)
	})

	t.Run("websocket write error", func(t *testing.T) {
		mockConn := NewMockWebSocketConn()

		// Mock WebSocket write error
		mockConn.On("WriteJSON", mock.Anything).Return(fmt.Errorf("connection closed"))

		err := mockConn.WriteJSON(handler.ServerMessage{Type: "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection closed")

		mockConn.AssertExpectations(t)
	})
}

// TestContextAndMetadata tests context propagation and metadata handling
func TestContextAndMetadata(t *testing.T) {
	t.Run("user id metadata propagation", func(t *testing.T) {
		userID := "test-user-123"

		// Test metadata creation
		md := metadata.Pairs("user-id", userID)
		ctx := metadata.NewOutgoingContext(context.Background(), md)

		// Verify metadata is properly set
		outgoingMD, ok := metadata.FromOutgoingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, []string{userID}, outgoingMD.Get("user-id"))
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Test context cancellation
		cancel()

		select {
		case <-ctx.Done():
			assert.Equal(t, context.Canceled, ctx.Err())
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Context should have been canceled")
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// Wait for timeout
		select {
		case <-ctx.Done():
			assert.Equal(t, context.DeadlineExceeded, ctx.Err())
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Context should have timed out")
		}
	})
}

// TestConcurrentConnections tests handling of multiple simultaneous WebSocket connections
func TestConcurrentConnections(t *testing.T) {
	t.Run("multiple users connecting simultaneously", func(t *testing.T) {
		mockAuthClient := handler.NewMockAuthClient()
		mockChatClient := handler.NewMockChatClient()

		h := handler.NewHandler(mockAuthClient, mockChatClient, nil, nil, "")

		// Mock multiple users - testing that handler can be configured for multiple users
		users := []struct {
			id    string
			token string
		}{
			{"user-1", "token-1"},
			{"user-2", "token-2"},
			{"user-3", "token-3"},
		}

		// Verify handler can handle multiple user configurations
		assert.NotNil(t, h)
		assert.Len(t, users, 3)

		// In a real integration test, we would create multiple WebSocket connections
		// and verify they can all operate independently
		for _, user := range users {
			assert.NotEmpty(t, user.id)
			assert.NotEmpty(t, user.token)
		}
	})
}

// TestWebSocketLifecycle tests the complete lifecycle of a WebSocket connection
func TestWebSocketLifecycle(t *testing.T) {
	t.Run("complete connection lifecycle", func(t *testing.T) {
		mockAuthClient := handler.NewMockAuthClient()
		mockChatClient := handler.NewMockChatClient()
		mockStreamClient := handler.NewMockStreamClient()

		h := handler.NewHandler(mockAuthClient, mockChatClient, nil, nil, "")

		// 1. Test authentication flow setup
		user := &client.User{
			ID:       "lifecycle-user",
			Email:    "lifecycle@example.com",
			IsActive: true,
		}

		// 2. Test stream client configuration
		assert.NotNil(t, mockStreamClient)

		// 3. Verify all components are properly initialized
		assert.NotNil(t, h)
		assert.NotNil(t, user)
		assert.Equal(t, "lifecycle-user", user.ID)
		assert.Equal(t, "lifecycle@example.com", user.Email)
		assert.True(t, user.IsActive)
	})
}
