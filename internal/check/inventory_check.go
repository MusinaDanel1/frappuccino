package check

import (
	"frappuccino/internal/utils"
	"frappuccino/models"
	"net/http"
)

func Check_Inventory(w http.ResponseWriter, r *http.Request, ingredient models.InventoryItem) bool {
	if ingredient.Name == "" {
		utils.SendError(w, utils.StatusBadRequest, "Empty ingredient name in inventory items!")
		return false
	}
	if ingredient.Quantity <= 0 {
		utils.SendError(w, utils.StatusBadRequest, "Invalid quantity in inventory items! Quantity should be more than 0!")
		return false
	}
	if ingredient.Unit == "" {
		utils.SendError(w, utils.StatusBadRequest, "Empty ingredient unit in inventory items! Please specify (mg/g/kg/oz/lb/ml/l/dl/fl oz/pc/dozen/cup/tsp/tbsp/shots)!")
		return false
	}
	if !CheckUnit(ingredient.Unit) {
		utils.SendError(w, utils.StatusBadRequest, "Invalid unit of measurement! Please specify (mg/g/kg/oz/lb/ml/l/dl/fl oz/pc/dozen/cup/tsp/tbsp/shots)!")
		return false
	}
	return true
}

func CheckUnit(unit string) bool {
	units := []string{"mg", "g", "kg", "oz", "lb", "ml", "l", "cl", "dl", "fl oz", "pc", "dozen", "cup", "tsp", "tbsp", "shots"}
	for _, u := range units {
		if u == unit {
			return true
		}
	}
	return false
}
