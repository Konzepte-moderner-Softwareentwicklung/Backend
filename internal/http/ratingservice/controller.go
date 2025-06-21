package ratingservice

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func (svc *Service) setupRoutes() {
	svc.WithHandlerFunc("/{user}", svc.HandleGetRatings, http.MethodGet)
}

// HandleGetRatings godoc
// @Summary      Get ratings for user
// @Description  Retrieves all ratings associated with a specific user.
// @Tags         ratings
// @Accept       json
// @Produce      json
// @Param        user path string true "User ID (UUID)"
// @Success      200  {array}  []ratingservice.Rating  "List of ratings"
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /ratings/{user} [get]
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
