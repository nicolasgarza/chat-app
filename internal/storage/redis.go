package storage

import (
	pb "chat_app/pb"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient(addr string) (*redis.Client, error) {
	// init and return redis client
}

func SaveMessage(client *redis.Client, message *pb.ChatMessage) error {
	// Save a message to redis
}

func GetMessages(client *redis.Client) ([]*pb.ChatMessage, error) {
	// Retrieve messages from Redis
}
