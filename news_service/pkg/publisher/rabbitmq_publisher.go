package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"news_service/pkg/config"
	"news_service/pkg/model"
	"news_service/pkg/util"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const (
	QueueName = "news_queue"
)

type RabbitMQPublisher struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	logger  *util.Logger
}

func NewRabbitMQPublisher(cfg *config.AppConfig, logger *util.Logger) (*RabbitMQPublisher, error) {
	conn, err := amqp091.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue '%s': %w", QueueName, err)
	}

	logger.Printf("Successfully connected to RabbitMQ and declared queue '%s'", QueueName)

	return &RabbitMQPublisher{
		conn:    conn,
		channel: ch,
		logger:  logger,
	}, nil
}

func (p *RabbitMQPublisher) Close() {
	if p.channel != nil {
		p.channel.Close()
		p.logger.Println("RabbitMQ channel closed.")
	}
	if p.conn != nil {
		p.conn.Close()
		p.logger.Println("RabbitMQ connection closed.")
	}
}

func (p *RabbitMQPublisher) PublishNewsEvent(ctx context.Context, event model.NewsEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal NewsEvent to JSON: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = p.channel.PublishWithContext(ctx,
		"",
		QueueName,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message to RabbitMQ: %w", err)
	}

	p.logger.Printf(" [x] Sent '%s' event for ID: %s", event.Type, event.Payload.ID)
	return nil
}
