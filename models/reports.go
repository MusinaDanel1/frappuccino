package models

type MenuItemR struct {
	ID          int
	Name        string
	Description string
	Price       float64
	Relevance   float64
}

type OrderItemR struct {
	ID        int
	Customer  string
	Items     []string
	Total     float64
	Relevance float64
}

type InventoryItemR struct {
	ID        int
	Name      string
	Quantity  int
	Relevance float64
}
