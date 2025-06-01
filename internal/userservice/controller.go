package userservice

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/hasher"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/jwt"
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
	*jwt.Encoder
	service *UserService
}

func New(svc *UserService, secret []byte) *UserController {
	svr := &UserController{
		Server:         server.NewServer(),
		AuthMiddleware: auth.NewAuthMiddleware(secret),
		Encoder:        jwt.NewEncoder(secret),
		service:        svc,
	}

	svr.setupRoutes()

	return svr
}

func (c *UserController) setupRoutes() {
	// user
	c.WithHandlerFunc("/", c.GetUsers, http.MethodGet)
	c.WithHandlerFunc("/", c.CreateUser, http.MethodPost)
	c.WithHandlerFunc("/{id}", c.EnsureJWT(c.UpdateUser), http.MethodPut)
	c.WithHandlerFunc("/{id}", c.EnsureJWT(c.DeleteUser), http.MethodDelete)
	c.WithHandlerFunc("/{id}", c.GetUser, http.MethodGet)

	// login
	c.WithHandlerFunc("/login", c.GetLoginToken, http.MethodPost)
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

func (c *UserController) GetLoginToken(w http.ResponseWriter, r *http.Request) {
	credentials := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "",
		Password: "",
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Fehler beim Lesen der Anfrage", http.StatusBadRequest)
		return
	}

	user, err := c.service.repo.GetUserByEmail(credentials.Email)
	if err != nil {
		http.Error(w, "Nutzer nicht gefunden", http.StatusInternalServerError)
		return
	}

	// Verify the password
	if err := hasher.VerifyPassword(user.Password, credentials.Password); err != nil {
		http.Error(w, "Falsches Passwort", http.StatusUnauthorized)
		return
	}

	token, err := c.Encoder.EncodeUUID(user.ID, time.Duration(24*time.Hour))
	if err != nil {
		http.Error(w, "Fehler beim Generieren des Tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"token": token})
	if err != nil {
		http.Error(w, "Fehler beim Kodieren der Antwort", http.StatusInternalServerError)
	}
}
