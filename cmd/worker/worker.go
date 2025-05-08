package worker

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
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

func StartWorker(rabbitCh *amqp.Channel, redisClient *redis.Client) {
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

// func main() {
// 	// Connect to RabbitMQ
// 	rabbitConn, rabbitCh := rabbit.SetupRabbitMQ()
// 	defer rabbitConn.Close()
// 	defer rabbitCh.Close()

// 	// Redis Setup
// 	redisClient := localRedis.SetupRedis()
// 	defer redisClient.Close()

// 	// Declare the same queue as in the main.go
// 	q, err := rabbitCh.QueueDeclare(
// 		"file_processing", // queue name
// 		false,             // durable
// 		false,             // delete when unused
// 		false,             // exclusive
// 		false,             // no-wait
// 		nil,               // arguments
// 	)
// 	if err != nil {
// 		log.Fatalf("Failed to declare a queue: %s", err)
// 	}

// 	fmt.Println(q.Name)

// 	// Consume messages from the queue
// 	msgs, err := rabbitCh.Consume(
// 		q.Name, // queue
// 		"",     // consumer
// 		false,  // auto-ack
// 		false,  // exclusive
// 		false,  // no-local
// 		false,  // no-wait
// 		nil,    // args
// 	)
// 	if err != nil {
// 		log.Fatalf("Failed to register a consumer: %s", err)
// 	}

// 	err = rabbitCh.Qos(
// 		1,     // prefetch count
// 		0,     // prefetch size
// 		false, // global
// 	)

// 	log.Println("Worker started. Waiting for messages...", msgs)

// 	// Process messages in a goroutine
// 	forever := make(chan bool)
// 	go processMessages(msgs)

// 	<-forever // Keep the main function running
// }

// func processMessages(msgs <-chan amqp.Delivery) {
// 	for d := range msgs {
// 		log.Printf("Received a message: %s", d.Body)

// 		// Parse the task request
// 		var task TaskRequest
// 		err := json.Unmarshal(d.Body, &task)
// 		if err != nil {
// 			log.Printf("Error parsing task: %s", err)
// 			d.Nack(false, false) // Negative acknowledgment, don't requeue
// 			continue
// 		}

// 		// Generate the Excel file using your existing function
// 		result := xlsx.ProcessXlsx()

// 		// Create the task response
// 		response := &TaskResponse{}

// 		// Set response fields based on result
// 		if result.Status {
// 			response.Status = "completed"
// 			response.FileURL = result.FileURL
// 		} else {
// 			response.Status = "failed"
// 			response.ErrorMsg = result.Message
// 		}

// 		// Convert response to JSON
// 		if err != nil {
// 			log.Printf("Error marshaling response: %s", err)
// 			d.Nack(false, true) // Negative acknowledgment, requeue
// 			continue
// 		}

// 		// Store the result in Redis
// 		// ctx := context.Background()
// 		// err = redisClient.Set(ctx, "task:"+task.TaskID, responseJSON, 24*time.Hour).Err()
// 		// if err != nil {
// 		// 	log.Printf("Error storing result in Redis: %s", err)
// 		// 	d.Nack(false, true) // Negative acknowledgment, requeue
// 		// 	continue
// 		// }

// 		// // Publish a notification to Redis pub/sub
// 		// err = redisClient.Publish(ctx, "task_updates", responseJSON).Err()
// 		// if err != nil {
// 		// 	log.Printf("Error publishing notification: %s", err)
// 		// 	// Continue anyway, as the task was processed successfully
// 		// }

// 		log.Printf("Task %s processed successfully", task.TaskID)
// 		d.Ack(false) // Acknowledge the message
// 	}
// }
