package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pbChat "github.com/careerup-Inc/careerup-monorepo/proto/careerup/v1"
	"github.com/careerup-Inc/careerup-monorepo/services/chat-gateway/internal/client"
	"github.com/careerup-Inc/careerup-monorepo/services/chat-gateway/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Configuration (consider using a config file/library like Viper or envconfig)
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "8082" // Default gRPC port for chat-gateway
	}
	grpcAddr := fmt.Sprintf(":%s", grpcPort)

	llmServiceAddr := os.Getenv("LLM_SERVICE_ADDR")
	if llmServiceAddr == "" {
		llmServiceAddr = "llm-gateway-py:50054" // Default address for llm-gateway (service name in Docker)
	}

	log.Printf("Starting Chat Gateway gRPC server on %s", grpcAddr)
	log.Printf("Connecting to LLM Service at %s", llmServiceAddr)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer(
	// Add interceptors if needed (logging, metrics, auth propagation)
	// grpc.StreamInterceptor(...),
	)

	// Create LLM gRPC client
	llmClient, err := client.NewLLMClient(llmServiceAddr)
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}
	defer llmClient.Close() // Ensure connection is closed on shutdown

	// Create ILO gRPC client connection (reuse llmServiceAddr for now, or use env var ILO_SERVICE_ADDR)
	iloServiceAddr := os.Getenv("ILO_SERVICE_ADDR")
	if iloServiceAddr == "" {
		iloServiceAddr = "auth-core:9091"
	}
	connIlo, err := grpc.NewClient(iloServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to ILO service: %v", err)
	}
	defer connIlo.Close()

	iloClient := client.NewIloClient(connIlo)

	// Create and register Chat service implementation
	chatSvc := server.NewChatServer(llmClient, iloClient)
	// Use the correct registration function based on the generated code
	pbChat.RegisterConversationServiceServer(grpcServer, chatSvc)
	log.Println("ConversationService registered")

	// Optional: Register reflection service
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
