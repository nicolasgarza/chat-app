package chat

import (
	"chat_app/internal/ratelimit"
	pb "chat_app/pb"
	"context"
	// "chat_app/internal/storage"
)

type ChatServer struct {
	pb.UnimplementedChatServiceServer
	rateLimiter *ratelimit.RateLimiter
	// more
}

func NewChatServer(rateLimiter *ratelimit.RateLimiter) *ChatServer {
	// init new chat server
}

func (s *ChatServer) SendMessage(ctx context.Context, msg *pb.ChatMessage) (*pb.Empty, error) {
	// impl sending a message
	// return &pb.Empty{}, nil if successful
}

func (s *ChatServer) StreamMessages(empty *pb.Empty, stream pb.ChatService_StreamMessagesServer) error {
	// Implement streaming messages
}
