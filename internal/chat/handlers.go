package chat

import (
	// "chat_app/internal/storage"
	pb "chat_app/pb"
)

func HandleSendMessage(s *ChatServer, msg *pb.ChatMessage) error {
	// Handle logic for sending message
}

func HandleStreamMessages(s *ChatServer, stream pb.ChatService_StreamMessagesServer) error {
	// Handle streaming of messages
}
