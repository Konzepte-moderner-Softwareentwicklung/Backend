package ratingservice

import (
	"encoding/json"
	"net/http"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/ratingservice/repo"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/middleware/auth"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type RatingController struct {
	*server.Server
	*auth.AuthMiddleware
	service *RatingService
}

func NewRatingController(service *RatingService, jwtSecret []byte) *RatingController {
	c := &RatingController{
		Server:         server.NewServer(),
		AuthMiddleware: auth.NewAuthMiddleware(jwtSecret),
		service:        service,
	}
	c.setupRoutes()
	return c
}
func New(service *RatingService, jwtSecret []byte) *RatingController {
	return NewRatingController(service, jwtSecret)
}

func (c *RatingController) setupRoutes() {
	// Driver Ratings
	c.WithHandlerFunc("/driver", c.EnsureJWT(c.CreateDriverRating), http.MethodPost)
	c.WithHandlerFunc("/driver/{id}", c.GetDriverRatingByID, http.MethodGet)
	c.WithHandlerFunc("/driver/target/{targetId}", c.GetDriverRatingsByTarget, http.MethodGet)
	c.WithHandlerFunc("/driver/rater/{raterId}", c.GetDriverRatingsByRater, http.MethodGet)

	// Passenger Ratings
	c.WithHandlerFunc("/passenger", c.EnsureJWT(c.CreatePassengerRating), http.MethodPost)
	c.WithHandlerFunc("/passenger/{id}", c.GetPassengerRatingByID, http.MethodGet)
	c.WithHandlerFunc("/passenger/target/{targetId}", c.GetPassengerRatingsByTarget, http.MethodGet)
	c.WithHandlerFunc("/passenger/rater/{raterId}", c.GetPassengerRatingsByRater, http.MethodGet)
}

// --- DRIVER RATINGS ---

func (c *RatingController) CreateDriverRating(w http.ResponseWriter, r *http.Request) {
	var rating repo.DriverRating
	if err := json.NewDecoder(r.Body).Decode(&rating); err != nil {
		http.Error(w, "Ungültiger Request Body", http.StatusBadRequest)
		return
	}

	err := c.service.CreateDriverRating(rating)
	if err != nil {
		http.Error(w, "Fehler beim Speichern der Bewertung", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]string{"id": rating.ID.String()}); err != nil {
		c.Error(w, "Fehler beim Schreiben der Antwort", http.StatusInternalServerError)
		return
	}
}

func (c *RatingController) GetDriverRatingByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	rating, err := c.service.GetDriverRatingByID(id)
	if err != nil {
		http.Error(w, "Bewertung nicht gefunden", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(rating); err != nil {
		c.Error(w, "Fehler beim Codieren der Bewertung", http.StatusInternalServerError)
		return
	}
}

func (c *RatingController) GetDriverRatingsByTarget(w http.ResponseWriter, r *http.Request) {
	targetID, err := parseUUID(mux.Vars(r)["targetId"])
	if err != nil {
		http.Error(w, "Ungültige Target-ID", http.StatusBadRequest)
		return
	}

	ratings, err := c.service.GetDriverRatingsByTarget(targetID)
	if err != nil {
		http.Error(w, "Fehler beim Abrufen der Bewertungen", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		c.Error(w, "Fehler beim Codieren der Bewertungen", http.StatusInternalServerError)
		return
	}
}

func (c *RatingController) GetDriverRatingsByRater(w http.ResponseWriter, r *http.Request) {
	raterID, err := parseUUID(mux.Vars(r)["raterId"])
	if err != nil {
		http.Error(w, "Ungültige Rater-ID", http.StatusBadRequest)
		return
	}

	ratings, err := c.service.GetDriverRatingsByRater(raterID)
	if err != nil {
		http.Error(w, "Fehler beim Abrufen der Bewertungen", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		c.Error(w, "Fehler beim Codieren der Bewertungen", http.StatusInternalServerError)
		return
	}
}

// --- PASSENGER RATINGS ---

func (c *RatingController) CreatePassengerRating(w http.ResponseWriter, r *http.Request) {
	var rating repo.PassengerRating
	if err := json.NewDecoder(r.Body).Decode(&rating); err != nil {
		http.Error(w, "Ungültiger Request Body", http.StatusBadRequest)
		return
	}

	err := c.service.CreatePassengerRating(rating)
	if err != nil {
		http.Error(w, "Fehler beim Speichern der Bewertung", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]string{"id": rating.ID.String()}); err != nil {
		c.Error(w, "Fehler beim Codieren der Bewertungs-ID", http.StatusInternalServerError)
		return
	}
}

func (c *RatingController) GetPassengerRatingByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	rating, err := c.service.GetPassengerRatingByID(id)
	if err != nil {
		http.Error(w, "Bewertung nicht gefunden", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(rating); err != nil {
		c.Error(w, "Fehler beim Codieren der Bewertung", http.StatusInternalServerError)
		return
	}
}

func (c *RatingController) GetPassengerRatingsByTarget(w http.ResponseWriter, r *http.Request) {
	targetID, err := parseUUID(mux.Vars(r)["targetId"])
	if err != nil {
		http.Error(w, "Ungültige Target-ID", http.StatusBadRequest)
		return
	}

	ratings, err := c.service.GetPassengerRatingsByTarget(targetID)
	if err != nil {
		http.Error(w, "Fehler beim Abrufen der Bewertungen", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		c.Error(w, "Fehler beim Codieren der Bewertungen", http.StatusInternalServerError)
		return
	}
}

func (c *RatingController) GetPassengerRatingsByRater(w http.ResponseWriter, r *http.Request) {
	raterID, err := parseUUID(mux.Vars(r)["raterId"])
	if err != nil {
		http.Error(w, "Ungültige Rater-ID", http.StatusBadRequest)
		return
	}

	ratings, err := c.service.GetPassengerRatingsByRater(raterID)
	if err != nil {
		http.Error(w, "Fehler beim Abrufen der Bewertungen", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		c.Error(w, "Fehler beim Codieren der Bewertungen", http.StatusInternalServerError)
		return
	}
}

// --- Helper ---

func parseUUID(idStr string) (uuid.UUID, error) {
	return uuid.Parse(idStr)
}
