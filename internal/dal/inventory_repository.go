package dal

import (
	"database/sql"
	"errors"
	"fmt"
	"frappuccino/models"
)

type InventoryRepository struct {
	db *sql.DB
}

type InventoryInterface interface {
	Create(ingredient models.InventoryItem) (int, error)
	GetByID(ingID int) (models.InventoryItem, error)
	Update(ingredient models.InventoryItem, id int) error
	Delete(ingID int) error
	List() ([]models.InventoryItem, error)
	CheckAndReserveInventory(tx *sql.Tx, items []models.OrderItem) (float64, bool, []models.InventoryUpdate, error)
}

func NewInventoryRepository(db *sql.DB) (*InventoryRepository, error) {
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return &InventoryRepository{db: db}, nil
}

func (repo *InventoryRepository) Create(ingredient models.InventoryItem) (int, error) {
	var exists bool
	queryCheck := `SELECT EXISTS(SELECT 1 FROM inventory WHERE name = $1)`
	err := repo.db.QueryRow(queryCheck, ingredient.Name).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("failed to check ingredient existence: %w", err)
	}
	if exists {
		return 0, errors.New("ingredient with this name already exists")
	}

	var id int
	queryInsert := `INSERT INTO inventory (name, quantity, unit) VALUES ($1, $2, $3) RETURNING ingredient_id`
	err = repo.db.QueryRow(queryInsert, ingredient.Name, ingredient.Quantity, ingredient.Unit).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create ingredient: %w", err)
	}

	return id, nil
}

func (repo *InventoryRepository) GetByID(ingID int) (models.InventoryItem, error) {
	var ingredient models.InventoryItem
	query := `SELECT ingredient_id, name, quantity, unit FROM inventory WHERE ingredient_id = $1`
	row := repo.db.QueryRow(query, ingID)
	if err := row.Scan(&ingredient.IngredientID, &ingredient.Name, &ingredient.Quantity, &ingredient.Unit); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.InventoryItem{}, errors.New("ingredient not found")
		}
		return models.InventoryItem{}, err
	}
	return ingredient, nil
}

func (repo *InventoryRepository) Update(ingredient models.InventoryItem, id int) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var oldQuantity float64
	queryGet := `SELECT quantity FROM inventory WHERE ingredient_id = $1`
	err = tx.QueryRow(queryGet, id).Scan(&oldQuantity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("ingredient not found")
		}
		return fmt.Errorf("failed to get current quantity: %w", err)
	}

	queryUpdate := `UPDATE inventory SET name = $1, quantity = $2, unit = $3 WHERE ingredient_id = $4`
	result, err := tx.Exec(queryUpdate, ingredient.Name, ingredient.Quantity, ingredient.Unit, id)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	numRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if numRows == 0 {
		return errors.New("ingredient not found")
	}

	quantityChange := ingredient.Quantity - oldQuantity
	transactionType := "addition"
	if quantityChange < 0 {
		transactionType = "deduction"
	}

	queryInsert := `INSERT INTO inventory_transactions (ingredient_id, quantity_change, transaction_type) VALUES ($1, $2, $3)`
	_, err = tx.Exec(queryInsert, id, quantityChange, transactionType)
	if err != nil {
		return fmt.Errorf("failed to insert transaction record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (repo *InventoryRepository) Delete(ingID int) error {
	query := `DELETE FROM inventory WHERE ingredient_id = $1`
	result, err := repo.db.Exec(query, ingID)
	if err != nil {
		return err
	}

	numRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if numRows == 0 {
		return errors.New("ingredient not found")
	}
	return nil
}

func (repo *InventoryRepository) List() ([]models.InventoryItem, error) {
	query := `SELECT ingredient_id, name, quantity, unit FROM inventory`
	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredients []models.InventoryItem
	for rows.Next() {
		var ingredient models.InventoryItem
		if err := rows.Scan(&ingredient.IngredientID, &ingredient.Name, &ingredient.Quantity, &ingredient.Unit); err != nil {
			return nil, err
		}
		ingredients = append(ingredients, ingredient)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ingredients, nil
}

func (repo *InventoryRepository) Close() error {
	return repo.db.Close()
}

func (r *InventoryRepository) CheckAndReserveInventory(tx *sql.Tx, items []models.OrderItem) (float64, bool, []models.InventoryUpdate, error) {
	var total float64
	inventoryMap := make(map[int]*models.InventoryUpdate) // Map для агрегации ингредиентов

	for _, item := range items {
		if item.ProductID == 0 {
			return 0, false, nil, fmt.Errorf("ProductID is empty")
		}

		var availableQuantity, ingredientQuantity int
		var price float64
		var ingredientID int
		var name string

		err := tx.QueryRow(`
			SELECT i.ingredient_id, i.name, i.quantity, mii.quantity AS ingredient_quantity, mi.price
			FROM inventory i
			JOIN menu_item_ingredients mii ON i.ingredient_id = mii.inventory_id
			JOIN menu_items mi ON mi.menu_item_id = mii.menu_item_id
			WHERE mi.menu_item_id = $1;
		`, item.ProductID).Scan(&ingredientID, &name, &availableQuantity, &ingredientQuantity, &price)

		if err == sql.ErrNoRows {
			return 0, false, nil, fmt.Errorf("no inventory found for ProductID: %d", item.ProductID)
		} else if err != nil {
			return 0, false, nil, fmt.Errorf("error fetching inventory: %w", err)
		}

		requiredQuantity := item.Quantity * ingredientQuantity
		if availableQuantity < requiredQuantity {
			return 0, false, nil, nil // Недостаточно ингредиентов
		}

		_, err = tx.Exec(`
			UPDATE inventory 
			SET quantity = quantity - $1
			WHERE ingredient_id = $2;
		`, requiredQuantity, ingredientID)

		if err != nil {
			return 0, false, nil, fmt.Errorf("error updating inventory: %w", err)
		}

		// Агрегируем данные по ingredient_id
		if update, exists := inventoryMap[ingredientID]; exists {
			update.QuantityUsed += requiredQuantity // Увеличиваем использованное количество
		} else {
			inventoryMap[ingredientID] = &models.InventoryUpdate{
				IngredientID: ingredientID,
				Name:         name,
				QuantityUsed: requiredQuantity,
				Remaining:    availableQuantity - requiredQuantity, // Фиксируем остаток при первой встрече ингредиента
			}
		}

		total += price * float64(item.Quantity)
	}

	// Конвертируем map в slice
	var inventoryUpdates []models.InventoryUpdate
	for _, update := range inventoryMap {
		inventoryUpdates = append(inventoryUpdates, *update)
	}

	return total, true, inventoryUpdates, nil
}
