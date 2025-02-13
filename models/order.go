package models

import "time"

type Order struct {
	ID                  int         `json:"order_id"`
	CustomerName        string      `json:"customer_name"`
	TotalAmount         float64     `json:"total_amount"`
	Items               []OrderItem `json:"items"`
	SpecialInstructions []string    `json:"special_instructions"`
	Status              string      `json:"status"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           string      `json:"update_at"`
}

type OrderItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type ChangeHistory struct {
	Timestamp string `json:"timestamp"`
	EventType string `json:"event_type"`
	OldValue  string `json:"old_value"`
	NewValue  string `json:"new_value"`
}
