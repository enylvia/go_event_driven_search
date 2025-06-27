// pkg/handler/news_handler.go
package handler

import (
	"context"
	"fmt"
	"net/http"
	"search_service/pkg/service"
	"search_service/pkg/util"
	"strconv"

	"github.com/gorilla/mux"
)

type NewsHandler struct {
	NewsService *service.NewsService
}

func NewNewsHandler(newsService *service.NewsService) *NewsHandler {
	return &NewsHandler{
		NewsService: newsService,
	}
}

// SearchNews menghandle request GET /news
func (h *NewsHandler) SearchNews(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	ctx := context.Background()
	news, total, err := h.NewsService.SearchNewsArticles(ctx, query, page, limit)
	if err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to search news articles", err.Error())
		return
	}

	response := map[string]interface{}{
		"total_hits": total,
		"page":       page,
		"limit":      limit,
		"articles":   news,
	}
	util.SendSuccessResponse(w, http.StatusOK, "News articles retrieved successfully", response)
}

func (h *NewsHandler) GetNewsByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		util.SendErrorResponse(w, http.StatusBadRequest, "News ID is required", nil)
		return
	}

	ctx := context.Background()
	news, err := h.NewsService.GetNewsArticleByID(ctx, id)
	if err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve news article", err.Error())
		return
	}

	if news == nil {
		util.SendErrorResponse(w, http.StatusNotFound, fmt.Sprintf("News article with ID '%s' not found", id), nil)
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, "News article retrieved successfully", news)
}
