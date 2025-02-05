package handler

import (
	"encoding/json"
	"frappuccino/internal/check"
	"frappuccino/internal/service"
	"frappuccino/internal/utils"
	"frappuccino/models"
	"log/slog"
	"net/http"
)

type InventoryHandler struct {
	inventoryService *service.InventoryService
	logger           *slog.Logger
}

func NewInventoryHandler(inventoryService *service.InventoryService, logFilePath string) (*InventoryHandler, error) {
	logger, err := utils.SetupLogger(logFilePath)
	if err != nil {
		return nil, err
	}

	return &InventoryHandler{
		inventoryService: inventoryService,
		logger:           logger,
	}, nil
}

func (h *InventoryHandler) CreateIngredient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed!")
		return
	}

	var ingredient models.InventoryItem
	if err := json.NewDecoder(r.Body).Decode(&ingredient); err != nil {
		utils.SendError(w, utils.StatusConflict, "Failed to decode ingredient item to struct!")
		h.logger.Error("Failed to decode ingredient item to struct!", slog.Any("error", err))
		slog.Error("Failed to decode ingredient item to struct!", slog.Any("error", err))
		return
	}
	if !check.Check_Inventory(w, r, ingredient) {
		return
	}
	if err := h.inventoryService.Create(ingredient); err != nil {
		utils.SendError(w, utils.StatusBadRequest, "Failed to create the ingredient!")
		slog.Error("Failed to create the ingredient!", slog.Any("error", err))
		h.logger.Error("Failed to create the ingredient!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ingredient)
	h.logger.Info("Inventory created", slog.String("IngredientID", ingredient.IngredientID))
	slog.Info("Inventory created", "IngredientID", ingredient.IngredientID)
}

func (h *InventoryHandler) ListInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed!")
		return
	}
	ingredients, err := h.inventoryService.List()
	if err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed to list inventory items!")
		slog.Error("Failed to list inventory items!", slog.Any("error", err))
		h.logger.Error("Failed to list inventory items!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ingredients)
	h.logger.Info("List of inventory items displayed")
	slog.Info("List of inventory items displayed")
}

func (h *InventoryHandler) GetIngredient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed!")
		return
	}

	ingID := r.URL.Path[len("/inventory/"):]
	ingredient, err := h.inventoryService.GetByID(ingID)
	if err != nil {
		utils.SendError(w, utils.StatusNotFound, "Ingredient item doesn't exist or unreachable!")
		slog.Error("Ingredient item doesn't exist or unreachable!", slog.Any("error", err))
		h.logger.Error("Ingredient item doesn't exist or unreachable!", slog.Any("error", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ingredient)
	h.logger.Info("Got inventory by its id", slog.String("IngredientID", ingredient.IngredientID))
	slog.Info("Got inventory by its id", "IngredientID", ingredient.IngredientID)
}

func (h *InventoryHandler) UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed!")
		return
	}
	ingID := r.URL.Path[len("/inventory/"):]
	var ingredient models.InventoryItem
	if err := json.NewDecoder(r.Body).Decode(&ingredient); err != nil {
		utils.SendError(w, utils.StatusConflict, "Failed to decode ingredient item to struct!")
		slog.Error("Failed to decode ingredient item to struct!", slog.Any("error", err))
		h.logger.Error("Failed to decode ingredient item to struct!", slog.Any("error", err))
		return
	}
	if err := h.inventoryService.Update(ingredient, ingID); err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed to update ingredient item!")
		slog.Error("Failed to update ingredient item!", slog.Any("error", err))
		h.logger.Error("Failed to update ingredient item!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	h.logger.Info("Inventory updated", slog.String("IngredientID", ingredient.IngredientID))
	slog.Info("Inventory updated", "IngredientID", ingredient.IngredientID)
}

func (h *InventoryHandler) DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.SendError(w, utils.StatusMethodNotAllowed, "Method not allowed.")
		return
	}
	ingID := r.URL.Path[len("/inventory/"):]
	if err := h.inventoryService.Delete(ingID); err != nil {
		utils.SendError(w, utils.StatusInternalServerError, "Failed to delete ingredient item!")
		h.logger.Error("Failed to delete ingredient item!", slog.Any("error", err))
		slog.Error("Failed to delete ingredient item!", slog.Any("error", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	h.logger.Info("Inventory deleted", slog.String("IngredientID", ingID))
	slog.Info("Inventory deleted", "IngredientID", ingID)
}
