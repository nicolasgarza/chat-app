package chat

import (
	// "chat_app/internal/storage"
	"chat_app/internal/logger"
	pb "chat_app/pb"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ChatServer) HandleSendMessage(msg *pb.ChatMessage) error {
	if !s.rateLimiter.Allow() {
		logger.Log.Warn("Rate limit exceeded for send message")
		return status.Error(codes.ResourceExhausted, "Rate Limit exceeded")
	}

	logger.Log.Info("Received message", zap.String("user", msg.User))
	// Handle logic for sending message
}

func HandleStreamMessages(s *ChatServer, stream pb.ChatService_StreamMessagesServer) error {
	// Handle streaming of messages
}
