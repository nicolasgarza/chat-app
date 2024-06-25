package chat

import (
	"chat_app/internal/logger"
	"chat_app/internal/ratelimit"
	pb "chat_app/pb"
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	// "chat_app/internal/storage"
)

type ChatServer struct {
	pb.UnimplementedChatServiceServer
	rateLimiter *ratelimit.RateLimiter
	// more
}

func NewChatServer(rateLimiter *ratelimit.RateLimiter) *ChatServer {
	return &ChatServer{
		rateLimiter: rateLimiter,
	}
}

func (s *ChatServer) SendMessage(ctx context.Context, msg *pb.ChatMessage) (*pb.Empty, error) {
	if !s.rateLimiter.Allow() {
		return nil, status.Errorf(codes.ResourceExhausted, "Rate limit exceeded")
	}

	// TODO: Add logic to process and store the SendMessage
	// EX: save to redis

	logger.Log.Info("Message sent", zap.String("user", msg.User), zap.String("message", msg.Message))
	return &pb.Empty{}, nil
}

func (s *ChatServer) StreamMessages(empty *pb.Empty, stream pb.ChatService_StreamMessagesServer) error {
	// TODO: Implement logic to stream messages
	// Set up a goroutime to listen for new messages and send them over the stream

	for {
		// EX: send a message every 5 seconds
		time.Sleep(5 * time.Second)
		msg := &pb.ChatMessage{
			User:      "System",
			Message:   "This is a test message",
			Timestamp: time.Now().Unix(),
		}
		if err := stream.Send(msg); err != nil {
			return err
		}

	}
}
