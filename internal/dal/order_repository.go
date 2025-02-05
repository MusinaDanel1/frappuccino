package dal

import (
	"database/sql"
	"errors"
	"fmt"
	"frappuccino/models"
	"time"

	"github.com/lib/pq"
)

type OrderRepository struct {
	db *sql.DB
}

type OrderInterface interface {
	Create(order models.Order) error
	GetByID(orderID string) (models.Order, error)
	Update(order models.Order, id string) error
	Delete(orderID string) error
	List() ([]models.Order, error)
	GetOrderedItemsCount(startDate, endDate *string) (map[string]int, error)
}

func NewOrderRepository(dsn string) (*OrderRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &OrderRepository{db: db}, nil
}

func (repo *OrderRepository) Create(order models.Order) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	orderQuery := `
		INSERT INTO orders (customer_name, total_amount, special_instructions, status) 
		VALUES ($1, $2, $3, $4) 
		RETURNING order_id, created_at, updated_at`

	var orderID int
	var createdAt, updatedAt time.Time

	err = tx.QueryRow(orderQuery, order.CustomerName, order.TotalAmount, pq.Array(order.SpecialInstructions), order.Status).
		Scan(&orderID, &createdAt, &updatedAt)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert order: %w", err)
	}

	orderItemQuery := `
		INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_order) 
		VALUES ($1, $2, $3, 
			(SELECT price FROM menu_items WHERE menu_item_id = $2))`

	for _, item := range order.Items {
		_, err := tx.Exec(orderItemQuery, orderID, item.ProductID, item.Quantity)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert order item %s: %w", item.ProductID, err)
		}
	}

	statusHistoryQuery := `
		INSERT INTO order_status_history (order_id, status) 
		VALUES ($1, $2)`

	_, err = tx.Exec(statusHistoryQuery, orderID, order.Status)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert order status history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (repo *OrderRepository) GetByID(orderID string) (models.Order, error) {
	var order models.Order
	query := `SELECT order_id, customer_name, total_amount, special_instructions, status, created_at, updated_at 
	          FROM orders WHERE order_id = $1`
	row := repo.db.QueryRow(query, orderID)

	if err := row.Scan(&order.ID, &order.CustomerName, &order.TotalAmount, pq.Array(&order.SpecialInstructions),
		&order.Status, &order.CreatedAt, &order.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Order{}, fmt.Errorf("order with ID %s not found", orderID)
		}
		return models.Order{}, fmt.Errorf("failed to scan order: %w", err)
	}

	orderItemsQuery := `SELECT menu_item_id, quantity FROM order_items WHERE order_id = $1`
	rows, err := repo.db.Query(orderItemsQuery, orderID)
	if err != nil {
		return models.Order{}, fmt.Errorf("failed to fetch order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return models.Order{}, fmt.Errorf("failed to scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	if err := rows.Err(); err != nil {
		return models.Order{}, fmt.Errorf("error iterating over order items: %w", err)
	}

	return order, nil
}

func (repo *OrderRepository) Update(order models.Order, id string) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, item := range order.Items {
		if item.Quantity == 0 {
			deleteQuery := `
				DELETE FROM order_items 
				WHERE order_id = $1 AND menu_item_id = $2`
			_, err := tx.Exec(deleteQuery, id, item.ProductID)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete order item %s: %w", item.ProductID, err)
			}
		} else {
			updateQuery := `
				UPDATE order_items 
				SET quantity = $1, price_at_order = (SELECT price FROM menu_items WHERE menu_item_id = $2)
				WHERE order_id = $3 AND menu_item_id = $2`
			result, err := tx.Exec(updateQuery, item.Quantity, item.ProductID, id)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update order item %s: %w", item.ProductID, err)
			}
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected == 0 {
				insertQuery := `
					INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_order)
					VALUES ($1, $2, $3, (SELECT price FROM menu_items WHERE menu_item_id = $2))`
				_, err := tx.Exec(insertQuery, id, item.ProductID, item.Quantity)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to insert order item %s: %w", item.ProductID, err)
				}
			}
		}
	}

	orderQuery := `
		UPDATE orders
		SET customer_name = $1, total_amount = $2, special_instructions = $3, status = $4, updated_at = CURRENT_TIMESTAMP
		WHERE order_id = $5`
	_, err = tx.Exec(orderQuery, order.CustomerName, order.TotalAmount, pq.Array(order.SpecialInstructions), order.Status, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order: %w", err)
	}

	if order.Status != "pending" && order.Status != "completed" && order.Status != "canceled" {

		statusHistoryQuery := `
			INSERT INTO order_status_history (order_id, status) 
			VALUES ($1, $2)`
		_, err := tx.Exec(statusHistoryQuery, id, order.Status)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert order status history: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (repo *OrderRepository) Delete(orderID string) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	statusHistoryQuery := `
		INSERT INTO order_status_history (order_id, status) 
		VALUES ($1, 'cancelled')`
	_, err = tx.Exec(statusHistoryQuery, orderID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert cancelled status history: %w", err)
	}

	deleteOrderItemsQuery := `
		DELETE FROM order_items 
		WHERE order_id = $1`
	_, err = tx.Exec(deleteOrderItemsQuery, orderID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete order items: %w", err)
	}

	deleteOrderQuery := `
		DELETE FROM orders 
		WHERE order_id = $1`
	result, err := tx.Exec(deleteOrderQuery, orderID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete order: %w", err)
	}

	numRows, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if numRows == 0 {
		tx.Rollback()
		return errors.New("order not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (repo *OrderRepository) List() ([]models.Order, error) {
	query := `SELECT order_id, customer_name, total_amount, special_instructions, status, created_at, updated_at FROM orders`
	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.CustomerName, &order.TotalAmount, pq.Array(&order.SpecialInstructions),
			&order.Status, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		orderItemsQuery := `SELECT menu_item_id, quantity FROM order_items WHERE order_id = $1`
		orderItemsRows, err := repo.db.Query(orderItemsQuery, order.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query order items: %w", err)
		}
		defer orderItemsRows.Close()

		for orderItemsRows.Next() {
			var item models.OrderItem
			if err := orderItemsRows.Scan(&item.ProductID, &item.Quantity); err != nil {
				return nil, fmt.Errorf("failed to scan order item: %w", err)
			}
			order.Items = append(order.Items, item)
		}

		if err := orderItemsRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating over order items: %w", err)
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over orders: %w", err)
	}

	return orders, nil
}

// new endpoint
func (repo *OrderRepository) GetOrderedItemsCount(startDate, endDate *string) (map[string]int, error) {
	query := `
		SELECT mi.name, COALESCE(SUM(oi.quantity), 0) 
		FROM order_items oi
		JOIN menu_items mi ON oi.menu_item_id = mi.menu_item_id
		JOIN orders o ON oi.order_id = o.order_id
		WHERE ($1::timestamptz IS NULL OR o.created_at >= $1::timestamptz)
		  AND ($2::timestamptz IS NULL OR o.created_at <= $2::timestamptz)
		GROUP BY mi.name
	`

	var start, end interface{}
	if startDate != nil && *startDate != "" {
		start = *startDate
	} else {
		start = nil
	}
	if endDate != nil && *endDate != "" {
		end = *endDate
	} else {
		end = nil
	}

	rows, err := repo.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query ordered items count: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var itemName string
		var count int
		if err := rows.Scan(&itemName, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[itemName] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}
	return result, nil
}
