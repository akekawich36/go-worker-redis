package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"
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

var (
	rabbitConn *amqp.Connection
	rabbitCh   *amqp.Channel
)

func setupRabbitMQ() error {
	var err error
	rabbitConn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}

	rabbitCh, err = rabbitConn.Channel()
	if err != nil {
		return err
	}

	_, err = rabbitCh.QueueDeclare(
		"taskQueue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	e := echo.New()

	err := setupRabbitMQ()
	if err != nil {
		log.Fatalf("Failed to set up RabbitMQ: %s", err)
	}
	defer func() {
		if rabbitCh != nil {
			rabbitCh.Close()
		}
		if rabbitConn != nil {
			rabbitConn.Close()
		}
	}()

	e.Static("/public", "public")

	e.GET("/api/export-xlsx", handleXlsxFile)
	e.Logger.Fatal(e.Start(":8080"))
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
