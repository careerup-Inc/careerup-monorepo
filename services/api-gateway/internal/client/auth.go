package client

import (
	"context"
	"log"

	v1 "github.com/careerup-Inc/careerup-monorepo/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClient struct {
	client v1.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &AuthClient{
		client: v1.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}

func (c *AuthClient) Register(ctx context.Context, email, password, firstName, lastName string) (*v1.User, error) {
	resp, err := c.client.Register(ctx, &v1.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	})
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		return nil, err
	}
	return resp, nil
}

func (c *AuthClient) Login(ctx context.Context, email, password string) (*v1.LoginResponse, error) {
	resp, err := c.client.Login(ctx, &v1.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		log.Printf("Failed to login user: %v", err)
		return nil, err
	}
	return resp, nil
}

func (c *AuthClient) GetCurrentUser(ctx context.Context, userID string) (*v1.User, error) {
	resp, err := c.client.GetCurrentUser(ctx, &v1.GetCurrentUserRequest{})
	if err != nil {
		log.Printf("Failed to get current user: %v", err)
		return nil, err
	}
	return resp, nil
}

func (c *AuthClient) UpdateUser(ctx context.Context, userID string, req *v1.UpdateUserRequest) (*v1.User, error) {
	resp, err := c.client.UpdateUser(ctx, req)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		return nil, err
	}
	return resp, nil
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (*v1.User, error) {
	resp, err := c.client.ValidateToken(ctx, &v1.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		log.Printf("Failed to validate token: %v", err)
		return nil, err
	}
	return resp.User, nil
}
