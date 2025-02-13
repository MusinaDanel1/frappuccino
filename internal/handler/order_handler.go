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
	"strconv"
	"strings"
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

	if err := h.orderService.CreateOrder(&order); err != nil {
		h.logger.Error("Failed to create order!", slog.Any("error", err))

		switch {
		case strings.Contains(err.Error(), "insufficient ingredient"):
			utils.SendError(w, utils.StatusBadRequest, err.Error())
		case strings.Contains(err.Error(), "order already exists"):
			utils.SendError(w, utils.StatusConflict, err.Error())
		default:
			utils.SendError(w, utils.StatusInternalServerError, "Failed to create order.")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
	slog.Info("Order created", "ID", order.ID)
	h.logger.Info("Order created", slog.Int("ID", order.ID))
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

	orderIDStr := r.URL.Path[len("/orders/"):]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid inventory ID", http.StatusBadRequest)
		return
	}

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
	h.logger.Info("Got order by its id", slog.Int("ID", order.ID))
}

func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed!")
		return
	}
	orderIDStr := r.URL.Path[len("/orders/"):]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid inventory ID", http.StatusBadRequest)
		return
	}
	existingOrder, err := h.orderService.GetByID(orderID)
	if err != nil {
		utils.SendError(w, utils.StatusNotFound, "Order doesn't exist")
		slog.Error("Order doesn't exist", slog.Any("error", err))
		h.logger.Error("Order doesn't exist", slog.Any("error", err))
		return
	}
	if existingOrder.Status == "completed" {
		utils.SendError(w, utils.StatusConflict, "Cannot update a completed order!")
		slog.Warn("Attempted to update a completed order", "ID", orderID)
		h.logger.Error("Attempted to update a completed order", slog.Any("error", err))
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
	if order.Status == "" {
		order.Status = existingOrder.Status
	}

	if err := h.orderService.Update(order, orderID); err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed to update order!")
		slog.Error("Failed to update order!", slog.Any("error", err))
		h.logger.Error("Failed to update order!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(order)
	slog.Info("Order updated", "ID", orderID)
	h.logger.Info("Order updated", slog.Int("ID", order.ID))
}

func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed!")
		return
	}

	orderIDStr := r.URL.Path[len("/orders/"):]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid inventory ID", http.StatusBadRequest)
		return
	}

	if err := h.orderService.Delete(orderID); err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed delete order!")
		slog.Error("Failed delete order!", slog.Any("error", err))
		h.logger.Error("Failed delete order!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	slog.Info("Order deleted", "ID", orderID)
	h.logger.Info("Order deleted", slog.Int("ID", orderID))
}

func (h *OrderHandler) CloseOrder(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/orders/") || !strings.HasSuffix(r.URL.Path, "/close") {
		return
	}
	orderIDStr := r.URL.Path[len("/orders/") : len(r.URL.Path)-len("/close")]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid inventory ID", http.StatusBadRequest)
		return
	}
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
	slog.Info("Order completed", "ID", orderID)
	h.logger.Info("Order completed", slog.Int("ID", orderID))
}

func (h *OrderHandler) GetOrderedItemsCount(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	startDatePtr, endDatePtr, ok := check.Check_Date(w, r, startDate, endDate)
	if !ok {
		return
	}

	counts, err := h.orderService.GetOrderedItemsCount(startDatePtr, endDatePtr)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get ordered items count: %v", err), http.StatusInternalServerError)
		return
	}

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

	for _, order := range request.Orders {
		if !check.Check_Orders(w, r, order) {
			return
		}
	}

	response, err := h.orderService.ProcessBulkOrders(request.Orders)
	if err != nil {
		http.Error(w, "Failed to process orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
