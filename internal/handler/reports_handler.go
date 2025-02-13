package handler

import (
	"context"
	"encoding/json"
	"frappuccino/internal/service"
	"frappuccino/internal/utils"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type ReportHandler struct {
	service *service.ReportService
	logger  *slog.Logger
}

func NewReportsHandler(service *service.ReportService, logFilePath string) (*ReportHandler, error) {
	logger, err := utils.SetupLogger(logFilePath)
	if err != nil {
		return nil, err
	}

	return &ReportHandler{
		service: service,
		logger:  logger,
	}, nil
}

func (h *ReportHandler) GetTotalSales(w http.ResponseWriter, r *http.Request) {
	// Создаем контекст для запроса
	ctx := context.Background()

	total, err := h.service.GetTotalSales(ctx)
	if err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed to calculate total sales!")
		h.logger.Error("Failed to calculate total sales!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"total_sales": total})
}

func (h *ReportHandler) GetPopularItems(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	popularity, err := h.service.GetPopularItems(ctx)
	if err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed to fetch popular items!")
		h.logger.Error("Failed to fetch popular items!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(popularity)
}

func (h *ReportHandler) GetOrderedItemsByPeriod(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	query := r.URL.Query()
	period := query.Get("period")
	month := query.Get("month")
	year := query.Get("year")

	if period != "day" && period != "month" {
		utils.SendError(w, utils.StatusBadRequest, "Invalid period parameter! Must be 'day' or 'month'.")
		h.logger.Error("Invalid period parameter", slog.String("period", period))
		return
	}

	if period == "day" && month == "" {
		utils.SendError(w, utils.StatusBadRequest, "Month parameter is required when period=day.")
		h.logger.Error("Missing month parameter for period=day")
		return
	}

	if period == "month" && year == "" {
		utils.SendError(w, utils.StatusBadRequest, "Year parameter is required when period=month.")
		h.logger.Error("Missing year parameter for period=month")
		return
	}

	orders, err := h.service.GetOrderedItemsByPeriod(ctx, period, month, year)
	if err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed to fetch ordered items!")
		h.logger.Error("Failed to fetch ordered items!", slog.Any("error", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]map[string]int{"orders_by_period": orders})
}

func (h *ReportHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	filters := strings.Split(r.URL.Query().Get("filter"), ",")
	if len(filters) == 1 && filters[0] == "" {
		filters = []string{"all"}
	}

	minPrice := math.Inf(-1)
	if minStr := r.URL.Query().Get("minPrice"); minStr != "" {
		minPrice, _ = strconv.ParseFloat(minStr, 64)
	}

	maxPrice := math.Inf(1)
	if maxStr := r.URL.Query().Get("maxPrice"); maxStr != "" {
		maxPrice, _ = strconv.ParseFloat(maxStr, 64)
	}

	ctx := context.Background()
	response, err := h.service.Search(ctx, query, filters, minPrice, maxPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
