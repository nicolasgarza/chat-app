package main

import (
	"context"
	"log"
	"time"

	pb "chat_app/pb"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewChatServiceClient(conn)

	// Start a new goroutine to recieve messages
	go recieveMessages(client)

	// Send a message every 5 seconds
	for {
		sendMessage(client)
		time.Sleep(5 * time.Second)
	}
}

func sendMessage(client pb.ChatServiceClient) {
	msg := &pb.ChatMessage{
		User:      "Test User",
		Message:   "Hello World",
		Timestamp: time.Now().Unix(),
	}

	_, err := client.SendMessage(context.Background(), msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	} else {
		log.Printf("Message sent: %s", msg.Message)
	}
}

func recieveMessages(client pb.ChatServiceClient) {
	stream, err := client.StreamMessages(context.Background(), &pb.Empty{})
	if err != nil {
		log.Fatalf("Error opening stream: %v", err)
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Error recieving messages: %v", err)
			return
		}
		log.Printf("Received: %s from %s", msg.Message, msg.User)
	}
}
