package chat

import (
	// "chat_app/internal/storage"
	"chat_app/internal/logger"
	pb "chat_app/pb"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidateMessage(msg *pb.ChatMessage) error {
	if msg.User == "" {
		return status.Error(codes.InvalidArgument, "User must not be empty")
	}
	if msg.Message == "" {
		return status.Error(codes.InvalidArgument, "Message must not be empty")
	}
	if msg.Timestamp == 0 {
		return status.Error(codes.InvalidArgument, "Message must not be empty")
	}

	return nil
}

func LogMessageReceived(msg *pb.ChatMessage) {
	logger.Log.Info("Received message",
		zap.String("user", msg.User),
		zap.Int64("timestamp", msg.Timestamp))
}

func LogStreamEnded(err error) {
	if err != nil {
		logger.Log.Error("Client disconnected from message stream with error", zap.Error(err))
	} else {
		logger.Log.Info("Client disconnected from message stream")
	}
}
