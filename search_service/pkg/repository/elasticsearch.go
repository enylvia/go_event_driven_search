package repository

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v9"
	"log"
	"search_service/pkg/config"
)

type ElasticSearchRepository struct {
	Client *elasticsearch.Client
}

func NewElasticSearchRepository(cfg *config.AppConfig) (*ElasticSearchRepository, error) {
	esCfg := elasticsearch.Config{Addresses: []string{
		cfg.ElasticSearchURL,
	}}
	esClient, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, err
	}

	return &ElasticSearchRepository{Client: esClient}, nil
}

func (r *ElasticSearchRepository) Ping() error {
	res, err := r.Client.Info()
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error connecting to Elasticsearch: %s", res.String())
	}

	log.Println("Successfully connected to Elasticsearch!")
	return nil
}
