package chat

import (
	"chat_app/internal/logger"
	"chat_app/internal/ratelimit"
	"chat_app/internal/storage"
	pb "chat_app/pb"
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	// "chat_app/internal/storage"
)

type ChatServer struct {
	pb.UnimplementedChatServiceServer
	rateLimiter *ratelimit.RateLimiter
	redisClient *redis.Client
}

func NewChatServer(rateLimiter *ratelimit.RateLimiter, redisClient *redis.Client) *ChatServer {
	return &ChatServer{
		rateLimiter: rateLimiter,
		redisClient: redisClient,
	}
}

func (s *ChatServer) SendMessage(ctx context.Context, msg *pb.ChatMessage) (*pb.Empty, error) {
	if !s.rateLimiter.Allow() {
		return nil, status.Errorf(codes.ResourceExhausted, "Rate limit exceeded")
	}

	if err := storage.SaveMessage(s.redisClient, msg); err != nil {
		logger.Log.Error("Failed to save message", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to save message")
	}

	if err := storage.PublishMessage(s.redisClient, "chat_messages", msg); err != nil {
		logger.Log.Error("Failed to publish message", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to publish message")
	}

	logger.Log.Info("Message sent", zap.String("user", msg.User), zap.String("message", msg.Message))
	return &pb.Empty{}, nil
}

func (s *ChatServer) StreamMessages(empty *pb.Empty, stream pb.ChatService_StreamMessagesServer) error {
	pubsub := storage.SubscribeToMessages(s.redisClient, "chat_messages")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		var chatMessage pb.ChatMessage
		if err := json.Unmarshal([]byte(msg.Payload), &chatMessage); err != nil {
			logger.Log.Error("Failed to unmarshal message", zap.Error(err))
			continue
		}

		if err := stream.Send(&chatMessage); err != nil {
			logger.Log.Error("Failed to send message to stream", zap.Error(err))
			return err
		}
	}

	return nil
}
