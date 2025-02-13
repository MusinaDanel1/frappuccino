package dal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"frappuccino/models"
	"time"
)

type OrderRepository struct {
	db *sql.DB
}

type OrderInterface interface {
	Create(order *models.Order) (int, error)
	GetByID(orderID int) (models.Order, error)
	Update(order models.Order, id int) error
	Delete(orderID int) error
	List() ([]models.Order, error)
	GetOrderedItemsCount(startDate, endDate *string) (map[string]int, error)
	CreateOrder(tx *sql.Tx, order models.Order, total float64) (int, error)
	BeginTransaction() (*sql.Tx, error)
}

func NewOrderRepository(db *sql.DB) (*OrderRepository, error) {
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return &OrderRepository{db: db}, nil
}

func (repo *OrderRepository) Create(order *models.Order) (int, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	specialInstructionsJSON, err := json.Marshal(order.SpecialInstructions)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to marshal special instructions: %w", err)
	}

	existsQuery := `
		SELECT order_id FROM orders 
		WHERE customer_name = $1 AND total_amount = $2 
		AND special_instructions = $3::jsonb AND status = $4`

	var existingOrderID int
	err = tx.QueryRow(existsQuery, order.CustomerName, order.TotalAmount, specialInstructionsJSON, order.Status).Scan(&existingOrderID)
	if err == nil {
		tx.Rollback()
		return 0, fmt.Errorf("order already exists with ID %d", existingOrderID)
	} else if err != sql.ErrNoRows {
		tx.Rollback()
		return 0, fmt.Errorf("failed to check existing order: %w", err)
	}

	orderQuery := `
		INSERT INTO orders (customer_name, total_amount, special_instructions, status) 
		VALUES ($1, $2, $3::jsonb, $4) 
		RETURNING order_id, created_at, updated_at`

	var orderID int
	var createdAt, updatedAt time.Time

	err = tx.QueryRow(orderQuery, order.CustomerName, order.TotalAmount, specialInstructionsJSON, order.Status).
		Scan(&orderID, &createdAt, &updatedAt)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to insert order: %w", err)
	}

	orderItemQuery := `
		INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_order) 
		VALUES ($1, $2, $3, 
			(SELECT price FROM menu_items WHERE menu_item_id = $2))`

	for _, item := range order.Items {
		_, err := tx.Exec(orderItemQuery, orderID, item.ProductID, item.Quantity)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to insert order item %d: %w", item.ProductID, err)
		}
	}

	statusHistoryQuery := `
		INSERT INTO order_status_history (order_id, status) 
		VALUES ($1, $2)`

	_, err = tx.Exec(statusHistoryQuery, orderID, order.Status)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to insert order status history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return orderID, nil
}

func (repo *OrderRepository) GetByID(orderID int) (models.Order, error) {
	var order models.Order
	var specialInstructionsJSON []byte

	query := `SELECT order_id, customer_name, total_amount, special_instructions, status, created_at, updated_at 
	          FROM orders WHERE order_id = $1`
	row := repo.db.QueryRow(query, orderID)

	if err := row.Scan(&order.ID, &order.CustomerName, &order.TotalAmount, &specialInstructionsJSON,
		&order.Status, &order.CreatedAt, &order.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Order{}, fmt.Errorf("order with ID %d not found", orderID)
		}
		return models.Order{}, fmt.Errorf("failed to scan order: %w", err)
	}

	if len(specialInstructionsJSON) > 0 {
		if err := json.Unmarshal(specialInstructionsJSON, &order.SpecialInstructions); err != nil {
			return models.Order{}, fmt.Errorf("failed to unmarshal special_instructions: %w", err)
		}
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

func (repo *OrderRepository) Update(order models.Order, id int) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, item := range order.Items {
		if item.Quantity == 0 {
			deleteQuery := `DELETE FROM order_items WHERE order_id = $1 AND menu_item_id = $2`
			_, err := tx.Exec(deleteQuery, id, item.ProductID)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete order item %d: %w", item.ProductID, err)
			}
		} else {
			updateQuery := `
				UPDATE order_items 
				SET quantity = $1, price_at_order = (SELECT price FROM menu_items WHERE menu_item_id = $2)
				WHERE order_id = $3 AND menu_item_id = $2`
			result, err := tx.Exec(updateQuery, item.Quantity, item.ProductID, id)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update order item %d: %w", item.ProductID, err)
			}

			rowsAffected, _ := result.RowsAffected()
			if rowsAffected == 0 {
				insertQuery := `
					INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_order)
					VALUES ($1, $2, $3, (SELECT price FROM menu_items WHERE menu_item_id = $2))`
				_, err := tx.Exec(insertQuery, id, item.ProductID, item.Quantity)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to insert order item %d: %w", item.ProductID, err)
				}
			}
		}
	}

	specialInstructionsJSON, err := json.Marshal(order.SpecialInstructions)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to marshal special_instructions: %w", err)
	}

	orderQuery := `
		UPDATE orders
		SET customer_name = $1, total_amount = $2, special_instructions = $3, status = $4, updated_at = CURRENT_TIMESTAMP
		WHERE order_id = $5`
	_, err = tx.Exec(orderQuery, order.CustomerName, order.TotalAmount, specialInstructionsJSON, order.Status, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order: %w", err)
	}

	if order.Status != "pending" && order.Status != "completed" && order.Status != "canceled" {
		statusHistoryQuery := `INSERT INTO order_status_history (order_id, status) VALUES ($1, $2)`
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

func (repo *OrderRepository) Delete(orderID int) error {
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
		var specialInstructionsJSON []byte

		if err := rows.Scan(&order.ID, &order.CustomerName, &order.TotalAmount, &specialInstructionsJSON,
			&order.Status, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if err := json.Unmarshal(specialInstructionsJSON, &order.SpecialInstructions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal special instructions: %w", err)
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

func (r *OrderRepository) BeginTransaction() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *OrderRepository) CreateOrder(tx *sql.Tx, order models.Order, total float64) (int, error) {
	var orderID int
	query := `
		INSERT INTO orders (customer_name, total_amount, status) 
		VALUES ($1, $2, 'accepted') RETURNING order_id;
	`
	err := tx.QueryRow(query, order.CustomerName, total).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	if len(order.Items) == 0 {
		return 0, fmt.Errorf("order must contain at least one item")
	}

	for _, item := range order.Items {
		_, err := tx.Exec(`
			INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_order) 
			VALUES ($1, $2, $3, (SELECT price FROM menu_items WHERE menu_item_id = $2));
		`, orderID, item.ProductID, item.Quantity)
		if err != nil {
			return 0, fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	return orderID, nil
}
