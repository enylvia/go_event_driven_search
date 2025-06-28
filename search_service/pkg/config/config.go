package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type AppConfig struct {
	AppPort          string
	ElasticSearchURL string
	RabbitMQURL      string
}

func LoadConfig() *AppConfig {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found, assuming environment variables are set.")
	}
	return &AppConfig{
		AppPort:          getEnv("APP_PORT", "8080"),
		ElasticSearchURL: getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
