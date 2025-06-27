package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esapi"
	"log"
	"net/http"
	"search_service/pkg/config"
	"search_service/pkg/model"
	"strings"
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
func (r *ElasticSearchRepository) IndexDocument(ctx context.Context, indexName string, doc model.DocumentNews) error {
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: doc.ID,
		Body:       strings.NewReader(string(docJSON)),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.Client)
	if err != nil {
		return fmt.Errorf("failed to perform index request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch returned an error during indexing: %s", res.String())
	}

	log.Printf("Document ID %s indexed successfully to index '%s'.", doc.ID, indexName)
	return nil
}

func (r *ElasticSearchRepository) SearchDocuments(ctx context.Context, indexName string, query string, size, from int) ([]model.DocumentNews, int64, error) {
	var buf bytes.Buffer
	// Buat query dasar
	searchBody := map[string]interface{}{}
	if query == "" {
		searchBody["query"] = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	} else {
		searchBody["query"] = map[string]interface{}{
			"match": map[string]interface{}{
				"content": map[string]interface{}{
					"query":     query,
					"fuzziness": "AUTO",
				},
			},
		}
	}

	searchBody["size"] = size
	searchBody["from"] = from

	if err := json.NewEncoder(&buf).Encode(searchBody); err != nil {
		return nil, 0, fmt.Errorf("failed to encode search query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{indexName},
		Body:  &buf,
	}

	res, err := req.Do(ctx, r.Client)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to perform search request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("elasticsearch returned an error during search: %s", res.String())
	}

	var rMap map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&rMap); err != nil {
		return nil, 0, fmt.Errorf("failed to parse search response: %w", err)
	}

	var news []model.DocumentNews
	totalHits := int64(0)

	hits, found := rMap["hits"].(map[string]interface{})
	if found {
		total := hits["total"].(map[string]interface{})["value"].(float64)
		totalHits = int64(total)

		for _, hit := range hits["hits"].([]interface{}) {
			source := hit.(map[string]interface{})["_source"]
			jsonBytes, _ := json.Marshal(source)
			var doc model.DocumentNews
			if err := json.Unmarshal(jsonBytes, &doc); err != nil {
				log.Printf("Warning: Failed to unmarshal document from search result: %v", err)
				continue
			}
			news = append(news, doc)
		}
	}

	return news, totalHits, nil
}
func (r *ElasticSearchRepository) GetDocumentByID(ctx context.Context, indexName, docID string) (*model.DocumentNews, error) {
	req := esapi.GetRequest{
		Index:      indexName,
		DocumentID: docID,
	}

	res, err := req.Do(ctx, r.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to perform get document request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return nil, nil // Dokumen tidak ditemukan
		}
		return nil, fmt.Errorf("elasticsearch returned an error during get document: %s", res.String())
	}

	var rMap map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&rMap); err != nil {
		return nil, fmt.Errorf("failed to parse get document response: %w", err)
	}

	found, _ := rMap["found"].(bool)
	if !found {
		return nil, nil
	}

	source := rMap["_source"]
	jsonBytes, err := json.Marshal(source)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal source to JSON: %w", err)
	}

	var doc model.DocumentNews
	if err := json.Unmarshal(jsonBytes, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document from get result: %w", err)
	}

	return &doc, nil
}
func (r *ElasticSearchRepository) UpdateDocument(ctx context.Context, indexName string, docID string, updates map[string]interface{}) error {
	updateBody := map[string]interface{}{
		"doc": updates,
	}
	updateJSON, err := json.Marshal(updateBody)
	if err != nil {
		return fmt.Errorf("failed to marshal update document: %w", err)
	}

	req := esapi.UpdateRequest{
		Index:      indexName,
		DocumentID: docID,
		Body:       strings.NewReader(string(updateJSON)),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.Client)
	if err != nil {
		return fmt.Errorf("failed to perform update request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch returned an error during update: %s", res.String())
	}

	log.Printf("Document ID %s updated successfully in index '%s'.", docID, indexName)
	return nil
}
func (r *ElasticSearchRepository) DeleteDocument(ctx context.Context, indexName, docID string) error {
	req := esapi.DeleteRequest{
		Index:      indexName,
		DocumentID: docID,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.Client)
	if err != nil {
		return fmt.Errorf("failed to perform delete request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			log.Printf("Document ID %s not found in index '%s'. Nothing to delete.", docID, indexName)
			return nil
		}
		return fmt.Errorf("elasticsearch returned an error during delete: %s", res.String())
	}

	log.Printf("Document ID %s deleted successfully from index '%s'.", docID, indexName)
	return nil
}
