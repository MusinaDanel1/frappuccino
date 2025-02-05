package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"frappuccino/internal/dal"
	"frappuccino/models"
	"math/rand"
	"os"
	"strconv"
	"time"
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

func (s *OrderService) CreateOrder(order models.Order) error {
	requiredIngredients := make(map[string]float64)

	orderID := strconv.FormatInt(time.Now().UnixNano(), 10) + strconv.Itoa(rand.Intn(1000))
	order.ID = orderID
	for _, item := range order.Items {
		menuItem, err := s.menuRepo.GetByID(item.ProductID)
		if err != nil {
			return errors.New("menu item not found: " + item.ProductID)
		}

		for _, ingredient := range menuItem.Ingredients {
			totalQuantity := ingredient.Quantity * float64(item.Quantity)
			requiredIngredients[ingredient.IngredientID] += totalQuantity
		}
	}

	for ingredientID, requiredQuantity := range requiredIngredients {
		inventoryItem, err := s.inventoryRepo.GetByID(ingredientID)
		if err != nil {
			return errors.New("ingredient not found: " + ingredientID)
		}
		if inventoryItem.Quantity < requiredQuantity {
			return errors.New("insufficient ingredient: " + inventoryItem.Name)
		}
	}

	for ingredientID, requiredQuantity := range requiredIngredients {
		inventoryItem, _ := s.inventoryRepo.GetByID(ingredientID)
		inventoryItem.Quantity -= requiredQuantity
		if err := s.inventoryRepo.Update(inventoryItem, ingredientID); err != nil {
			return errors.New("failed to update ingredient quantity: " + inventoryItem.Name)
		}
	}

	order.CreatedAt = time.Now().Format(time.RFC3339)
	if err := s.orderRepo.Create(order); err != nil {
		return errors.New("failed to create order")
	}
	return nil
}

func (s *OrderService) GetByID(orderID string) (models.Order, error) {
	return s.orderRepo.GetByID(orderID)
}

func (s *OrderService) Update(order models.Order, id string) error {
	return s.orderRepo.Update(order, id)
}

func (s *OrderService) Delete(orderID string) error {
	return s.orderRepo.Delete(orderID)
}

func (s *OrderService) List() ([]models.Order, error) {
	return s.orderRepo.List()
}

func (s *OrderService) GetOrderedItemsCount(startDate, endDate *string) (map[string]int, error) {
	return s.orderRepo.GetOrderedItemsCount(startDate, endDate)
}

func (s *OrderService) RecordChangeHistory(orderID string, changes []models.ChangeHistory) error {
	if _, err := os.Stat("/order_history.json"); os.IsExist(err) {
		os.RemoveAll("/order_history.json")
	}
	_, err := os.Create("/order_history.json")
	if err != nil {
		fmt.Printf("Failed fo create a file %s: %v\n", "/order_history.json", err)
	}
	file, err := os.OpenFile("/order_history.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open history file: %v", err)
	}
	defer file.Close()

	for _, change := range changes {
		change.Timestamp = time.Now().Format(time.RFC3339)
		change.EventType = fmt.Sprintf("%s_changed", orderID)
		encoder := json.NewEncoder(file)
		if err := encoder.Encode(change); err != nil {
			return fmt.Errorf("failed to write history change: %v", err)
		}
	}

	return nil
}

func (s *OrderService) CollectOrderChanges(existingOrder, newOrder models.Order) []models.ChangeHistory {
	var changeHistory []models.ChangeHistory

	if existingOrder.CustomerName != newOrder.CustomerName {
		changeHistory = append(changeHistory, models.ChangeHistory{
			OldValue: existingOrder.CustomerName,
			NewValue: newOrder.CustomerName,
		})
	}

	if !areOrdersEqual(existingOrder.Items, newOrder.Items) {
		changeHistory = append(changeHistory, models.ChangeHistory{
			OldValue: fmt.Sprintf("%v", existingOrder.Items),
			NewValue: fmt.Sprintf("%v", newOrder.Items),
		})
	}

	return changeHistory
}

func areOrdersEqual(oldItems, newItems []models.OrderItem) bool {
	if len(oldItems) != len(newItems) {
		return false
	}

	for i := range oldItems {
		if oldItems[i].ProductID != newItems[i].ProductID || oldItems[i].Quantity != newItems[i].Quantity {
			return false
		}
	}

	return true
}
