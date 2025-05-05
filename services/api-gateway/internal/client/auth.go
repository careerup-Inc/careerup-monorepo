package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/careerup-Inc/careerup-monorepo/proto/careerup/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthClientInterface interface {
	Register(ctx context.Context, req *RegisterRequest) (*User, error)
	Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	ValidateToken(ctx context.Context, token string) (*User, error)
	// GetCurrentUser(ctx context.Context, token string) (*User, error) // Don't need this for the implement as we already have the user in the context so no need to implement this
	UpdateUser(ctx context.Context, req *UpdateUserRequest) (*User, error)
}

type AuthClient struct {
	conn   *grpc.ClientConn
	client pb.AuthServiceClient
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %s %v", addr, err)
	} else {
		log.Printf("Connected to auth service at %s", addr)
	}

	return &AuthClient{
		conn:   conn,
		client: pb.NewAuthServiceClient(conn),
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type User struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	IsActive  bool     `json:"isActive"`
	Hometown  string   `json:"hometown"`
	Interests []string `json:"interests"`
}

type UpdateUserRequest struct {
	Token     string   `json:"-"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Hometown  string   `json:"hometown"`
	Interests []string `json:"interests"`
}

func (c *AuthClient) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	resp, err := c.client.Register(ctx, &pb.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		return nil, err
	}

	return convertProtoUser(resp.User), nil
}

func (c *AuthClient) Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error) {
	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	resp, err := c.client.Login(ctx, &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpireIn,
	}, nil
}

func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	resp, err := c.client.RefreshToken(ctx, &pb.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpireIn,
	}, nil
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return nil, err
	}

	return convertProtoUser(resp.User), nil
}

func (c *AuthClient) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*User, error) {
	if req.FirstName == "" && req.LastName == "" && req.Hometown == "" && len(req.Interests) == 0 {
		return nil, fmt.Errorf("at least one field is required to update")
	}

	resp, err := c.client.UpdateUser(ctx, &pb.UpdateUserRequest{
		Token:     req.Token,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Hometown:  req.Hometown,
		Interests: req.Interests,
	})
	if err != nil {
		return nil, err
	}

	return convertProtoUser(resp.User), nil
}

func convertProtoUser(protoUser *pb.User) *User {
	return &User{
		ID:        protoUser.Id,
		Email:     protoUser.Email,
		FirstName: protoUser.FirstName,
		LastName:  protoUser.LastName,
		Hometown:  protoUser.Hometown,
		Interests: protoUser.Interests,
		IsActive:  protoUser.IsActive,
	}
}
