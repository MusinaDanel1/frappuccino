package models

type InventoryItem struct {
	IngredientID int     `json:"ingredient_id"`
	Name         string  `json:"name"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	Price        float64 `json:"price"`
}

type InventoryUpdate struct {
	IngredientID int
	Name         string
	QuantityUsed int
	Remaining    int
}
