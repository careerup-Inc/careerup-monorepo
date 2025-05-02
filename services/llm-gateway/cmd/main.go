package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pbllm "github.com/careerup-Inc/careerup-monorepo/proto/llm/v1"                    // Corrected import path
	"github.com/careerup-Inc/careerup-monorepo/services/llm-gateway/internal/service" // Corrected import path
)

func main() {
	// Configuration (consider using a config file/library)
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50053" // Default gRPC port for llm-gateway
	}
	grpcAddr := fmt.Sprintf(":%s", grpcPort)

	log.Printf("Starting LLM Gateway gRPC server on %s", grpcAddr)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer(
	// Add interceptors if needed (logging, metrics, auth)
	// grpc.UnaryInterceptor(...),
	// grpc.StreamInterceptor(...),
	)

	// Create and register LLM service implementation
	llmSvc, err := service.NewLLMService()
	if err != nil {
		log.Fatalf("Failed to create LLM service: %v", err)
	}
	pbllm.RegisterLLMServiceServer(grpcServer, llmSvc)
	log.Println("LLMService registered")

	// Optional: Register reflection service on gRPC server.
	// Makes it easy for tools like grpcurl to interact with the server.
	reflection.Register(grpcServer)
	log.Println("gRPC reflection registered")

	// Start gRPC server in a goroutine
	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gRPC server...")

	grpcServer.GracefulStop()
	log.Println("gRPC server stopped.")
}
