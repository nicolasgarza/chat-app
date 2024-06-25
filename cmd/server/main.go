package main

import (
	"log"
	// "net"
	"chat_app/config"
	"chat_app/internal/chat"
	"chat_app/internal/logger"
	"chat_app/internal/ratelimit"
	pb "chat_app/pb"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"golang.org/x/time/rate"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	err := logger.InitLogger()
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	defer logger.Log.Sync()

	logger.Log.Info("Application started")

	rateLimiter := ratelimit.NewRateLimiter(rate.Every(time.Second), 100)

	// init chatserver
	chatServer := chat.NewChatServer(rateLimiter)

	// initialize grpc server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Log.Fatal("Failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	// register chatservice
	pb.RegisterChatServiceServer(grpcServer, chatServer)

	// start listening
	logger.Log.Info("Staarting gRPC server", zap.String("Address", lis.Addr().String()))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Log.Fatal("Failed to serve", zap.Error(err))
	}
}
