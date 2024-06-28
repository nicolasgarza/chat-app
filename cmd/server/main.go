package main

import (
	"log"
	// "net"
	"chat_app/config"
	"chat_app/internal/chat"
	"chat_app/internal/logger"
	"chat_app/internal/ratelimit"
	"chat_app/internal/storage"
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

	rateLimiter := ratelimit.NewRateLimiter(
		rate.Limit(float64(time.Second)/float64(config.AppConfig.RateLimit.Rate)),
		config.AppConfig.RateLimit.Burst,
	)

	// init chatserver
	redisClient, err := storage.NewRedisClient("localhost:6379")
	if err != nil {
		logger.Log.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	logger.Log.Info("Successfully connected to Redis")
	chatServer := chat.NewChatServer(rateLimiter, redisClient)

	go startWebServer(redisClient)

	// initialize grpc server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Log.Fatal("Failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(chatServer.AuthInterceptor),
	)

	// register chatservice
	pb.RegisterChatServiceServer(grpcServer, chatServer)

	// start listening
	logger.Log.Info("Starting gRPC server", zap.String("Address", lis.Addr().String()))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Log.Fatal("Failed to serve", zap.Error(err))
	}
}
