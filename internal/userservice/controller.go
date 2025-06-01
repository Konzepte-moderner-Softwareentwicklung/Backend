package userservice

import (
	"encoding/json"
	"net/http"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/middleware/auth"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/userservice/repo"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	UserIdHeader = "UserId"
)

type UserController struct {
	*server.Server
	*auth.AuthMiddleware
	service *UserService
}

func New(svc *UserService, secret []byte) *UserController {
	svr := &UserController{
		Server:         server.NewServer(),
		AuthMiddleware: auth.NewAuthMiddleware(secret),
		service:        svc,
	}

	svr.setupRoutes()

	return svr
}

func (c *UserController) setupRoutes() {
	c.WithHandlerFunc("/users", c.GetUsers, http.MethodGet)
	c.WithHandlerFunc("/users", c.CreateUser, http.MethodPost)
	c.WithHandlerFunc("/users/{id}", c.EnsureJWT(c.UpdateUser), http.MethodPut)
	c.WithHandlerFunc("/users/{id}", c.EnsureJWT(c.DeleteUser), http.MethodDelete)

	c.WithHandlerFunc("/users/{id}", c.GetUser, http.MethodGet)
}

func (c *UserController) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := c.service.GetUsers()
	if err != nil {
		http.Error(w, "Fehler beim Laden der Benutzer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		http.Error(w, "Fehler beim Kodieren der Antwort", http.StatusInternalServerError)
	}
}

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user repo.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Fehler beim Lesen der Anfrage", http.StatusBadRequest)
		return
	}

	user.ID = uuid.New()
	err = c.service.CreateUser(user)
	if err != nil {
		http.Error(w, "Fehler beim Erstellen des Benutzers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"id": user.ID.String()})
	if err != nil {
		http.Error(w, "Fehler beim Kodieren der Antwort", http.StatusInternalServerError)
	}
}

func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "ID fehlt", http.StatusBadRequest)
		return
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Fehler beim Parsen der ID", http.StatusBadRequest)
		return
	}

	user, err := c.service.GetUserByID(uid)
	if err != nil {
		http.Error(w, "Fehler beim Laden des Benutzers", http.StatusInternalServerError)
		return
	}

	// remove password from response
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, "Fehler beim Kodieren der Antwort", http.StatusInternalServerError)
	}
}

func (c *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := r.Header.Get(UserIdHeader)
	if id == "" || id != vars["id"] {
		http.Error(w, "ID fehlt oder ungültig", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Fehler beim Parsen der ID", http.StatusBadRequest)
		return
	}

	var user repo.User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Fehler beim Lesen der Anfrage", http.StatusBadRequest)
		return
	}

	err = c.service.UpdateUser(uid, user)
	if err != nil {
		http.Error(w, "Fehler beim Aktualisieren des Benutzers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"id": user.ID.String()})
	if err != nil {
		http.Error(w, "Fehler beim Kodieren der Antwort", http.StatusInternalServerError)
	}
}

func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := r.Header.Get(UserIdHeader)
	if id == "" || id != vars["id"] {
		http.Error(w, "ID fehlt oder ungültig", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Fehler beim Parsen der ID", http.StatusBadRequest)
		return
	}

	err = c.service.DeleteUser(uid)
	if err != nil {
		http.Error(w, "Fehler beim Löschen des Benutzers", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
