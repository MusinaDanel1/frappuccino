package service

import (
	"errors"
	"frappuccino/internal/dal"
	"frappuccino/models"
	"log"
)

type InventoryService struct {
	repo dal.InventoryInterface
}

func NewIngredientService(repo dal.InventoryInterface) *InventoryService {
	return &InventoryService{
		repo: repo,
	}
}

func (s *InventoryService) Create(ingredient models.InventoryItem) error {
	if err := s.repo.Create(ingredient); err != nil {
		log.Default().Printf("error creating ingredient: %v\n", err)
		return errors.New("failed to create ingredient")
	}
	return nil
}

func (s *InventoryService) GetByID(ingID string) (models.InventoryItem, error) {
	return s.repo.GetByID(ingID)
}

func (s *InventoryService) Update(ingredient models.InventoryItem, id string) error {
	return s.repo.Update(ingredient, id)
}

func (s *InventoryService) Delete(ingID string) error {
	return s.repo.Delete(ingID)
}

func (s *InventoryService) List() ([]models.InventoryItem, error) {
	return s.repo.List()
}
