package handler

import (
	"encoding/json"
	"fmt"
	"frappuccino/internal/check"
	"frappuccino/internal/service"
	"frappuccino/internal/utils"
	"frappuccino/models"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type OrderHandler struct {
	orderService *service.OrderService
	logger       *slog.Logger
}

func NewOrderHandler(orderService *service.OrderService, logFilePath string) (*OrderHandler, error) {
	logger, err := utils.SetupLogger(logFilePath)
	if err != nil {
		return nil, err
	}

	return &OrderHandler{
		orderService: orderService,
		logger:       logger,
	}, nil
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed!")

		return
	}
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		utils.SendError(w, utils.StatusConflict, "Failed to decode order to struct!")
		slog.Error("Failed to decode order to struct!", slog.Any("error", err))
		h.logger.Error("Failed to decode order to struct!", slog.Any("error", err))
		return
	}

	if !check.Check_Orders(w, r, order) {
		return
	}

	if err := h.orderService.CreateOrder(order); err != nil {
		utils.SendError(w, utils.StatusBadRequest, "Failed to create menu item!")
		slog.Error("Failed to create menu item!", slog.Any("error", err))
		h.logger.Error("Failed to create menu item!", slog.Any("error", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	slog.Info("Order created", "ID", order.ID)
	h.logger.Info("Order created", slog.String("ID", order.ID))
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed!")
		return
	}

	orders, err := h.orderService.List()
	if err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed to list orders!")
		slog.Error("Failed to list orders!", slog.Any("error", err))
		h.logger.Error("Failed to list orders!", slog.Any("error", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
	slog.Info("List of orders displayed")
	h.logger.Info("List of orders displayed")
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed!")
		return
	}

	orderID := r.URL.Path[len("/orders/"):]

	order, err := h.orderService.GetByID(orderID)
	if err != nil {
		utils.SendError(w, utils.StatusNotFound, "Order doesn't exist or unreachable!")
		slog.Error("Order doesn't exist or unreachable!", slog.Any("error", err))
		h.logger.Error("Order doesn't exist or unreachable!", slog.Any("error", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
	slog.Info("Got order by its id", "ID", orderID)
	h.logger.Info("Got order by its id", slog.String("ID", order.ID))
}

func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed!")
		return
	}
	orderID := r.URL.Path[len("/orders/"):]
	existingOrder, err := h.orderService.GetByID(orderID)
	if err != nil {
		utils.SendError(w, utils.StatusNotFound, "Order doesn't exist")
		slog.Error("Order doesn't exist", slog.Any("error", err))
		h.logger.Error("Order doesn't exist", slog.Any("error", err))
		return
	}
	if existingOrder.Status == "closed" {
		utils.SendError(w, utils.StatusConflict, "Cannot update a close order!")
		slog.Warn("Attempted to update a close order", "ID", orderID)
		h.logger.Error("Attempted to update a close order", slog.Any("error", err))
		return
	}
	var order models.Order

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		utils.SendError(w, utils.StatusConflict, "Failed to decode order to struct!")
		slog.Error("Failed to decode order to struct!", slog.Any("error", err))
		h.logger.Error("Failed to decode order to struct!", slog.Any("error", err))
		return
	}
	order.ID = orderID
	changeHistory := h.orderService.CollectOrderChanges(existingOrder, order)
	if len(changeHistory) > 0 {
		if err := h.orderService.RecordChangeHistory(orderID, changeHistory); err != nil {
			utils.SendError(w, utils.StatusInternalServerError, "Failed to record change history!")
			slog.Error("Failed to record change hidtory!", slog.Any("error", err))
			h.logger.Error("Failed to record change hidtory!", slog.Any("error", err))
			return
		}
	}
	if order.Status == "" {
		order.Status = existingOrder.Status
	}

	if order.CreatedAt == "" {
		order.CreatedAt = time.Now().Format(time.RFC3339)
	}
	if err := h.orderService.Update(order, orderID); err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed to update order!")
		slog.Error("Failed to update order!", slog.Any("error", err))
		h.logger.Error("Failed to update order!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	slog.Info("Order updated", "ID", orderID)
	h.logger.Info("Order updated", slog.String("ID", order.ID))
}

func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed!")
		return
	}

	orderID := r.URL.Path[len("/orders/"):]

	if err := h.orderService.Delete(orderID); err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed delete order!")
		slog.Error("Failed delete order!", slog.Any("error", err))
		h.logger.Error("Failed delete order!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	slog.Info("Order deleted", "ID", orderID)
	h.logger.Info("Order deleted", slog.String("ID", orderID))
}

func (h *OrderHandler) CloseOrder(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/orders/") || !strings.HasSuffix(r.URL.Path, "/close") {
		return
	}
	orderID := r.URL.Path[len("/orders/") : len(r.URL.Path)-len("/close")]
	fmt.Println(orderID)

	order, err := h.orderService.GetByID(orderID)
	if err != nil {
		utils.SendError(w, utils.StatusNotFound, "Order doesn't exist or unreachable!")
		slog.Error("Order doesn't exist or unreachable!", slog.Any("error", err))
		h.logger.Error("Order doesn't exist or unreachable!", slog.Any("error", err))
		return
	}

	order.Status = "completed"

	if err := h.orderService.Update(order, orderID); err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed to update order!")
		slog.Error("Failed to update order!", slog.Any("error", err))
		h.logger.Error("Failed to update order!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	slog.Info("Order closed", "ID", orderID)
	h.logger.Info("Order closed", slog.String("ID", orderID))
}

func (h *OrderHandler) GetOrderedItemsCount(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	// Если параметры пустые, передаём nil
	var startDatePtr, endDatePtr *string
	if startDate != "" {
		startDatePtr = &startDate
	}
	if endDate != "" {
		endDatePtr = &endDate
	}

	// Получаем данные
	counts, err := h.orderService.GetOrderedItemsCount(startDatePtr, endDatePtr)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get ordered items count: %v", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем JSON-ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

func (h *OrderHandler) BatchProcessOrders(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Orders []models.Order `json:"orders"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.orderService.ProcessBulkOrders(request.Orders)
	if err != nil {
		http.Error(w, "Failed to process orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
