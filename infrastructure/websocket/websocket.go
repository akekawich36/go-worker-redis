package websocket

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Room struct {
	ID         string
	Clients    map[*websocket.Conn]bool
	Broadcast  chan []byte
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
	mutex      sync.Mutex
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	rooms      = make(map[string]*Room)
	roomsMutex sync.Mutex
)

func HandleWebSocket(c echo.Context, redisClient *redis.Client, ctx context.Context) error {
	roomId := c.Param("roomId")

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	room := GetRoom(roomId)

	room.mutex.Lock()
	room.Clients[ws] = true
	room.mutex.Unlock()

	fmt.Println("WebSocket Connected!")

	defer func() {
		room.mutex.Lock()
		delete(room.Clients, ws)
		room.mutex.Unlock()
		ws.Close()
		fmt.Println("WebSocket Disconnected!")
	}()

	pubsub := redisClient.Subscribe(ctx, "file_download")
	defer pubsub.Close()

	// Start a goroutine to listen for Redis messages
	go func() {
		channel := pubsub.Channel()
		for msg := range channel {
			log.Printf("Received Redis message: %s", msg.Payload)

			// Forward message to all clients in the room
			room.mutex.Lock()
			for client := range room.Clients {
				err := client.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
				if err != nil {
					log.Printf("error sending message to client: %v", err)
				}
			}
			room.mutex.Unlock()
		}
	}()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		log.Printf("received from client in room %s: %s", roomId, message)

		// You can handle client messages here if needed
	}
	return nil
}

func GetRoom(roomID string) *Room {
	roomsMutex.Lock()
	defer roomsMutex.Unlock()

	if room, exists := rooms[roomID]; exists {
		return room
	}

	// Create a new room
	room := &Room{
		ID:         roomID,
		Clients:    make(map[*websocket.Conn]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}

	rooms[roomID] = room

	fmt.Printf("Created new room: %s\n", roomID)
	return room
}
