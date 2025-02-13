package handler

import (
	"encoding/json"
	"frappuccino/internal/check"
	"frappuccino/internal/service"
	"frappuccino/internal/utils"
	"frappuccino/models"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type MenuHandler struct {
	menuService *service.MenuService
	logger      *slog.Logger
}

func NewMenuHandler(menuService *service.MenuService, logFilePath string) (*MenuHandler, error) {
	logger, err := utils.SetupLogger(logFilePath)
	if err != nil {
		return nil, err
	}

	return &MenuHandler{
		menuService: menuService,
		logger:      logger,
	}, nil
}

func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed.")
		return
	}
	var item models.MenuItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		utils.SendError(w, utils.StatusConflict, "Failed to decode menu item to struct!")
		slog.Error("Failed to decode menu item to struct!", slog.Any("error", err))
		h.logger.Error("Failed to decode menu item to struct!", slog.Any("error", err))
		return
	}
	if !check.Check_Menu(w, r, item) {
		return
	}
	if err := h.menuService.CreateMenuItem(&item); err != nil {
		if strings.Contains(err.Error(), "menu item already exists") {
			utils.SendError(w, utils.StatusConflict, "Menu item with this name already exists!")
			return
		}
		utils.SendError(w, utils.StatusBadRequest, "Failed to create the menu item!")
		slog.Error("Failed to create the menu item!", slog.Any("error", err))
		h.logger.Error("Failed to create the menu item!", slog.Any("error", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
	h.logger.Info("Menu item created", slog.Int("MenuItemID", item.ID))
	slog.Info("Menu item created", "MenuItemID", item.ID)
}

func (h *MenuHandler) LissMenu(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed!")
	}

	ingredients, err := h.menuService.List()
	if err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed list menu items!")
		h.logger.Error("Failed to list menu items!", slog.Any("error", err))
		slog.Error("Failed to list menu items!", slog.Any("error", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ingredients)
	slog.Info("List of menu items displayed")
	h.logger.Info("List of menu items displayed")
	slog.Info("List of menu items displayed")
}

func (h *MenuHandler) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	itemIDStr := r.URL.Path[len("/menu/"):]
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid inventory ID", http.StatusBadRequest)
		return
	}
	item, err := h.menuService.GetByID(itemID)
	if err != nil {
		utils.SendError(w, utils.StatusNotFound, "Menu item doesn't exist or unreachable!")
		h.logger.Error("Menu item doesn't exist or unreachable!", slog.Any("error", err))
		slog.Error("Menu item doesn't exist or unreachable!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
	h.logger.Info("Got menu item by its id", slog.Int("ID", itemID))
	slog.Info("Got menu item by its id", "ID", itemID)
}

func (h *MenuHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed!")
		return
	}
	itemIDStr := r.URL.Path[len("/menu/"):]
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid inventory ID", http.StatusBadRequest)
		return
	}
	var item models.MenuItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		utils.SendError(w, utils.StatusConflict, "Failed decode menu item to struct!")
		h.logger.Error("Failed decode menu item to struct!", slog.Any("error", err))
		slog.Error("Failed to decode menu item to struct!", slog.Any("error", err))
		return
	}
	if !check.Check_Menu(w, r, item) {
		return
	}
	if err := h.menuService.Update(item, itemID); err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed update menu item!")
		h.logger.Error("Failed to update menu item!", slog.Any("error", err))
		slog.Error("Failed to update menu item!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	h.logger.Info("Menu item updated", slog.Int("ID", itemID))
	slog.Info("Menu item updated", "ID", itemID)
}

func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed.")
		return
	}
	itemIDStr := r.URL.Path[len("/menu/"):]
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid inventory ID", http.StatusBadRequest)
		return
	}
	if err := h.menuService.Delete(itemID); err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed delete menu item.")
		slog.Error("Failed delete menu item.", slog.Any("error", err))
		h.logger.Error("Failed delete menu item.", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	h.logger.Info("Menu item deleted", slog.Int("ID", itemID))
	slog.Info("Menu item deleted", "ID", itemID)
}
