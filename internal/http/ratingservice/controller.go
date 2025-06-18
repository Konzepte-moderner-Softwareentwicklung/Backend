package ratingservice

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (svc *Service) setupRoutes() {
	svc.WithHandlerFunc("/{user}", svc.HandleGetRatings, http.MethodGet)
}

func (svc *Service) HandleGetRatings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]

	userID, err := uuid.Parse(user)
	if err != nil {
		svc.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ratings, err := svc.GetRatings(userID)
	if err != nil {
		svc.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		svc.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
