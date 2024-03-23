package rabbitmq

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"os"
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
	conn, err := amqp091.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	return &RabbitMQ{conn: conn, channel: ch}, nil
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
	if err := r.channel.PublishWithContext(
		ctx,
		exchange,   // exchange name
		routingKey, // routing key (RoomID in our case)
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
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

	return msgs, nil
}
