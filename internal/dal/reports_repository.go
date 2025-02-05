package dal

import (
	"database/sql"
	"fmt"
)

type ReportRepository struct {
	db *sql.DB
}

type ReportInterface interface {
	SearchOrdersAndMenuItems(q string, filter string, minPrice, maxPrice *float64) (map[string]interface{}, error)
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

func (repo *ReportRepository) SearchOrdersAndMenuItems(q string, filter string, minPrice, maxPrice *float64) (map[string]interface{}, error) {
	query := `
	WITH search_query AS (
		SELECT plainto_tsquery($1) AS query
	)
	SELECT json_build_object(
		'menu_items', (
			SELECT json_agg(json_build_object(
				'id', mi.menu_item_id,
				'name', mi.name,
				'description', mi.description,
				'price', mi.price,
				'relevance', ts_rank_cd(mi.search_vector, query.query)
			))
			FROM menu_items mi, search_query query
			WHERE ($2 = 'menu' OR $2 = 'all')
			  AND mi.search_vector @@ query.query
			  AND (mi.price >= $3 OR $3 IS NULL)
			  AND (mi.price <= $4 OR $4 IS NULL)
		),
		'orders', (
			SELECT json_agg(json_build_object(
				'id', o.order_id,
				'customer_name', o.customer_name,
				'items', ARRAY(
					SELECT mi.name FROM order_items oi 
					JOIN menu_items mi ON oi.menu_item_id = mi.menu_item_id 
					WHERE oi.order_id = o.order_id
				),
				'total', o.total_amount,
				'relevance', ts_rank_cd(o.search_vector, query.query)
			))
			FROM orders o, search_query query
			WHERE ($2 = 'orders' OR $2 = 'all')
			  AND o.search_vector @@ query.query
			  AND (o.total_amount >= $3 OR $3 IS NULL)
			  AND (o.total_amount <= $4 OR $4 IS NULL)
		),
		'total_matches', (
			SELECT COUNT(*) FROM (
				SELECT 1 FROM menu_items mi, search_query query WHERE ($2 = 'menu' OR $2 = 'all') AND mi.search_vector @@ query.query
				UNION ALL
				SELECT 1 FROM orders o, search_query query WHERE ($2 = 'orders' OR $2 = 'all') AND o.search_vector @@ query.query
			) AS matches
		)
	);
	`

	// Если цены не указаны, передаем NULL
	var minP, maxP interface{}
	if minPrice != nil {
		minP = *minPrice
	} else {
		minP = nil
	}
	if maxPrice != nil {
		maxP = *maxPrice
	} else {
		maxP = nil
	}

	var result map[string]interface{}
	err := repo.db.QueryRow(query, q, filter, minP, maxP).Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}

	return result, nil
}
