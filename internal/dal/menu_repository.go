package dal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"frappuccino/models"
	"strconv"

	"github.com/lib/pq"
)

type MenuRepository struct {
	db *sql.DB
}

type MenuInterface interface {
	Create(menuItem models.MenuItem) error
	GetByID(ingID string) (models.MenuItem, error)
	Update(item models.MenuItem, id string) error
	Delete(menuID string) error
	List() ([]models.MenuItem, error)
}

func NewMenuRepository(dsn string) (*MenuRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &MenuRepository{db: db}, nil
}

func (repo *MenuRepository) Create(menuItem models.MenuItem) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}

	query := `INSERT INTO menu_items (name, description, price, categories, allergens) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING menu_item_id`

	var menuItemID int
	err = tx.QueryRow(query, menuItem.Name, menuItem.Description, menuItem.Price,
		pq.Array(menuItem.Category), pq.Array(menuItem.Allergens)).Scan(&menuItemID)
	if err != nil {
		tx.Rollback()
		return err
	}

	ingredientQuery := `INSERT INTO menu_item_ingredients (menu_item_id, inventory_id, quantity) VALUES ($1, $2, $3)`
	for _, ingredient := range menuItem.Ingredients {
		ingredientID, err := strconv.Atoi(ingredient.IngredientID)
		if err != nil {
			tx.Rollback()
			return err
		}

		_, err = tx.Exec(ingredientQuery, menuItemID, ingredientID, ingredient.Quantity)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (repo *MenuRepository) GetByID(menuID string) (models.MenuItem, error) {
	idInt, err := strconv.Atoi(menuID)
	if err != nil {
		return models.MenuItem{}, fmt.Errorf("invalid menu ID: %w", err)
	}

	query := `
	SELECT 
	    mi.menu_item_id, 
	    mi.name, 
	    COALESCE(mi.description, '') AS description, 
	    mi.price, 
	    COALESCE(
	        jsonb_agg(
	            jsonb_build_object('ingredient_id', mii.inventory_id::text, 'quantity', mii.quantity)
	        ) FILTER (WHERE mii.inventory_id IS NOT NULL), '[]'::jsonb
	    ) AS ingredients,
	    COALESCE(mi.categories, ARRAY[]::TEXT[]) AS categories, 
	    COALESCE(mi.allergens, ARRAY[]::TEXT[]) AS allergens
	FROM 
	    menu_items mi
	LEFT JOIN 
	    menu_item_ingredients mii 
	    ON mi.menu_item_id = mii.menu_item_id
	WHERE 
	    mi.menu_item_id = $1
	GROUP BY 
	    mi.menu_item_id;
	`

	var (
		id              int
		name            string
		description     string
		price           float64
		ingredientsJSON string
		categories      []string
		allergens       []string
	)

	row := repo.db.QueryRow(query, idInt)
	if err := row.Scan(&id, &name, &description, &price, &ingredientsJSON, pq.Array(&categories), pq.Array(&allergens)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.MenuItem{}, errors.New("menu item not found")
		}
		return models.MenuItem{}, fmt.Errorf("failed to scan row: %w", err)
	}

	var ingredients []models.MenuItemIngredient
	if err := json.Unmarshal([]byte(ingredientsJSON), &ingredients); err != nil {
		return models.MenuItem{}, fmt.Errorf("failed to unmarshal ingredients: %w", err)
	}

	item := models.MenuItem{
		ID:          strconv.Itoa(id),
		Name:        name,
		Description: description,
		Price:       price,
		Ingredients: ingredients,
		Category:    categories,
		Allergens:   allergens,
	}

	return item, nil
}

func (repo *MenuRepository) Update(item models.MenuItem, id string) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin update transaction: %w", err)
	}
	defer tx.Rollback()

	updateMenuQuery := `
		UPDATE menu_items 
		SET 
			name = $1, 
			price = $2,
			description = $3, 
			categories = $4,
			allergens = $5 
		WHERE 
			menu_item_id = $6
	`
	_, err = tx.Exec(updateMenuQuery, item.Name, item.Price, item.Description, pq.Array(item.Category), pq.Array(item.Allergens), id)
	if err != nil {
		return fmt.Errorf("failed to update menu item: %w", err)
	}

	deleteIngredientsQuery := `DELETE FROM menu_item_ingredients WHERE menu_item_id = $1`
	_, err = tx.Exec(deleteIngredientsQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete old ingredients: %w", err)
	}

	for _, ingredient := range item.Ingredients {
		if ingredient.Quantity == 0 {
			fmt.Println("Warning: Quantity is zero for ingredient:", ingredient.IngredientID)
		}
	}

	insertIngredientsQuery := `
		INSERT INTO menu_item_ingredients (menu_item_id, inventory_id, quantity)
		VALUES ($1, $2, $3)
	`
	for _, ingredient := range item.Ingredients {
		_, err := tx.Exec(insertIngredientsQuery, id, ingredient.IngredientID, ingredient.Quantity)
		if err != nil {
			return fmt.Errorf("failed to insert ingredient: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (repo *MenuRepository) Delete(menuID string) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	ingredientQuery := `DELETE FROM menu_item_ingredients WHERE menu_item_id = $1`
	_, err = tx.Exec(ingredientQuery, menuID)
	if err != nil {
		return fmt.Errorf("failed to delete ingredients: %w", err)
	}

	menuQuery := `DELETE FROM menu_items WHERE menu_item_id = $1`
	result, err := tx.Exec(menuQuery, menuID)
	if err != nil {
		return fmt.Errorf("failed to delete menu item: %w", err)
	}

	numRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if numRows == 0 {
		return errors.New("menu item not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (repo *MenuRepository) List() ([]models.MenuItem, error) {
	query := `
	SELECT 
	    mi.menu_item_id, 
	    mi.name, 
	    COALESCE(mi.description, '') AS description, 
	    mi.price, 
	    COALESCE(
	        jsonb_agg(
	            jsonb_build_object('ingredient_id', mii.inventory_id::text, 'quantity', mii.quantity)
	        ) FILTER (WHERE mii.inventory_id IS NOT NULL), '[]'::jsonb
	    ) AS ingredients,
	    COALESCE(mi.categories, ARRAY[]::TEXT[]) AS categories, 
	    COALESCE(mi.allergens, ARRAY[]::TEXT[]) AS allergens
	FROM 
	    menu_items mi
	LEFT JOIN 
	    menu_item_ingredients mii 
	    ON mi.menu_item_id = mii.menu_item_id
	GROUP BY 
	    mi.menu_item_id;
	`

	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query menu items: %w", err)
	}
	defer rows.Close()

	var menuItems []models.MenuItem
	for rows.Next() {
		var (
			id              int
			name            string
			description     string
			price           float64
			ingredientsJSON string
			categories      []string
			allergens       []string
		)

		if err := rows.Scan(&id, &name, &description, &price, &ingredientsJSON, pq.Array(&categories), pq.Array(&allergens)); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var ingredients []models.MenuItemIngredient
		if err := json.Unmarshal([]byte(ingredientsJSON), &ingredients); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ingredients: %w", err)
		}

		menuItems = append(menuItems, models.MenuItem{
			ID:          strconv.Itoa(id),
			Name:        name,
			Description: description,
			Price:       price,
			Ingredients: ingredients,
			Category:    categories,
			Allergens:   allergens,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return menuItems, nil
}
