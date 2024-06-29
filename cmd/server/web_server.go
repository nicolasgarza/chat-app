package main

import (
	"chat_app/internal/storage"
	pb "chat_app/pb"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
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
	r := mux.NewRouter()

	r.HandleFunc("/", handleHome)
	r.HandleFunc("/room/", handleRoom)
	r.HandleFunc("/ws/{roomName}", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, redisClient)
	})

	log.Println("Starting web server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/home.html"))
	tmpl.Execute(w, nil)
}

func handleRoom(w http.ResponseWriter, r *http.Request) {
	roomName := r.URL.Query().Get("roomName")
	if roomName == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	tmpl := template.Must(template.ParseFiles("templates/room.html"))
	tmpl.Execute(w, map[string]string{"RoomName": roomName})
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	vars := mux.Vars(r)
	roomName := vars["roomName"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// Subscribe to Redis channel for the specific room
	pubsub := storage.SubscribeToMessages(redisClient, "chat_messages:"+roomName)
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
			"room":      chatMessage.Room,
		}

		if err := conn.WriteJSON(messageData); err != nil {
			log.Println("Error writing to WebSocket:", err)
			return
		}
	}
}
