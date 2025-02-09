package handler

import (
	"context"
	"encoding/json"
	"frappuccino/internal/service"
	"frappuccino/internal/utils"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type ReportsHandler struct {
	service *service.ReportsService
	logger  *slog.Logger
}

func NewReportsHandler(service *service.ReportsService, logFilePath string) (*ReportsHandler, error) {
	logger, err := utils.SetupLogger(logFilePath)
	if err != nil {
		return nil, err
	}

	return &ReportsHandler{
		service: service,
		logger:  logger,
	}, nil
}

func (h *ReportsHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	filters := strings.Split(r.URL.Query().Get("filter"), ",")
	if len(filters) == 1 && filters[0] == "" {
		filters = []string{"all"}
	}

	minPrice, _ := strconv.ParseFloat(r.URL.Query().Get("minPrice"), 64)
	maxPrice, _ := strconv.ParseFloat(r.URL.Query().Get("maxPrice"), 64)

	ctx := context.Background()
	response, err := h.service.Search(ctx, query, filters, minPrice, maxPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
