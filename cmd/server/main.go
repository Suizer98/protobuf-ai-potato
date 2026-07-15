package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	chatv1 "github.com/Suizer98/protobuf-ai-potato/gen/go/chat/v1"
	"github.com/Suizer98/protobuf-ai-potato/internal/config"
	"github.com/Suizer98/protobuf-ai-potato/internal/llm"
	"github.com/Suizer98/protobuf-ai-potato/internal/server"
	"github.com/Suizer98/protobuf-ai-potato/internal/session"
)

func main() {
	cfg := config.Load()

	provider, err := newProvider(cfg)
	if err != nil {
		log.Fatalf("provider: %v", err)
	}

	listener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	chatv1.RegisterChatServiceServer(grpcServer, server.NewChatServer(provider, session.NewStore()))

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(chatv1.ChatService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	go func() {
		log.Printf("gRPC listening on %s (provider=%s)", cfg.GRPCAddr, cfg.LLMProvider)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("serve: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Printf("shutting down")
	grpcServer.GracefulStop()
}

func newProvider(cfg config.Config) (llm.Provider, error) {
	switch cfg.LLMProvider {
	case "mock":
		return llm.NewMockProvider(), nil
	case "openai":
		return llm.NewOpenAIProvider(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL, cfg.OpenAIModel), nil
	default:
		return nil, fmt.Errorf("unsupported LLM_PROVIDER %q (use mock or openai)", cfg.LLMProvider)
	}
}
