package models

type SearchResult struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Price       float64  `json:"price"`
	Items       []string `json:"items"`
	Total       float64  `json:"total"`
	Relevance   float64  `json:"relevance"`
}

type SearchResponse struct {
	MenuItems    []SearchResult `json:"menu_items"`
	Orders       []SearchResult `json:"orders"`
	TotalMatches int            `json:"total matches"`
}
