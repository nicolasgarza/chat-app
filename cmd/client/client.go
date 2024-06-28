package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "chat_app/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewChatServiceClient(conn)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Enter password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// try to log in first
	ctx, err := login(client, username, password)
	if err != nil {
		// if login fails, try to register:
		ctx, err = register(client, username, password)
		if err != nil {
			log.Fatalf("Failed to authenticate: %v", err)
		}
	}

	// start goroutine to recieve messages
	go recieveMessages(client, ctx)

	// send messages from user input
	for {
		fmt.Print("Enter message (or 'quit' to exit): ")
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)

		if message == "quit" {
			return
		}

		sendMessage(client, ctx, username, message)
	}
}

func login(client pb.ChatServiceClient, username, password string) (context.Context, error) {
	authResp, err := client.Login(context.Background(), &pb.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to login: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", authResp.Token)
	return ctx, nil
}

func register(client pb.ChatServiceClient, username, password string) (context.Context, error) {
	authResp, err := client.Register(context.Background(), &pb.RegisterRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register: %v", err)
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", authResp.Token)
	return ctx, nil
}

func sendMessage(client pb.ChatServiceClient, ctx context.Context, username, message string) {
	msg := &pb.ChatMessage{
		User:      username,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	_, err := client.SendMessage(ctx, msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	} else {
		log.Printf("Message sent: %s", msg.Message)
	}
}

func recieveMessages(client pb.ChatServiceClient, ctx context.Context) {
	stream, err := client.StreamMessages(ctx, &pb.Empty{})
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
