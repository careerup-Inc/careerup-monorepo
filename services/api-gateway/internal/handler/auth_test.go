package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthClient struct {
	mock.Mock
}

func NewMockAuthClient() *MockAuthClient {
	return &MockAuthClient{}
}

func (m *MockAuthClient) Register(req *client.RegisterRequest) (*client.User, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.User), args.Error(1)
}

func (m *MockAuthClient) Login(req *client.LoginRequest) (*client.TokenResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.TokenResponse), args.Error(1)
}

func (m *MockAuthClient) RefreshToken(refreshToken string) (*client.TokenResponse, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.TokenResponse), args.Error(1)
}

func (m *MockAuthClient) ValidateToken(token string) (*client.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.User), args.Error(1)
}

func (m *MockAuthClient) GetCurrentUser(token string) (*client.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.User), args.Error(1)
}

func (m *MockAuthClient) UpdateUser(token string, req *client.UpdateUserRequest) (*client.User, error) {
	args := m.Called(token, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.User), args.Error(1)
}

func TestAuthHandler_Register(t *testing.T) {
	app := fiber.New()
	mockClient := NewMockAuthClient()
	handler := NewAuthHandler(mockClient)
	app.Post("/api/v1/auth/register", handler.Register)

	t.Run("successful registration", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":     "test@example.com",
			"password":  "password123",
			"firstName": "John",
			"lastName":  "Doe",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		expectedUser := &client.User{
			ID:        "123",
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			IsActive:  true,
		}

		mockClient.On("Register", mock.Anything).Return(expectedUser, nil)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "123", response["id"])
		assert.Equal(t, "test@example.com", response["email"])
	})

	t.Run("invalid request body", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email": "invalid-email",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		mockClient.On("Register", mock.Anything).Return(nil, fiber.NewError(fiber.StatusBadRequest, "Invalid request body"))

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	app := fiber.New()
	mockClient := NewMockAuthClient()
	handler := NewAuthHandler(mockClient)
	app.Post("/api/v1/auth/login", handler.Login)

	t.Run("successful login", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		expectedToken := &client.TokenResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			ExpiresIn:    3600,
		}

		mockClient.On("Login", mock.Anything).Return(expectedToken, nil)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "access_token", response["access_token"])
		assert.Equal(t, "refresh_token", response["refresh_token"])
	})

	t.Run("invalid credentials", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		mockClient.On("Login", mock.Anything).Return(nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials"))

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAuthHandler_ValidateToken(t *testing.T) {
	app := fiber.New()
	mockClient := NewMockAuthClient()
	handler := NewAuthHandler(mockClient)
	app.Get("/api/v1/auth/validate", handler.ValidateToken)

	t.Run("successful token validation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/validate", nil)
		req.Header.Set("Authorization", "Bearer valid_token")

		expectedUser := &client.User{
			ID:        "123",
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			IsActive:  true,
		}

		mockClient.On("ValidateToken", "valid_token").Return(expectedUser, nil)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "123", response["id"])
		assert.Equal(t, "test@example.com", response["email"])
	})

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/validate", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/validate", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")

		mockClient.On("ValidateToken", "invalid_token").Return(nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token"))

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
