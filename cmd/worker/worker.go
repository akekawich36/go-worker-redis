package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"

	xlsx "go-worker-redis/pkg/xlsx"
)

type TaskMessage struct {
	TaskID string `json:"task_id"`
	Action string `json:"action"`
}

type TaskResponse struct {
	TaskID   string `json:"task_id"`
	Status   string `json:"status"`
	FileURL  string `json:"file_url,omitempty"`
	ErrorMsg string `json:"error_msg,omitempty"`
}

func StartWorker(rabbitCh *amqp.Channel, redisClient *redis.Client, ctx context.Context) {
	fmt.Println("Worker started...")

	// Declare the queue to ensure it exists
	_, err := rabbitCh.QueueDeclare(
		"file_processing", // queue name
		false,             // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Consume messages from the queue
	msgs, err := rabbitCh.Consume(
		"file_processing", // queue
		"",                // consumer
		false,             // auto-ack (we use manual ack)
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	fmt.Println("Worker is running...")

	// Process messages
	go func() {
		for msg := range msgs {
			var task TaskMessage
			err := json.Unmarshal(msg.Body, &task)
			if err != nil {
				log.Printf("Failed to parse message: %v\n", err)
				msg.Nack(false, false) // Reject without requeue
				continue
			}

			switch task.Action {
			case "generate_xlsx":
				fmt.Println("generate xlsx was running")
				processFile(task.TaskID, redisClient, ctx)
			default:
				log.Printf("Unknown action: %s", task.Action)
				msg.Nack(false, false) // Reject without requeue
				continue
			}

			// Acknowledge message
			msg.Ack(false)
		}
	}()

	select {} // Keep worker running
}

func processFile(TaskID string, redisClient *redis.Client, ctx context.Context) {
	fmt.Println("Start processing xlsx")

	defer redisClient.Close()

	filePath, err := xlsx.ProcessXlsx()

	taskResponse := TaskResponse{}

	if err != nil {
		taskResponse.Status = "error"
		taskResponse.ErrorMsg = err.Error()
		fmt.Printf("Error processing XLSX: %v\n", err)
	}

	if err != nil {
		fmt.Printf("Error marshaling response: %v\n", err)
		return
	} else {
		taskResponse.TaskID = TaskID
		taskResponse.Status = "Success"
		taskResponse.FileURL = fmt.Sprintf("/public/%s", filePath)
		fmt.Printf("XLSX processing completed: %s\n", filePath)
	}
	fmt.Println(taskResponse)
	responseJSON, err := json.Marshal(taskResponse)

	err = redisClient.Publish(ctx, "file_download", responseJSON).Err()
	if err != nil {
		fmt.Printf("Error publishing to Redis: %v\n", err)
		return
	}

	fmt.Println("File processing notification sent to clients")
}
