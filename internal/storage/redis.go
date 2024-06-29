package storage

import (
	"chat_app/internal/logger"
	pb "chat_app/pb"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
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
	key := fmt.Sprintf("chat:message:%s:%d", message.Room, message.Timestamp)

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = client.ZAdd(ctx, key, &redis.Z{
		Score:  float64(message.Timestamp),
		Member: jsonMessage,
	}).Err()
	if err != nil {
		return err
	}

	// trim sorted set to keep only last 100 messages
	go func() {
		err := client.ZRemRangeByRank(ctx, key, 0, -101).Err()
		if err != nil {
			logger.Log.Error("Failed to trim message list", zap.Error(err), zap.String("room", message.Room))
		}
	}()

	return nil
}

func GetMessages(client *redis.Client, room string) ([]*pb.ChatMessage, error) {
	ctx := context.Background()

	var cursor uint64
	var messages []*pb.ChatMessage
	pattern := fmt.Sprintf("chat:message:%s:*", room)

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

func GetLastNMessages(client *redis.Client, room string, n int) ([]*pb.ChatMessage, error) {
	ctx := context.Background()
	key := fmt.Sprintf("chat:messages:%s", room)

	results, err := client.ZRevRange(ctx, key, 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}

	var messages []*pb.ChatMessage
	for _, result := range results {
		var msg pb.ChatMessage
		err = json.Unmarshal([]byte(result), &msg)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	// reverse order
	for i := len(messages)/2 - 1; i >= 0; i-- {
		opp := len(messages) - 1 - i
		messages[i], messages[opp] = messages[opp], messages[i]
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
	key := fmt.Sprintf("user:%s", username)

	return client.HGet(ctx, key, "password").Result()
}

func SaveToken(client *redis.Client, username, token string, expiration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("token:%s", username)

	err := client.Set(ctx, key, token, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to save token: %v", err)
	}

	return nil
}

func GetToken(client *redis.Client, username string) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("token:%s", username)

	return client.Get(ctx, key).Result()
}

func deleteToken(client *redis.Client, username string) error {
	ctx := context.Background()
	key := fmt.Sprintf("token: %s", username)

	return client.Del(ctx, key).Err()
}
