package main

import (
	"fmt"
	// "log"
	// "net"
	// "google.golang.org/grpc"
	// pb "chat_app/chat"
	"chat_app/internal/chat"
	// "chat_app/config"
	"chat_app/internal/logger"
	"chat_app/internal/ratelimit"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	fmt.Println("Entry point")
	err := logger.InitLogger()
	if err != nil {
		// handle
	}
	defer logger.Sync()

	// init rate limiter
	rateLimiter := ratelimit.NewRateLimiter(rate.Every(time.Second), 100)

	// init chatserver
	chatServer := chat.NewChatServer(rateLimiter)

	// initialize grpc server
	// register chatservice
	// start listening
}
