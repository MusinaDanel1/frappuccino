package dal

import (
	"context"
	"database/sql"
	"fmt"
	"frappuccino/models"
	"log"
	"strings"

	"github.com/lib/pq"
)

type ReportRepository struct {
	db *sql.DB
}

type ReportInterface interface {
	GetTotalSales(ctx context.Context) (float64, error)
	GetPopularItems(ctx context.Context) ([]string, error)
	FullTextSearch(ctx context.Context, query string, filters []string, minPrice, maxPrice float64) (map[string]interface{}, error)
}

func NewReportRepository(db *sql.DB) (*ReportRepository, error) {
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return &ReportRepository{db: db}, nil
}

func (s *ReportRepository) GetTotalSales(ctx context.Context) (float64, error) {
	query := `
	SELECT o.quantity, mi.price
	FROM orders o
	JOIN menu_items mi ON o.product_id = mi.menu_item_id`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("could not load orders: %w", err)
	}
	defer rows.Close()

	var totalSales float64
	for rows.Next() {
		var quantity int
		var price float64
		if err := rows.Scan(&quantity, &price); err != nil {
			return 0, fmt.Errorf("could not scan order row: %w", err)
		}
		totalSales += float64(quantity) * price
	}

	return totalSales, nil
}

func (s *ReportRepository) GetPopularItems(ctx context.Context) ([]string, error) {
	query := `
	SELECT product_id 
	FROM orders 
	GROUP BY product_id 
	ORDER BY SUM(quantity) DESC 
	LIMIT 1`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("could not load orders: %w", err)
	}
	defer rows.Close()

	var popularItems []string
	for rows.Next() {
		var productID string
		if err := rows.Scan(&productID); err != nil {
			return nil, fmt.Errorf("could not scan order row: %w", err)
		}
		popularItems = append(popularItems, productID)
	}

	return popularItems, nil
}

func (r *ReportRepository) FullTextSearch(ctx context.Context, query string, filters []string, minPrice, maxPrice float64) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	var totalMatches int

	log.Println("query:", query)
	log.Println("filters:", filters)
	log.Println("minPrice:", minPrice)
	log.Println("maxPrice:", maxPrice)

	tsxQuery := fmt.Sprintf("plainto_tsquery('simple', %s)", pq.QuoteLiteral(strings.ReplaceAll(query, " ", " <-> ")))

	if contains(filters, "menu") || contains(filters, "all") {
		menuItems := []map[string]interface{}{}
		menuQuery := `
			SELECT menu_item_id, name, description, price, 
				   ts_rank_cd(to_tsvector('simple', name || ' ' || COALESCE(description, '')), ` + tsxQuery + `) AS relevance
			FROM menu_items
			WHERE to_tsvector('simple', name || ' ' || COALESCE(description, '')) @@ ` + tsxQuery + `
			AND price BETWEEN $1 AND $2
			ORDER BY relevance DESC;
		`
		rows, err := r.db.QueryContext(ctx, menuQuery, minPrice, maxPrice)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var item models.MenuItemR

			if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Relevance); err != nil {
				return nil, err
			}

			menuItemMap := map[string]interface{}{
				"id":          item.ID,
				"name":        item.Name,
				"description": item.Description,
				"price":       item.Price,
				"relevance":   item.Relevance,
			}
			menuItems = append(menuItems, menuItemMap)
		}
		result["menu_items"] = menuItems
		totalMatches += len(menuItems)
	}

	if contains(filters, "orders") || contains(filters, "all") {
		orders := []map[string]interface{}{}
		orderQuery := `
			SELECT subquery.order_id, subquery.customer_name, subquery.total_amount, subquery.items,
			       ts_rank_cd(
				              to_tsvector('simple', subquery.customer_name) || 
							  to_tsvector('simple', COALESCE(array_to_string(subquery.items, ' '), '')), 
			                  plainto_tsquery('simple', $1)) AS relevance
			FROM (
				SELECT o.order_id, o.customer_name, o.total_amount,
				       COALESCE(array_agg(m.name), ARRAY[]::text[]) AS items
				FROM orders o
				LEFT JOIN order_items oi ON o.order_id = oi.order_id
				LEFT JOIN menu_items m ON oi.menu_item_id = m.menu_item_id
				GROUP BY o.order_id, o.customer_name, o.total_amount
			) AS subquery
			WHERE to_tsvector('simple', subquery.customer_name) @@ plainto_tsquery('simple', $1)
			OR
			 to_tsvector('simple',COALESCE(array_to_string(subquery.items, ' '), ''))  @@ plainto_tsquery('simple', $1)
			AND subquery.total_amount BETWEEN $2 AND $3
			ORDER BY relevance DESC;
		`
		rows, err := r.db.QueryContext(ctx, orderQuery, query, minPrice, maxPrice)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var order models.OrderItemR

			if err := rows.Scan(&order.ID, &order.Customer, pq.Array(&order.Items), &order.Total, &order.Relevance); err != nil {
				return nil, err
			}

			orderMap := map[string]interface{}{
				"id":            order.ID,
				"customer_name": order.Customer,
				"items":         order.Items,
				"total":         order.Total,
				"relevance":     order.Relevance,
			}

			orders = append(orders, orderMap)
		}
		result["orders"] = orders
		totalMatches += len(orders)
	}

	if contains(filters, "inventory") || contains(filters, "all") {
		inventoryItems := []map[string]interface{}{}

		inventoryQuery := `
			SELECT ingredient_id, name, quantity, 
            ts_rank_cd(to_tsvector('simple', COALESCE(name, '')), plainto_tsquery('simple', $1)) AS relevance
            FROM inventory
            WHERE to_tsvector('simple', COALESCE(name, '')) @@ plainto_tsquery('simple', $1)
            ORDER BY relevance DESC;
		`
		rows, err := r.db.QueryContext(ctx, inventoryQuery, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var item models.InventoryItemR

			if err := rows.Scan(&item.ID, &item.Name, &item.Quantity, &item.Relevance); err != nil {
				return nil, err
			}

			inventoryMap := map[string]interface{}{
				"id":        item.ID,
				"name":      item.Name,
				"quantity":  item.Quantity,
				"relevance": item.Relevance,
			}

			inventoryItems = append(inventoryItems, inventoryMap)
		}
		result["inventory"] = inventoryItems
		totalMatches += len(inventoryItems)
	}
	result["total_matches"] = totalMatches

	return result, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (r *ReportRepository) GetOrderedItems(ctx context.Context) ([]models.Order, error) {
	query := `
	SELECT order_id, customer_name, total_amount, status, created_at
	FROM orders`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("could not fetch ordered items: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.CustomerName, &order.TotalAmount, &order.Status, &order.CreatedAt); err != nil {
			return nil, fmt.Errorf("could not scan order row: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}
