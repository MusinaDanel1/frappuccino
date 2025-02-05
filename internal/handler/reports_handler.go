package handler

import (
	"database/sql"
	"encoding/json"
	"frappuccino/internal/service"
	"frappuccino/internal/utils"
	"log/slog"
	"net/http"
	"time"
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

func (h *ReportsHandler) GetTotalSales(w http.ResponseWriter, r *http.Request) {
	total, err := h.service.TotalSales()
	if err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed to calculate total sales!")
		h.logger.Error("Failed to calculate total sales!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"total_sales": total})
}

func (h *ReportsHandler) GetPopularItems(w http.ResponseWriter, r *http.Request) {
	popularity, err := h.service.PopularItems()
	if err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed to fetch popular items!")
		h.logger.Error("Failed to fetch popular items!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(popularity)
}

func GetNumberOfOrderedItems(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		startDate := r.URL.Query().Get("startDate")
		endDate := r.URL.Query().Get("endDate")

		// Validate date format (optional)
		if startDate != "" {
			if _, err := time.Parse("2006-01-02", startDate); err != nil {
				http.Error(w, "Invalid startDate format. Use YYYY-MM-DD.", http.StatusBadRequest)
				return
			}
		}
		if endDate != "" {
			if _, err := time.Parse("2006-01-02", endDate); err != nil {
				http.Error(w, "Invalid endDate format. Use YYYY-MM-DD.", http.StatusBadRequest)
				return
			}
		}

		// Execute SQL query
		rows, err := db.Query(`
            SELECT 
                mi.name AS item_name,
                SUM(oi.quantity) AS total_quantity
            FROM 
                order_items oi
            JOIN 
                menu_items mi ON oi.menu_item_id = mi.menu_item_id
            JOIN 
                orders o ON oi.order_id = o.order_id
            WHERE 
                ($1 = '' OR o.created_at >= $1::timestamp) AND 
                ($2 = '' OR o.created_at <= $2::timestamp)
            GROUP BY 
                mi.name;
        `, startDate, endDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Prepare response
		result := make(map[string]int)
		for rows.Next() {
			var itemName string
			var totalQuantity int
			if err := rows.Scan(&itemName, &totalQuantity); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			result[itemName] = totalQuantity
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
