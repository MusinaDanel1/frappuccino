package check

import (
	"frappuccino/internal/utils"
	"frappuccino/models"
	"net/http"
)

func Check_Orders(w http.ResponseWriter, r *http.Request, orders models.Order) bool {
	if orders.CustomerName == "" {
		utils.SendError(w, utils.StatusBadRequest, "Empty Customer name in inventory items!")
		return false
	}
	for i := range orders.Items {
		Check_OrderItem(w, r, orders.Items[i])
	}

	if len(orders.Items) <= 0 {

		utils.SendError(w, utils.StatusBadRequest, "Invalid quantity in items! Quantity should be more than 0!")
		return false
	}
	return true
}

func Check_OrderItem(w http.ResponseWriter, r *http.Request, orderItem models.OrderItem) bool {
	if orderItem.Quantity <= 0 {
		utils.SendError(w, utils.StatusBadRequest, "Invalid quantity in items! Quantity should be more than 0!")
		return false
	}
	return true
}
