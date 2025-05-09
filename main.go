package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"

	worker "go-worker-redis/cmd/worker"
	rabbit "go-worker-redis/infrastructure/rabbit"
	localRedis "go-worker-redis/infrastructure/redis"
	ws "go-worker-redis/infrastructure/websocket"
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

	go worker.StartWorker(rabbitCh, redisClient, ctx)

	e.Static("/public", "public")

	e.GET("/ws/:roomId", func(c echo.Context) error {
		return ws.HandleWebSocket(c, redisClient, ctx)
	})

	e.GET("/api/export-xlsx", func(c echo.Context) error {
		return handleXlsxFile(c, rabbitCh)
	})

	e.Logger.Fatal(e.Start(":8080"))
}

func handleXlsxFile(c echo.Context, rabbitCh *amqp.Channel) error {
	if rabbitCh == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Rabbit Channel is not available",
		})
	}

	taskID := fmt.Sprintf("task_%d", time.Now().UnixNano())

	taskMessage := map[string]string{
		"task_id": taskID,
		"action":  "generate_xlsx",
	}
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
		"taskId":  taskID,
		"time":    time.Now(),
		"message": "Export task queued successfully",
		"status":  "processing",
	})
}
