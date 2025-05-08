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

	// room := GetRoom(roomId)

	fmt.Println("Connected!")

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv from %s: %s", roomId, message)
	}

	subscriber := redisClient.Subscribe(ctx, "file_download")
	defer subscriber.Close()

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

	return room
}
