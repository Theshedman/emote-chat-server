package rabbitmq

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	ExchangeName = "chat_messages"
	QueueName    = "message"
)

type RabbitMQ struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

func New() (*RabbitMQ, error) {
	uri := os.Getenv("RABBITMQ_URL")
	if uri == "" {
		log.Fatal("RABBITMQ_URL environment variable not set")
	}

	for attempts := 1; ; attempts++ {
		conn, err := amqp091.Dial(uri)
		if err == nil {
			ch, err := conn.Channel()

			if err != nil {
				// Close connection on channel error
				err := conn.Close()
				if err != nil {
					return nil, fmt.Errorf("failed to close RabbitMQ connection: %w", err)
				}

				return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
			}

			return &RabbitMQ{conn: conn, channel: ch}, nil
		}

		waitTime := time.Duration(attempts*5) * time.Second // Adjust backoff
		log.Printf("Connection failed. Retrying in %v (attempt %d): %v", waitTime, attempts, err)
		time.Sleep(waitTime + time.Duration(rand.Intn(1000))*time.Millisecond) // Add jitter
	}
}

func (r *RabbitMQ) Close() {
	err := r.channel.Close()
	if err != nil {
		return
	}

	err = r.conn.Close()
	if err != nil {
		return
	}
}

func (r *RabbitMQ) DeclareExchange(name string) error {
	return r.channel.ExchangeDeclare(
		name,
		"fanout", // exchange type
		true,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	for { // Retry loop for publishing
		if err := r.channel.PublishWithContext(
			ctx,
			exchange,
			routingKey,
			false,
			false,
			amqp091.Publishing{
				ContentType: "application/json",
				Body:        body,
			},
		); err != nil {
			r.Close()
			r, err = New() // Attempt to reconnect
			if err != nil {
				return fmt.Errorf("failed to reconnect for publishing: %w", err)
			}

			time.Sleep(1 * time.Second) // Example: 1-second delay
		} else {
			return nil // Success!
		}
	}
}

func (r *RabbitMQ) Consume(exchange, queueName string) (<-chan amqp091.Delivery, error) {
	// Declare the queue
	_, err := r.channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind the queue to the exchange
	if err := r.channel.QueueBind(queueName, "", exchange, false, nil); err != nil {
		return nil, fmt.Errorf("failed to bind queue to exchange: %w", err)
	}

	// Consume messages from the queue
	msgs, err := r.channel.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to consume messages: %w", err)
	}

	// Reconnection Logic
	for {
		select {
		case _, ok := <-msgs:
			if !ok {
				// Channel closed, likely due to connection loss
				r.Close()
				r, err = New() // Attempt to reconnect
				if err != nil {
					return nil, fmt.Errorf("failed to reconnect: %w", err)
				}

				// Re-declare queue and re-bind
				_, err = r.channel.QueueDeclare(queueName, true, false, false, false, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to re-declare queue: %w", err)
				}
				err = r.channel.QueueBind(queueName, "", exchange, false, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to re-bind queue: %w", err)
				}

				// Retry Consume
				msgs, err = r.channel.Consume(queueName, "", true, false, false, false, nil)
				if err != nil {
					log.Printf("Failed to consume after reconnection: %v. Retrying...", err)
				}
			} else {
				return msgs, nil // Successful consumption
			}
		}
	}
}
