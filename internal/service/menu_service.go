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

func (s *MenuService) CreateMenuItem(item *models.MenuItem) error {
	menuItemID, err := s.repo.Create(*item)
	if err != nil {
		return fmt.Errorf("failed to create menu item: %w", err)
	}
	item.ID = menuItemID // Присваиваем полученный ID
	return nil
}

func (s *MenuService) GetByID(menuID int) (models.MenuItem, error) {
	return s.repo.GetByID(menuID)
}

func (s *MenuService) Update(item models.MenuItem, id int) error {
	return s.repo.Update(item, id)
}

func (s *MenuService) Delete(menuID int) error {
	return s.repo.Delete(menuID)
}

func (s *MenuService) List() ([]models.MenuItem, error) {
	return s.repo.List()
}
