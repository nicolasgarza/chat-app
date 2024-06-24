package main

import (
	"log"
	// "net"
	// "google.golang.org/grpc"
	// pb "chat_app/chat"
	"chat_app/config"
	"chat_app/internal/chat"
	"chat_app/internal/logger"
	"chat_app/internal/ratelimit"
	"time"

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

	// init rate limiter
	rateLimiter := ratelimit.NewRateLimiter(rate.Every(time.Second), 100)

	// init chatserver
	chatServer := chat.NewChatServer(rateLimiter)

	// initialize grpc server
	// register chatservice
	// start listening
}
