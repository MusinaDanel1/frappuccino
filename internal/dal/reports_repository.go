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
	FullTextSearch(ctx context.Context, query string, filters []string, minPrice, maxPrice float64) (*models.SearchResponse, error)
}

func NewReportsRepository(dsn string) (*ReportRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &ReportRepository{db: db}, nil
}

func (r *ReportRepository) FullTextSearch(ctx context.Context, query string, filters []string, minPrice, maxPrice float64) (*models.SearchResponse, error) {
	var menuItems []models.SearchResult
	var orders []models.SearchResult

	log.Println("query:", query)
	log.Println("filters:", filters)
	log.Println("minPrice:", minPrice)
	log.Println("maxPrice:", maxPrice)

	// Преобразуем запрос в phraseto_tsquery
	tsxQuery := fmt.Sprintf("plainto_tsquery('simple', %s)", pq.QuoteLiteral(strings.ReplaceAll(query, " ", " <-> ")))

	if contains(filters, "menu") || contains(filters, "all") {
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
			var item models.SearchResult
			if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Relevance); err != nil {
				return nil, err
			}
			menuItems = append(menuItems, item)
		}
	}

	if contains(filters, "orders") || contains(filters, "all") {
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
			var order models.SearchResult
			var items []string
			if err := rows.Scan(&order.ID, &order.Name, &order.Total, pq.Array(&items), &order.Relevance); err != nil {
				return nil, err
			}
			order.Items = items
			orders = append(orders, order)
		}
	}

	return &models.SearchResponse{
		MenuItems:    menuItems,
		Orders:       orders,
		TotalMatches: len(menuItems) + len(orders),
	}, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
