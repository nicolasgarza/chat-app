# chat-app

An HTML and WebSockets chat application built in Go. 

Features:

- Multiple chatrooms

- Multiple users

- Web UI to view chatrooms

## Installation
**Redis** - Install Redis to handle authentication https://redis.io/downloads/

```
$ git clone https://github.com/nicolasgarza/chat-app/
$ cd chat-app
$ redis-server
$ go run cmd/server/*.go
```

In a separate terminal window, run this command:
```
$ go run cmd/client/client.go
```

### Using the CLI
client.go is a CLI that spawns a new user and allows you to send messages.

### Adding users
Open another terminal shell and run client.go to spawn a new user in a chatroom.

### Sending messages
Once you run client.go, you will be prompted to send messages.

### Viewing messages
Open http://localhost:8080 in a browser to view the messages in different chatrooms.

### Terminating the sessions
Use Ctrl+c in the terminal session where the server is running to terminate the application.
