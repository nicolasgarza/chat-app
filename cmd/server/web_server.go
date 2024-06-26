package main

import (
	"chat_app/internal/storage"
	pb "chat_app/pb"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for this example
	},
}

func startWebServer(redisClient *redis.Client) {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, redisClient)
	})

	log.Println("Starting web server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// Subscribe to Redis channel
	pubsub := storage.SubscribeToMessages(redisClient, "chat_messages")
	defer pubsub.Close()

	channel := pubsub.Channel()

	for msg := range channel {
		var chatMessage pb.ChatMessage
		if err := json.Unmarshal([]byte(msg.Payload), &chatMessage); err != nil {
			log.Println("Error unmarshalling message:", err)
			continue
		}

		// Create a map of the message data to avoid copying the mutex
		messageData := map[string]interface{}{
			"user":      chatMessage.User,
			"message":   chatMessage.Message,
			"timestamp": chatMessage.Timestamp,
		}

		if err := conn.WriteJSON(messageData); err != nil {
			log.Println("Error writing to WebSocket:", err)
			return
		}
	}
}
