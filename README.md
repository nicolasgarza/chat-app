# chat-app

This is a chat app built in Go. 

Features:

- Multiple chatrooms

- Multiple users

- Web ui to view chatrooms

## Usage

1. Git clone the repository

2. cd to the directory

3. Run the following commands in seperate terminal shells:

- redis-server (make sure you have redis installed)

- go run cmd/server/*.go

- go run cmd/client/client.go

client.go offers a CLI to send messages to chatrooms. You can open another terminal shell and run client.go again to add another user to the chatroom.

Open localhost:8080 to view the messages in the chatrooms.
