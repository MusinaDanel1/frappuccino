package check

import (
	"frappuccino/internal/utils"
	"frappuccino/models"
	"net/http"
)

func Check_Menu(w http.ResponseWriter, r *http.Request, item models.MenuItem) bool {
	if item.Name == "" {
		utils.SendError(w, utils.StatusBadRequest, "Empty menu name in menu items!")
		return false
	}
	if item.Description == "" {
		utils.SendError(w, utils.StatusBadRequest, "Empty menu description in menu items!")
		return false
	}
	if item.Price <= 0 {
		utils.SendError(w, utils.StatusBadRequest, "Menu item can't be less than 0 in menu items!")
		return false
	}
	if len(item.Category) == 0 {
		utils.SendError(w, utils.StatusBadRequest, "Empty menu category in menu items!")
		return false
	}
	if len(item.Ingredients) == 0 {
		utils.SendError(w, utils.StatusBadRequest, "Empty ingredient list in menu items!")
		return false
	}
	return true
}
