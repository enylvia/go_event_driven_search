package main

import (
	"log"
	"search_service/pkg/app"
	"search_service/pkg/config"
	"search_service/pkg/repository"
)

func main() {
	cfg := config.LoadConfig()
	log.Printf("Loaded configurations: %+v", cfg)

	esRepo, err := repository.NewElasticSearchRepository(cfg)
	if err != nil {
		log.Fatalf("Failed to create Elasticsearch repository: %v", err)
	}
	err = esRepo.Ping()
	if err != nil {
		log.Fatalf("Failed to ping Elasticsearch: %v", err)
	}

	application := app.NewApplication(cfg, esRepo)
	application.StartApplication()

}
