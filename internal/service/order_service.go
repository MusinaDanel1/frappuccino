package service

import (
	"fmt"
	"frappuccino/internal/dal"
	"frappuccino/models"
)

type OrderService struct {
	menuRepo      dal.MenuInterface
	orderRepo     dal.OrderInterface
	inventoryRepo dal.InventoryInterface
}

func NewOrderService(orderRepo dal.OrderInterface, inventoryRepo dal.InventoryInterface, menuRepo dal.MenuInterface) *OrderService {
	return &OrderService{
		menuRepo:      menuRepo,
		orderRepo:     orderRepo,
		inventoryRepo: inventoryRepo,
	}
}

func (s *OrderService) CreateOrder(order *models.Order) error {
	requiredIngredients := make(map[int]float64)

	for _, item := range order.Items {
		menuItem, err := s.menuRepo.GetByID(item.ProductID)
		if err != nil {
			return fmt.Errorf("menu item not found: %d", item.ProductID)
		}

		for _, ingredient := range menuItem.Ingredients {
			totalQuantity := ingredient.Quantity * float64(item.Quantity)
			requiredIngredients[ingredient.IngredientID] += totalQuantity
		}
	}

	for ingredientID, requiredQuantity := range requiredIngredients {
		inventoryItem, err := s.inventoryRepo.GetByID(ingredientID)
		if err != nil {
			return fmt.Errorf("ingredient not found: %d", ingredientID)
		}
		if inventoryItem.Quantity < requiredQuantity {
			return fmt.Errorf("insufficient ingredient: %s", inventoryItem.Name)
		}
	}

	orderID, err := s.orderRepo.Create(order)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	order.ID = orderID

	for ingredientID, requiredQuantity := range requiredIngredients {
		inventoryItem, _ := s.inventoryRepo.GetByID(ingredientID)
		inventoryItem.Quantity -= requiredQuantity
		if err := s.inventoryRepo.Update(inventoryItem, ingredientID); err != nil {
			return fmt.Errorf("failed to update ingredient quantity: %s", inventoryItem.Name)
		}
	}

	return nil
}

func (s *OrderService) GetByID(orderID int) (models.Order, error) {
	return s.orderRepo.GetByID(orderID)
}

func (s *OrderService) Update(order models.Order, id int) error {
	return s.orderRepo.Update(order, id)
}

func (s *OrderService) Delete(orderID int) error {
	return s.orderRepo.Delete(orderID)
}

func (s *OrderService) List() ([]models.Order, error) {
	return s.orderRepo.List()
}

func (s *OrderService) GetOrderedItemsCount(startDate, endDate *string) (map[string]int, error) {
	return s.orderRepo.GetOrderedItemsCount(startDate, endDate)
}

func (s *OrderService) ProcessBulkOrders(orders []models.Order) (map[string]interface{}, error) {
	var processedOrders []map[string]interface{}
	var totalRevenue float64
	accepted, rejected := 0, 0
	var inventoryUpdates []map[string]interface{}

	tx, err := s.orderRepo.BeginTransaction()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for _, order := range orders {
		total, sufficient, updates, err := s.inventoryRepo.CheckAndReserveInventory(tx, order.Items)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to check inventory: %w", err)
		}

		if !sufficient {
			processedOrders = append(processedOrders, map[string]interface{}{
				"customer_name": order.CustomerName,
				"status":        "rejected",
				"reason":        "insufficient_inventory",
			})
			rejected++
			continue
		}

		orderID, err := s.orderRepo.CreateOrder(tx, order, total)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create order: %w", err)
		}

		processedOrders = append(processedOrders, map[string]interface{}{
			"order_id":      orderID,
			"customer_name": order.CustomerName,
			"status":        "accepted",
			"total":         total,
		})
		totalRevenue += total
		accepted++

		// Формируем список обновлений инвентаря
		for _, update := range updates {
			inventoryUpdates = append(inventoryUpdates, map[string]interface{}{
				"ingredient_id": update.IngredientID,
				"name":          update.Name,
				"quantity_used": update.QuantityUsed,
				"remaining":     update.Remaining,
			})
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return map[string]interface{}{
		"processed_orders": processedOrders,
		"summary": map[string]interface{}{
			"total_orders":      len(orders),
			"accepted":          accepted,
			"rejected":          rejected,
			"total_revenue":     totalRevenue,
			"inventory_updates": inventoryUpdates,
		},
	}, nil
}
