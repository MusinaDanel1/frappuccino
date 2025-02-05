package utils

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Status      int
	Description string `json:"error"`
}

var (
	StatusNoContent = Error{
		Status:      http.StatusNoContent,
		Description: "",
	}
	StatusBadRequest = Error{
		Status:      http.StatusBadRequest,
		Description: "",
	}
	StatusNotFound = Error{
		Status:      http.StatusNotFound,
		Description: "",
	}
	StatusMethodNotAllowed = Error{
		Status:      http.StatusMethodNotAllowed,
		Description: "",
	}
	StatusConflict = Error{
		Status:      http.StatusConflict,
		Description: "",
	}
	StatusInternalServerError = Error{
		Status:      http.StatusInternalServerError,
		Description: "",
	}
)

func SendError(w http.ResponseWriter, err Error, description string) {
	err.Description = description
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(err)
}
