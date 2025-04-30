package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	chatpb "github.com/careerup-Inc/careerup-monorepo/proto/careerup/v1"
	"github.com/careerup-Inc/careerup-monorepo/services/chat-gateway/internal/server" // Import the server package
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection" // Import reflection
)

const (
	defaultPort = "8082" // Default gRPC port for chat-gateway
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	addr := fmt.Sprintf(":%s", port)

	// Create a TCP listener
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	log.Printf("gRPC server listening on %s", addr)

	// Create a new gRPC server instance
	grpcServer := grpc.NewServer()

	// Create an instance of our GrpcServer implementation
	chatServer := server.NewGrpcServer()

	// Register the ConversationService server
	chatpb.RegisterConversationServiceServer(grpcServer, chatServer)

	// Register reflection service on gRPC server (optional, useful for debugging/tools like grpcurl)
	reflection.Register(grpcServer)

	// Channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to start the server
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	log.Println("Chat Gateway started successfully.")

	// Wait for termination signal
	<-sigChan
	log.Println("Received termination signal. Shutting down gRPC server...")

	// Graceful shutdown
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped gracefully.")
}
