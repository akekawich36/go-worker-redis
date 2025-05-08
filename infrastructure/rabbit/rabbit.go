package rabbit

import (
	"log"

	"github.com/streadway/amqp"
)

func SetupRabbitMQ() (*amqp.Connection, *amqp.Channel) {
	rabbitConn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	rabbitCh, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}

	_, err = rabbitCh.QueueDeclare(
		"file_processing",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %s", err)
	}

	return rabbitConn, rabbitCh
}
