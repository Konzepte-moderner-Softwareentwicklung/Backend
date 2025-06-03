package ratingservice

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/ratingservice/repo"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/jwt"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/middleware/auth"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	UserIdHeader = "UserId"
)

type RatingController struct {
	*server.Server
	*auth.AuthMiddleware
	*jwt.Encoder
	service *RatingService
}

func New(svc *RatingService, secret []byte) *RatingController {
	svr := &RatingController{
		Server:         server.NewServer(),
		AuthMiddleware: auth.NewAuthMiddleware(secret),
		Encoder:        jwt.NewEncoder(secret),
		service:        svc,
	}

	svr.setupRoutes()

	return svr
}

func (c *RatingController) setupRoutes() {
	c.WithHandlerFunc("/", c.GetRatings, http.MethodGet)
	c.WithHandlerFunc("/", c.EnsureJWT(c.CreateRating), http.MethodPost)
	c.WithHandlerFunc("/{id}", c.EnsureJWT(c.UpdateRating), http.MethodPut)
	c.WithHandlerFunc("/{id}", c.EnsureJWT(c.DeleteRating), http.MethodDelete)
	c.WithHandlerFunc("/{id}", c.GetRating, http.MethodGet)
}

func (c *RatingController) GetRatings(w http.ResponseWriter, r *http.Request) {
	ratings, err := c.service.GetRatings()
	if err != nil {
		http.Error(w, "Fehler beim Laden der Bewertungen", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		http.Error(w, "Fehler beim Kodieren der Antwort", http.StatusInternalServerError)
	}
}

func (c *RatingController) CreateRating(w http.ResponseWriter, r *http.Request) {
	var rating repo.Rating
	if err := json.NewDecoder(r.Body).Decode(&rating); err != nil {
		http.Error(w, "Fehler beim Lesen der Anfrage", http.StatusBadRequest)
		return
	}

	rating.ID = uuid.New()
	rating.CreatedAt = time.Now()

	if err := c.service.CreateRating(rating); err != nil {
		http.Error(w, "Fehler beim Erstellen der Bewertung", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"id": rating.ID.String()})
}

func (c *RatingController) GetRating(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "ID fehlt", http.StatusBadRequest)
		return
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Fehler beim Parsen der ID", http.StatusBadRequest)
		return
	}

	rating, err := c.service.GetRatingByID(uid)
	if err != nil {
		http.Error(w, "Fehler beim Laden der Bewertung", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rating)
}

func (c *RatingController) UpdateRating(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get(UserIdHeader)
	vars := mux.Vars(r)

	if id == "" || id != vars["id"] {
		http.Error(w, "ID fehlt oder ungültig", http.StatusBadRequest)
		return
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Fehler beim Parsen der ID", http.StatusBadRequest)
		return
	}

	var rating repo.Rating
	if err := json.NewDecoder(r.Body).Decode(&rating); err != nil {
		http.Error(w, "Fehler beim Lesen der Anfrage", http.StatusBadRequest)
		return
	}

	if err := c.service.UpdateRating(uid, rating); err != nil {
		http.Error(w, "Fehler beim Aktualisieren der Bewertung", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"id": uid.String()})
}

func (c *RatingController) DeleteRating(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get(UserIdHeader)
	vars := mux.Vars(r)

	if id == "" || id != vars["id"] {
		http.Error(w, "ID fehlt oder ungültig", http.StatusBadRequest)
		return
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Fehler beim Parsen der ID", http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteRating(uid); err != nil {
		http.Error(w, "Fehler beim Löschen der Bewertung", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
