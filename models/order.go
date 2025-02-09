package models

type Order struct {
	ID                  string      `json:"order_id"`
	CustomerName        string      `json:"customer_name"`
	TotalAmount         float64     `json:"total_amount"`
	Items               []OrderItem `json:"items"`
	SpecialInstructions []string    `json:"special_instructions"`
	Status              string      `json:"status"`
	CreatedAt           string      `json:"created_at"`
	UpdatedAt           string      `json:"update_at"`
}

type OrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type ChangeHistory struct {
	Timestamp string `json:"timestamp"`
	EventType string `json:"event_type"`
	OldValue  string `json:"old_value"`
	NewValue  string `json:"new_value"`
}
