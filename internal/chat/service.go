package chat

import (
	pb "chat_app/pb"
	"context"
	// "chat_app/internal/storage"
	// "chat_app/internal/auth"
)

type ChatServer struct {
	pb.UnimplementedChatServiceServer
	// more
}

func (s *ChatServer) SendMessage(ctx context.Context, msg *pb.ChatMessage) (*pb.Empty, error) {
	// impl sending a message
	// return &pb.Empty{}, nil if successful
}

func (s *ChatServer) StreamMessages(empty *pb.Empty, stream pb.ChatService_StreamMessagesServer) error {
	// Implement streaming messages
}
