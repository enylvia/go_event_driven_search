package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"search_service/pkg/repository"
	"search_service/pkg/util"
	"strings"
)

type AdminHandler struct {
	ESRepo *repository.ElasticSearchRepository
}

func NewAdminHandler(esRepo *repository.ElasticSearchRepository) *AdminHandler {
	return &AdminHandler{
		ESRepo: esRepo,
	}
}
func (h *AdminHandler) GetElasticsearchInfo(w http.ResponseWriter, r *http.Request) {
	res, err := h.ESRepo.Client.Info()
	if err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to get Elasticsearch info", err.Error())
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		util.SendErrorResponse(w, res.StatusCode, "Elasticsearch returned an error", res.String())
		return
	}

	var rMap map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&rMap); err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to parse Elasticsearch info response", err.Error())
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, "Elasticsearch info retrieved successfully", rMap)
}

func (h *AdminHandler) CreateIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	indexName := vars["name"]

	if indexName == "" {
		util.SendErrorResponse(w, http.StatusBadRequest, "Index name is required", nil)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		util.SendErrorResponse(w, http.StatusBadRequest, "Failed to read request body", err.Error())
		return
	}
	if !json.Valid(bodyBytes) {
		util.SendErrorResponse(w, http.StatusBadRequest, "Invalid JSON in request body", nil)
		return
	}
	existsRes, err := h.ESRepo.Client.Indices.Exists([]string{indexName}, h.ESRepo.Client.Indices.Exists.WithContext(context.Background()))
	if err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to check index existence", err.Error())
		return
	}
	defer existsRes.Body.Close()
	if !existsRes.IsError() {
		log.Printf("Index '%s' already exists. Deleting it first.", indexName)
		deleteRes, err := h.ESRepo.Client.Indices.Delete([]string{indexName}, h.ESRepo.Client.Indices.Delete.WithContext(context.Background()))
		if err != nil {
			util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to delete existing index", err.Error())
			return
		}
		defer deleteRes.Body.Close()
		if deleteRes.IsError() {
			util.SendErrorResponse(w, http.StatusInternalServerError, "Error response from deleting existing index", deleteRes.String())
			return
		}
		log.Printf("Index '%s' deleted successfully.", indexName)
	}
	createRes, err := h.ESRepo.Client.Indices.Create(
		indexName,
		h.ESRepo.Client.Indices.Create.WithBody(strings.NewReader(string(bodyBytes))),
		h.ESRepo.Client.Indices.Create.WithContext(context.Background()),
	)
	if err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to create index", err.Error())
		return
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Error response from Elasticsearch when creating index", createRes.String())
		return
	}

	util.SendSuccessResponse(w, http.StatusCreated, fmt.Sprintf("Index '%s' created successfully", indexName), nil)
}

func (h *AdminHandler) DeleteIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	indexName := vars["name"]
	if indexName == "" {
		util.SendErrorResponse(w, http.StatusBadRequest, "Index name is required", nil)
		return
	}

	deleteRes, err := h.ESRepo.Client.Indices.Delete([]string{indexName}, h.ESRepo.Client.Indices.Delete.WithContext(context.Background()))
	if err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to send delete index request", err.Error())
		return
	}
	defer deleteRes.Body.Close()

	if deleteRes.IsError() {
		if deleteRes.StatusCode == http.StatusNotFound {
			util.SendErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Index '%s' not found", indexName), deleteRes.String())
			return
		}
		util.SendErrorResponse(w, http.StatusInternalServerError, "Error response from Elasticsearch when deleting index", deleteRes.String())
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, fmt.Sprintf("Index '%s' deleted successfully", indexName), nil)
}
func (h *AdminHandler) GetExistElasticIndex(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	indexName := vars["name"]
	if indexName == "" {
		util.SendErrorResponse(w, http.StatusBadRequest, "Index name is required", nil)
		return
	}
	res, err := h.ESRepo.Client.Indices.Get([]string{indexName}, h.ESRepo.Client.Indices.Get.WithContext(context.Background()))
	if err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to get Elasticsearch info", err.Error())
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		util.SendErrorResponse(w, res.StatusCode, "Elasticsearch returned an error", res.String())
		return
	}

	var rMap map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&rMap); err != nil {
		util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to parse Elasticsearch info response", err.Error())
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, "Elasticsearch info retrieved successfully", rMap)
}
