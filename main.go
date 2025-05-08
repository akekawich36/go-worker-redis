package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"

	rabbit "go-worker-redis/infrastructure/rabbit"
	localRedis "go-worker-redis/infrastructure/redis"
)

type TaskResponse struct {
	TaskID   string `json:"task_id"`
	Status   string `json:"status"`
	FileURL  string `json:"file_url,omitempty"`
	ErrorMsg string `json:"error_msg,omitempty"`
}

type TaskRequest struct {
	TaskID string `json:"task_id"`
}

var ctx = context.Background()

var (
	rabbitConn  *amqp.Connection
	rabbitCh    *amqp.Channel
	redisClient *redis.Client
	upgrader    = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	e := echo.New()

	// RabbitMq Setup
	rabbitConn, rabbitCh := rabbit.SetupRabbitMQ()
	defer rabbitConn.Close()
	defer rabbitCh.Close()

	// Redis Setup
	redisClient := localRedis.SetupRedis()
	defer redisClient.Close()

	e.Static("/public", "public")

	e.GET("/ws", handleWebSocket)
	e.GET("/api/export-xlsx", handleXlsxFile)

	e.Logger.Fatal(e.Start(":8080"))
}

func handleWebSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	fmt.Println("Connected!")

	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = ws.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}

	subscriber := redisClient.Subscribe(ctx, "file_download")
	defer subscriber.Close()

	return nil
}

func handleXlsxFile(c echo.Context) error {
	taskMessage := map[string]string{"message": "generate_xlsx"}
	taskBytes, _ := json.Marshal(taskMessage)

	err := rabbitCh.Publish(
		"",
		"file_processing",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        taskBytes,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to publish task to queue",
		})
	}

	// Return the task ID to the client
	return c.JSON(http.StatusOK, map[string]interface{}{
		"time":    time.Now(),
		"message": "Export task queued successfully",
		"status":  "processing",
	})
}
