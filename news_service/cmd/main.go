package main

import (
	"context"
	"fmt"
	"news_service/pkg/config"
	"news_service/pkg/model"
	"news_service/pkg/publisher"
	"news_service/pkg/util"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := util.NewLogger()
	logger.Println("Starting News Service Publisher...")

	cfg := config.LoadConfig()
	logger.Printf("Loaded configurations: %+v", cfg)

	rmqPublisher, err := publisher.NewRabbitMQPublisher(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize RabbitMQ publisher: %v", err)
	}
	defer rmqPublisher.Close()

	logger.Println("RabbitMQ Publisher is ready. Sending simulated events...")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		for i := 1; ; i++ {
			select {
			case <-ctx.Done():
				logger.Println("Stopping event sending due to shutdown signal.")
				return
			default:
				doc := model.DocumentNews{
					ID:          fmt.Sprintf("simulated_news_%d", i),
					Title:       fmt.Sprintf("Simulated News Article %d: Update Penting", i),
					Content:     fmt.Sprintf("Ini adalah konten berita simulasi ke-%d. Mari kita cek apakah terindeks dengan benar.", i),
					Author:      "Simulator",
					CreatedAt:   time.Now(),
					Tags:        []string{"simulasi", "rabbitmq", fmt.Sprintf("tag-%d", i)},
					PublishedAt: time.Now(),
				}

				event := model.NewsEvent{
					Type:      "CREATED",
					Timestamp: time.Now(),
					Payload:   doc,
				}

				if err := rmqPublisher.PublishNewsEvent(ctx, event); err != nil {
					logger.Printf("Error publishing CREATE event for ID %s: %v", event.Payload.ID, err)
				}

				time.Sleep(3 * time.Second)

				if i == 1 {
					updatedDoc := model.DocumentNews{
						ID:          "simulated_news_1",
						Title:       "Simulated News Article 1: Content Diperbarui & Disempurnakan!",
						Content:     "Konten berita simulasi pertama telah diperbarui. Perubahan ini harus terlihat di Elasticsearch setelah indexing.",
						Author:      "Simulator Update",
						Tags:        []string{"simulasi", "rabbitmq", "update", "penting", "disempurnakan"},
						PublishedAt: time.Now().Add(-24 * time.Hour),
						UpdatedAt:   func() *time.Time { t := time.Now(); return &t }(),
					}
					updateEvent := model.NewsEvent{
						Type:      "UPDATED",
						Timestamp: time.Now(),
						Payload:   updatedDoc,
					}
					if err := rmqPublisher.PublishNewsEvent(ctx, updateEvent); err != nil {
						logger.Printf("Error publishing UPDATE event for ID %s: %v", updateEvent.Payload.ID, err)
					}
					time.Sleep(2 * time.Second)
				}

				if i == 3 {
					deleteEvent := model.NewsEvent{
						Type:      "DELETED",
						Timestamp: time.Now(),
						Payload:   model.DocumentNews{ID: "simulated_news_3"}, // Cukup ID
					}
					if err := rmqPublisher.PublishNewsEvent(ctx, deleteEvent); err != nil {
						logger.Printf("Error publishing DELETE event for ID %s: %v", deleteEvent.Payload.ID, err)
					}
					time.Sleep(2 * time.Second)
				}
			}
		}
	}()

	logger.Println("News Service Publisher is running. Press Ctrl+C to stop.")
	<-ctx.Done()
	logger.Println("News Service Publisher stopped.")
}
