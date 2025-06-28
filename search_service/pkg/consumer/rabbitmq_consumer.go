package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"search_service/pkg/config"
	"search_service/pkg/model"
	"search_service/pkg/service"
)

const (
	QueueName = "news_queue"
)

type RabbitMQConsumer struct {
	conn        *amqp091.Connection
	channel     *amqp091.Channel
	NewsService *service.NewsService
	Config      *config.AppConfig
}

func NewRabbitMQConsumer(cfg *config.AppConfig, newsService *service.NewsService) (*RabbitMQConsumer, error) {
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

	log.Printf("Successfully connected to RabbitMQ and declared queue '%s' (Consumer)", QueueName)

	return &RabbitMQConsumer{
		conn:        conn,
		channel:     ch,
		NewsService: newsService,
		Config:      cfg,
	}, nil
}

func (c *RabbitMQConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
		log.Println("RabbitMQ consumer channel closed.")
	}
	if c.conn != nil {
		c.conn.Close()
		log.Println("RabbitMQ consumer connection closed.")
	}
}

func (c *RabbitMQConsumer) StartConsuming(ctx context.Context) {
	msgs, err := c.channel.Consume(
		QueueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Println(" [*] Waiting for messages. To exit, press CTRL+C in search_service terminal.")

	go func() {
		for d := range msgs {
			log.Printf(" [x] Received a message: %s", string(d.Body))

			var event model.NewsEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("Error unmarshaling message: %v. Rejecting message.", err)
				d.Nack(false, false)
				continue
			}

			switch event.Type {
			case "CREATED":
				log.Printf("Processing CREATED event for ID: %s", event.Payload.ID)
				err = c.NewsService.IndexNews(ctx, event.Payload)
			case "UPDATED":
				log.Printf("Processing UPDATED event for ID: %s", event.Payload.ID)
				payloadMap := make(map[string]interface{})
				jsonBytes, _ := json.Marshal(event.Payload)
				json.Unmarshal(jsonBytes, &payloadMap)
				err = c.NewsService.UpdateNews(ctx, event.Payload.ID, payloadMap)
			case "DELETED":
				log.Printf("Processing DELETED event for ID: %s", event.Payload.ID)
				err = c.NewsService.DeleteNews(ctx, event.Payload.ID)
			default:
				log.Printf("Unknown event type: %s for ID: %s. Rejecting message.", event.Type, event.Payload.ID)
				d.Nack(false, false)
				continue
			}

			if err != nil {
				log.Printf("Failed to process event '%s' for ID %s: %v. Message will be Nacked (and potentially re-queued/dead-lettered).", event.Type, event.Payload.ID, err)
				d.Nack(false, true)
			} else {
				log.Printf("Successfully processed event '%s' for ID: %s. Acknowledging message.", event.Type, event.Payload.ID)
				d.Ack(false)
			}
		}
	}()
}
