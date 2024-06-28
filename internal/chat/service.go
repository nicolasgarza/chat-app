package chat

import (
	"chat_app/internal/logger"
	"chat_app/internal/ratelimit"
	"chat_app/internal/storage"
	pb "chat_app/pb"
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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
	// check if username already exists and hash password
	_, err := storage.GetUser(s.redisClient, req.Username)
	if err == nil {
		// user already exists
		return nil, status.Errorf(codes.AlreadyExists, "Username already exists")
	} else if err != redis.Nil {
		// unexpected error
		logger.Log.Error("Error checking user existence:", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Error checking user existence")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("Error hashing password: ", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to hash password")
	}

	err = storage.SaveUser(s.redisClient, req.Username, string(hashedPassword))
	if err != nil {
		logger.Log.Error("Error saving user:", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to save user: %v", err)
	}

	// store username and hashed password in db
	// Generate JWT token
	token, err := generateToken(req.Username)
	if err != nil {
		logger.Log.Error("Error generating token:", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to generate token")
	}

	err = storage.SaveToken(s.redisClient, req.Username, token, 24*time.Hour)
	if err != nil {
		logger.Log.Error("Error saving token:", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to save token: %v", err)
	}
	return &pb.AuthResponse{Token: token}, nil
}

func (s *ChatServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	// Retrieve hashed password for username from database
	hashedPassword, err := storage.GetUser(s.redisClient, req.Username)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "User not found: %v", err)
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid credentials")
	}

	// Generate JWT token
	token, err := generateToken(req.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate token")
	}

	// save token to redis
	err = storage.SaveToken(s.redisClient, req.Username, token, 24*time.Hour)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to save token: %v", err)
	}

	return &pb.AuthResponse{Token: token}, nil
}

func generateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString([]byte("dogdogdog"))
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
