package models

type SearchResult struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Quantity    *int     `json:"quantity,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Items       []string `json:"items"`
	Total       float64  `json:"total"`
	Relevance   float64  `json:"relevance"`
}

type SearchResponse struct {
	MenuItems     []SearchResult `json:"menu_items"`
	Orders        []SearchResult `json:"orders"`
	InventoryItem []SearchResult `json:"inventory"`
	TotalMatches  int            `json:"total matches"`
}
