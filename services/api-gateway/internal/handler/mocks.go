package handler

import (
	"context"

	chatpb "github.com/careerup-Inc/careerup-monorepo/proto/careerup/v1"
	"github.com/careerup-Inc/careerup-monorepo/services/api-gateway/internal/client"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

// --- Mock Auth Client ---

type MockAuthClient struct {
	mock.Mock
}

func NewMockAuthClient() *MockAuthClient {
	return &MockAuthClient{}
}

func (m *MockAuthClient) ValidateToken(ctx context.Context, token string) (*client.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.User), args.Error(1)
}

func (m *MockAuthClient) Register(ctx context.Context, req *client.RegisterRequest) (*client.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.User), args.Error(1)
}

func (m *MockAuthClient) Login(ctx context.Context, req *client.LoginRequest) (*client.TokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.TokenResponse), args.Error(1)
}

func (m *MockAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*client.TokenResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.TokenResponse), args.Error(1)
}

func (m *MockAuthClient) GetCurrentUser(ctx context.Context, token string) (*client.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.User), args.Error(1)
}

func (m *MockAuthClient) UpdateUser(ctx context.Context, req *client.UpdateUserRequest) (*client.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.User), args.Error(1)
}

// --- Mock Chat Client (Implementing ConversationService client interface) ---

// Define an interface for the ChatClient if you haven't already
// This makes mocking easier. Example:
type ChatClientInterface interface {
	// Ensure this matches the method signature in your actual client used by the handler
	Stream(ctx context.Context) (chatpb.ConversationService_StreamClient, error)
	// Add other methods if your ChatHandler uses them
}

type MockChatClient struct {
	mock.Mock
}

func NewMockChatClient() *MockChatClient {
	return &MockChatClient{}
}

// Mock the Stream method to return a mock stream client
// Ensure the return type matches the generated proto code for ConversationService_StreamClient
func (m *MockChatClient) Stream(ctx context.Context) (chatpb.ConversationService_StreamClient, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Return the mock stream client
	return args.Get(0).(chatpb.ConversationService_StreamClient), args.Error(1)
}

// --- Mock Stream Client (Implementing ConversationService_StreamClient) ---

// Mock implementation for the ConversationService_StreamClient interface
type MockStreamClient struct {
	mock.Mock
	// Add channels or other mechanisms if needed to simulate stream behavior
}

func NewMockStreamClient() *MockStreamClient {
	return &MockStreamClient{}
}

// Implement methods required by the ConversationService_StreamClient interface
func (m *MockStreamClient) Send(req *chatpb.StreamRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockStreamClient) Recv() (*chatpb.StreamResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chatpb.StreamResponse), args.Error(1)
}

// Implement other grpc.ClientStream methods (usually return nil/defaults for tests)
func (m *MockStreamClient) Header() (metadata.MD, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(metadata.MD), args.Error(1)
}

func (m *MockStreamClient) Trailer() metadata.MD {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(metadata.MD)
}

func (m *MockStreamClient) CloseSend() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStreamClient) Context() context.Context {
	args := m.Called()
	if args.Get(0) == nil {
		// Return a default context if not explicitly set in mock
		return context.Background()
	}
	return args.Get(0).(context.Context)
}

// SendMsg and RecvMsg are generally not needed to be mocked directly
// unless your code uses them explicitly over Send/Recv.
func (m *MockStreamClient) SendMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockStreamClient) RecvMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}
