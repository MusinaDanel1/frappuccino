package service

import (
	"fmt"
	"frappuccino/internal/dal"
	"frappuccino/models"
)

type MenuService struct {
	repo dal.MenuInterface
}

func NewMenuItemService(repo dal.MenuInterface) *MenuService {
	return &MenuService{
		repo: repo,
	}
}

func (s *MenuService) CreateMenuItem(item models.MenuItem) error {
	if err := s.repo.Create(item); err != nil {
		return fmt.Errorf("failed to create ingredient : %w", err)
	}
	return nil
}

func (s *MenuService) GetByID(menuID string) (models.MenuItem, error) {
	return s.repo.GetByID(menuID)
}

func (s *MenuService) Update(item models.MenuItem, id string) error {
	return s.repo.Update(item, id)
}

func (s *MenuService) Delete(menuID string) error {
	return s.repo.Delete(menuID)
}

func (s *MenuService) List() ([]models.MenuItem, error) {
	return s.repo.List()
}
