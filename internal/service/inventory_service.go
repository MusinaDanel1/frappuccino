package service

import (
	"errors"
	"fmt"
	"frappuccino/internal/dal"
	"frappuccino/models"
	"log"
	"sort"
)

type InventoryService struct {
	repo dal.InventoryInterface
}

func NewIngredientService(repo dal.InventoryInterface) *InventoryService {
	return &InventoryService{
		repo: repo,
	}
}

func (s *InventoryService) Create(ingredient *models.InventoryItem) error {
	id, err := s.repo.Create(*ingredient)
	if err != nil {
		log.Default().Printf("error creating ingredient: %v\n", err)
		return errors.New("failed to create ingredient")
	}

	ingredient.IngredientID = id // Присваиваем полученный ID

	return nil
}

func (s *InventoryService) GetByID(ingID int) (models.InventoryItem, error) {
	return s.repo.GetByID(ingID)
}

func (s *InventoryService) Update(ingredient models.InventoryItem, id int) error {
	return s.repo.Update(ingredient, id)
}

func (s *InventoryService) Delete(ingID int) error {
	return s.repo.Delete(ingID)
}

func (s *InventoryService) List() ([]models.InventoryItem, error) {
	return s.repo.List()
}

func (s *InventoryService) GetLeftOvers(sortBy string, page, pageSize int) ([]models.InventoryItem, int, error) {
	// Получаем весь инвентарь
	allItems, err := s.repo.List()
	if err != nil {
		return nil, 0, err
	}

	// Фильтруем ингредиенты, у которых количество больше 0
	var leftovers []models.InventoryItem
	for _, item := range allItems {
		if item.Quantity > 0 {
			leftovers = append(leftovers, item)
		}
	}

	// Сортировка
	switch sortBy {
	case "price":
		sort.Slice(leftovers, func(i, j int) bool {
			return leftovers[i].Price < leftovers[j].Price // Предполагается, что у InventoryItem есть поле Price
		})
	case "quantity":
		sort.Slice(leftovers, func(i, j int) bool {
			return leftovers[i].Quantity < leftovers[j].Quantity
		})
	}

	// Подсчет общего количества элементов
	totalItems := len(leftovers)
	totalPages := (totalItems + pageSize - 1) / pageSize // <-- Исправлено

	// Проверка выхода за границы
	if page < 1 || page > totalPages {
		return []models.InventoryItem{}, totalPages, fmt.Errorf("page %d out of range", page)
	}

	// Пагинация
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalItems {
		end = totalItems
	}

	return leftovers[start:end], totalPages, nil // <-- Возвращаем totalPages вместо totalItems
}
