package dal

import (
	"database/sql"
	"errors"
	"fmt"
	"frappuccino/models"
	"strconv"

	_ "github.com/lib/pq"
)

type InventoryRepository struct {
	db *sql.DB
}

type InventoryInterface interface {
	Create(ingredient models.InventoryItem) error
	GetByID(ingID string) (models.InventoryItem, error)
	Update(ingredient models.InventoryItem, id string) error
	Delete(ingID string) error
	List() ([]models.InventoryItem, error)
	CheckAndReserveInventory(tx *sql.Tx, items []models.OrderItem) (float64, bool, map[string]int)
}

func NewInventoryRepository(dsn string) (*InventoryRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &InventoryRepository{db: db}, nil
}

func (repo *InventoryRepository) Create(ingredient models.InventoryItem) error {
	query := `INSERT INTO inventory (name, quantity, unit) VALUES ($1, $2, $3)`
	_, err := repo.db.Exec(query, ingredient.Name, ingredient.Quantity, ingredient.Unit)
	if err != nil {
		return err
	}
	return nil
}

func (repo *InventoryRepository) GetByID(ingID string) (models.InventoryItem, error) {
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

func (repo *InventoryRepository) Update(ingredient models.InventoryItem, id string) error {
	query := `UPDATE inventory SET ingredient_id = $1, name = $2, quantity = $3, unit = $4 WHERE ingredient_id = $5`
	result, err := repo.db.Exec(query, ingredient.IngredientID, ingredient.Name, ingredient.Quantity, ingredient.Unit, id)
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

func (repo *InventoryRepository) Delete(ingID string) error {
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

func (r *InventoryRepository) CheckAndReserveInventory(tx *sql.Tx, items []models.OrderItem) (float64, bool, map[string]int) {
	var total float64
	inventoryUpdates := make(map[string]int) // Добавляем мапу для хранения обновлений

	for _, item := range items {
		if item.ProductID == "" {
			fmt.Println("Error: ProductID is empty")
			return 0, false, nil
		}

		menuItemID, err := strconv.Atoi(item.ProductID)
		if err != nil {
			fmt.Println("Invalid ProductID:", item.ProductID)
			return 0, false, nil
		}

		var availableQuantity, ingredientQuantity int
		var price float64

		err = tx.QueryRow(`
			SELECT i.quantity, mii.quantity AS ingredient_quantity, mi.price
			FROM inventory i
			JOIN menu_item_ingredients mii ON i.ingredient_id = mii.inventory_id
			JOIN menu_items mi ON mi.menu_item_id = mii.menu_item_id
			WHERE mi.menu_item_id = $1;
		`, menuItemID).Scan(&availableQuantity, &ingredientQuantity, &price)

		if err == sql.ErrNoRows {
			fmt.Println("No inventory found for ProductID:", menuItemID)
			return 0, false, nil
		} else if err != nil {
			fmt.Println("Error fetching inventory:", err)
			return 0, false, nil
		}

		fmt.Printf("ProductID: %d, Available: %d, Needed: %d\n", menuItemID, availableQuantity, item.Quantity*ingredientQuantity)

		if availableQuantity < item.Quantity*ingredientQuantity {
			fmt.Println("Not enough inventory")
			return 0, false, nil
		}

		// Обновляем инвентарь
		_, err = tx.Exec(`
			UPDATE inventory 
			SET quantity = inventory.quantity - (mii.quantity * $1)
			FROM menu_item_ingredients mii
			WHERE inventory.ingredient_id = mii.inventory_id
			AND mii.menu_item_id = $2;
		`, item.Quantity, menuItemID)

		if err != nil {
			fmt.Println("Error updating inventory:", err)
			return 0, false, nil
		}

		// Добавляем обновленный ингредиент в `inventory_updates`
		inventoryUpdates[item.ProductID] += item.Quantity * ingredientQuantity

		total += price * float64(item.Quantity)
	}

	return total, true, inventoryUpdates
}
