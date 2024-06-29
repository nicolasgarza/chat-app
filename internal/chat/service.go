package chat

import (
	"chat_app/internal/logger"
	"chat_app/internal/ratelimit"
	"chat_app/internal/storage"
	pb "chat_app/pb"
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

	if err := ValidateMessage(msg); err != nil {
		return nil, err
	}

	LogMessageReceived(msg)

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
	logger.Log.Info("New client connected to message stream")
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
			LogStreamEnded(err)
			return err
		}
	}

	LogStreamEnded(nil)
	return nil
}

func (s *ChatServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	return HandleRegister(s.redisClient, req)
}

func (s *ChatServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	return HandleLogin(s.redisClient, req)
}

func (s *ChatServer) AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logger.Log.Info("AuthInterceptor called for method", zap.String("method", info.FullMethod))
	if info.FullMethod == "/chat.ChatService/Login" || info.FullMethod == "/chat.ChatService/Register" {
		logger.Log.Info("Hit login or register point")
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Log.Error("No metadata provided")
		return nil, status.Errorf(codes.Unauthenticated, "No metadata provided")
	}

	token := md["authorization"]
	if len(token) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "No token provided")
	}

	claims := &jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token[0], claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("dogdogdog"), nil
	})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid token")
	}

	username := (*claims)["username"].(string)

	// verify token against redis
	storedToken, err := storage.GetToken(s.redisClient, username)
	if err != nil {
		if err == redis.Nil {
			// token doesn't exist in redis
			return nil, status.Errorf(codes.Unauthenticated, "Token not found or expired")
		}

		// some other error occured
		logger.Log.Error("Error retrieving token from redis: ", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Error verifying token")
	}

	if storedToken != token[0] {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid token")
	}

	newCtx := context.WithValue(ctx, "username", username)
	return handler(newCtx, req)
}
