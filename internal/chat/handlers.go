package chat

import (
	// "chat_app/internal/storage"
	"chat_app/internal/logger"
	"chat_app/internal/storage"
	pb "chat_app/pb"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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
	if msg.Room == "" {
		return status.Error(codes.InvalidArgument, "Room must not be empty")
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

func HandleRegister(redisClient *redis.Client, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	// check if username already exists and hash password
	_, err := storage.GetUser(redisClient, req.Username)
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

	err = storage.SaveUser(redisClient, req.Username, string(hashedPassword))
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

	err = storage.SaveToken(redisClient, req.Username, token, 24*time.Hour)
	if err != nil {
		logger.Log.Error("Error saving token:", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to save token: %v", err)
	}
	return &pb.AuthResponse{Token: token}, nil
}

func HandleLogin(redisClient *redis.Client, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	// Retrieve hashed password for username from database
	hashedPassword, err := storage.GetUser(redisClient, req.Username)
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
	err = storage.SaveToken(redisClient, req.Username, token, 24*time.Hour)
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
