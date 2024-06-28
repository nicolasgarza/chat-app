package storage

import (
	pb "chat_app/pb"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient(addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	// Test the connection
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func SaveMessage(client *redis.Client, message *pb.ChatMessage) error {
	ctx := context.Background()

	key := fmt.Sprintf("chat:message:%d", message.Timestamp)

	err := client.HSet(ctx, key,
		"user", message.User,
		"message", message.Message,
		"timestamp", message.Timestamp,
		"is_read", false,
		"channel", "general",
	).Err()
	if err != nil {
		return err
	}

	client.Expire(ctx, key, 24*time.Hour)

	return nil
}

func GetMessages(client *redis.Client) ([]*pb.ChatMessage, error) {
	ctx := context.Background()

	var cursor uint64
	var messages []*pb.ChatMessage
	pattern := "chat:message:*"

	for {
		var keys []string
		var err error

		keys, cursor, err := client.Scan(ctx, cursor, pattern, 10).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			jsonData, err := client.Get(ctx, key).Bytes()
			if err != nil {
				return nil, err
			}

			var msg pb.ChatMessage
			err = json.Unmarshal(jsonData, &msg)
			if err != nil {
				return nil, err
			}

			messages = append(messages, &msg)
		}

		// if cursor is 0, we've scanned all keys
		if cursor == 0 {
			break
		}
	}

	return messages, nil
}

func PublishMessage(client *redis.Client, channel string, message *pb.ChatMessage) error {
	ctx := context.Background()
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return client.Publish(ctx, channel, jsonMessage).Err()
}

func SubscribeToMessages(client *redis.Client, channel string) *redis.PubSub {
	return client.Subscribe(context.Background(), channel)
}

func SaveUser(client *redis.Client, username, hashedPassword string) error {
	ctx := context.Background()
	key := fmt.Sprintf("user:%s", username)

	return client.HSet(ctx, key,
		"username", username,
		"password", hashedPassword,
	).Err()
}

func GetUser(client *redis.Client, username string) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("token:%s", username)

	return client.HGet(ctx, key, "password").Result()
}

func SaveToken(client *redis.Client, username, token string, expiration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("token: %s", username)

	return client.HSet(ctx, key, expiration).Err()
}

func GetToken(client *redis.Client, username string) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("token: %s", username)

	return client.Get(ctx, key).Result()
}

func deleteToken(client *redis.Client, username string) error {
	ctx := context.Background()
	key := fmt.Sprintf("token: %s", username)

	return client.Del(ctx, key).Err()
}
