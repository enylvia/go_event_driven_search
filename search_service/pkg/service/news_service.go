package service

import (
	"context"
	"fmt"
	"log" // Ditambahkan
	"search_service/pkg/model"
	"search_service/pkg/repository"
	"time"
)

type NewsService struct {
	ESRepo    *repository.ElasticSearchRepository
	IndexName string // Nama indeks yang akan digunakan
}

func NewNewsService(esRepo *repository.ElasticSearchRepository) *NewsService {
	return &NewsService{
		ESRepo:    esRepo,
		IndexName: "news_articles",
	}
}

func (s *NewsService) IndexNews(ctx context.Context, doc model.DocumentNews) error {
	// Di sini bisa ada validasi data atau transformasi sebelum diindeks
	log.Printf("Service: Indexing news document with ID: %s", doc.ID)
	now := time.Now()
	doc.CreatedAt = now
	return s.ESRepo.IndexDocument(ctx, s.IndexName, doc)
}

func (s *NewsService) SearchNewsArticles(ctx context.Context, query string, page, limit int) ([]model.DocumentNews, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	from := (page - 1) * limit

	log.Printf("Service: Searching news for query '%s', page %d, limit %d", query, page, limit)
	news, total, err := s.ESRepo.SearchDocuments(ctx, s.IndexName, query, limit, from)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search news articles: %w", err)
	}
	return news, total, nil
}

func (s *NewsService) GetNewsArticleByID(ctx context.Context, id string) (*model.DocumentNews, error) {
	log.Printf("Service: Getting news document by ID: %s", id)
	doc, err := s.ESRepo.GetDocumentByID(ctx, s.IndexName, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get news article by ID: %w", err)
	}
	return doc, nil
}

// UpdateNews (akan dipanggil oleh consumer RabbitMQ)
func (s *NewsService) UpdateNews(ctx context.Context, docID string, updates map[string]interface{}) error {
	now := time.Now()
	updates["updated_at"] = now

	log.Printf("Service: Updating news document with ID: %s", docID)
	return s.ESRepo.UpdateDocument(ctx, s.IndexName, docID, updates)
}

// DeleteNews (akan dipanggil oleh consumer RabbitMQ)
func (s *NewsService) DeleteNews(ctx context.Context, docID string) error {
	log.Printf("Service: Deleting news document with ID: %s", docID)
	return s.ESRepo.DeleteDocument(ctx, s.IndexName, docID)
}
