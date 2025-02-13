package check

import (
	"frappuccino/internal/utils"
	"frappuccino/models"
	"net/http"
	"time"
)

func Check_Orders(w http.ResponseWriter, r *http.Request, orders models.Order) bool {
	if orders.CustomerName == "" {
		utils.SendError(w, utils.StatusBadRequest, "Empty Customer name in orders!")
		return false
	}
	if orders.TotalAmount < 0 {
		utils.SendError(w, utils.StatusBadRequest, "Invalid total amount in orders! Total amount should be more than 0!")
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

func Check_Date(w http.ResponseWriter, r *http.Request, startDate, endDate string) (*string, *string, bool) {
	const dateFormat = "2006-01-02" // Формат YYYY-MM-DD

	var startDatePtr, endDatePtr *string
	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse(dateFormat, startDate)
		if err != nil {
			utils.SendError(w, utils.StatusBadRequest, "Invalid 'startDate' format. Use 'YYYY-MM-DD'.")
			return nil, nil, false
		}
		startDatePtr = &startDate
	}

	if endDate != "" {
		end, err = time.Parse(dateFormat, endDate)
		if err != nil {
			utils.SendError(w, utils.StatusBadRequest, "Invalid 'endDate' format. Use 'YYYY-MM-DD'.")
			return nil, nil, false
		}
		endDatePtr = &endDate
	}

	if startDatePtr != nil && endDatePtr != nil && start.After(end) {
		utils.SendError(w, utils.StatusBadRequest, "'startDate' cannot be later than 'endDate'.")
		return nil, nil, false
	}

	return startDatePtr, endDatePtr, true
}

func Check_OrderItemheckFilters(w http.ResponseWriter, r *http.Request, filters []string) bool {
	validFilters := map[string]bool{"menu": true, "order": true, "all": true}

	for _, f := range filters {
		if !validFilters[f] {
			utils.SendError(w, utils.StatusBadRequest, "Invalid filter: "+f+". Allowed values: menu, order, all.")
			return false
		}
	}

	return true
}
