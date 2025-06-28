// pkg/config/config.go
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	RabbitMQURL string
}

func LoadConfig() *AppConfig {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, assuming environment variables are set.")
	}

	return &AppConfig{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
