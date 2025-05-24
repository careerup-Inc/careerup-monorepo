package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pbllm "github.com/careerup-Inc/careerup-monorepo/proto/llm/v1"
	"github.com/careerup-Inc/careerup-monorepo/services/llm-gateway/internal/handler"
	"github.com/careerup-Inc/careerup-monorepo/services/llm-gateway/internal/service"
)

func main() {
	// Configuration (consider using a config file/library)
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50053" // Default gRPC port for llm-gateway
	}
	grpcAddr := fmt.Sprintf(":%s", grpcPort)

	// HTTP admin port
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8090" // Default HTTP port for admin endpoints
	}
	httpAddr := fmt.Sprintf(":%s", httpPort)

	log.Printf("Starting LLM Gateway gRPC server on %s", grpcAddr)
	log.Printf("Starting LLM Gateway HTTP admin server on %s", httpAddr)

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

	// Create HTTP admin server
	adminHandler := handler.NewAdminHandler(llmSvc)
	mux := http.NewServeMux()
	adminHandler.RegisterRoutes(mux)

	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	// Start gRPC server in a goroutine
	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP admin server in a goroutine
	go func() {
		log.Printf("HTTP admin server listening at %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	grpcServer.GracefulStop()
	httpServer.Close()
	log.Println("Servers stopped.")
}
