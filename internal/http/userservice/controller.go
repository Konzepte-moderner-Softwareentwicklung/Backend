package userservice

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/hasher"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/userservice/repo"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/jwt"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/middleware/auth"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
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
	c.WithMiddleware(
		func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				origin := r.Header.Get("Origin")
				// Set CORS headers
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

				// Handle preflight requests
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusOK)
					return
				}

				h.ServeHTTP(w, r)
			})
		},)

	// user
	c.WithHandlerFunc("/", c.GetUsers, http.MethodGet)
	c.WithHandlerFunc("/", c.CreateUser, http.MethodPost)
	c.WithHandlerFunc("/", c.EnsureJWT(c.UpdateUser), http.MethodPut)
	c.WithHandlerFunc("/", c.EnsureJWT(c.DeleteUser), http.MethodDelete)
	c.WithHandlerFunc("/{id}", c.GetUser, http.MethodGet)

	// login
	c.WithHandlerFunc("/login", c.GetLoginToken, http.MethodPost)

	// passkey
	c.WithHandlerFunc("/webauthn/register/options", c.EnsureJWT(c.beginRegistration), http.MethodGet)
	c.WithHandlerFunc("/webauthn/register", c.EnsureJWT(c.finishRegistration), http.MethodPost)
	c.WithHandlerFunc("/webauthn/login/options", c.beginLogin, http.MethodGet)
	c.WithHandlerFunc("/webauthn/login", c.finishLogin, http.MethodPost)
}

func (c *UserController)beginRegistration(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get(UserIdHeader)
	uid, err := uuid.Parse(id)
	if err != nil {
		c.Error(w, "ungültige ID", http.StatusBadRequest)
		return
	}
	user, err := c.service.GetUserByID(uid)
	if err != nil {
		c.Error(w, "Benutzer nicht gefunden", http.StatusNotFound)
		return
	}

	options, sessionData, err := c.service.webauth.BeginRegistration(user)
	if err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.SessionData = *sessionData
	if err = c.service.repo.UpdateUser(user); err != nil {
		c.Error(w, "Fehler beim Speichern der Session-Daten", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(options)
}

func (c *UserController) finishRegistration(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get(UserIdHeader)
	uid, err := uuid.Parse(id)
	if err != nil {
		c.Error(w, "ungültige ID", http.StatusBadRequest)
		return
	}
	
	user, err := c.service.GetUserByID(uid)
	if err != nil {
		c.Error(w, "Benutzer nicht gefunden", http.StatusNotFound)
		return
	}
	sessionData := user.SessionData

	cred, err := c.service.webauth.FinishRegistration(user, sessionData, r)
	if err != nil {
		c.Error(w, fmt.Sprintf("Registrierung fehlgeschlagen [%v]", err), http.StatusBadRequest)
		return
	}
	user.AddCredential(cred)
	if err := c.service.repo.UpdateUser(user); err != nil {
		c.Error(w, "Fehler beim Aktualisieren des Benutzers", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Registrierung erfolgreich"))
}

func (c *UserController)beginLogin(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if _, err := mail.ParseAddress(email); email == "" || err != nil {
		c.Error(w, "ungültige E-Mail-Adresse", http.StatusBadRequest)
		return
	}
	user, err := c.service.repo.GetUserByEmail(email)
	if err != nil {
		c.Error(w, "Benutzer nicht gefunden", http.StatusNotFound)
		return
	}

	options, sessionData, err := c.service.webauth.BeginLogin(user)
	if err != nil {
		http.Error(w, "Login fehlgeschlagen", http.StatusInternalServerError)
		return
	}

	user.SessionData = *sessionData
	c.service.repo.UpdateUser(user)
	json.NewEncoder(w).Encode(options)
}

func (c *UserController)finishLogin(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if _, err := mail.ParseAddress(email); email == "" || err != nil {
		http.Error(w, "ungültige E-Mail-Adresse", http.StatusBadRequest)
		return
	}

	user,err := c.service.repo.GetUserByEmail(email)
	if err != nil {
		http.Error(w, "Benutzer nicht gefunden", http.StatusNotFound)
		return
	}
	sessionData := user.SessionData
	_, err = c.service.webauth.FinishLogin(user, sessionData, r)
	if err != nil {
		http.Error(w, "Authentifizierung fehlgeschlagen", http.StatusUnauthorized)
		return
	}
	token, err := c.Encoder.EncodeUUID(user.ID, time.Duration(24*time.Hour))
	if err != nil {
		c.Error(w, "Fehler beim Generieren des Tokens", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}


func (c *UserController) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := c.service.GetUsers()
	if err != nil {
		c.Error(w, "Fehler beim Laden der Benutzer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		c.Error(w, "Fehler beim Kodieren der Antwort", http.StatusInternalServerError)
	}
}

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user repo.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		c.Error(w, "Fehler beim Lesen der Anfrage", http.StatusBadRequest)
		return
	}

	user.ID = uuid.New()
	err = c.service.CreateUser(user)
	if err != nil {
		c.Error(w, "Fehler beim Erstellen des Benutzers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"id": user.ID.String()})
	if err != nil {
		c.Error(w, "Fehler beim Kodieren der Antwort", http.StatusInternalServerError)
	}
}

func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		c.Error(w, "ID fehlt", http.StatusBadRequest)
		return
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		c.Error(w, "Fehler beim Parsen der ID", http.StatusBadRequest)
		return
	}

	user, err := c.service.GetUserByID(uid)
	if err != nil {
		c.Error(w, "Fehler beim Laden des Benutzers", http.StatusInternalServerError)
		return
	}

	// remove password from response
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		c.Error(w, "Fehler beim Kodieren der Antwort", http.StatusInternalServerError)
	}
}

func (c *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get(UserIdHeader)
	if id == "" {
		c.Error(w, "ID fehlt oder ungültig", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		c.Error(w, "Fehler beim Parsen der ID", http.StatusBadRequest)
		return
	}

	var user repo.User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		c.Error(w, "Fehler beim Lesen der Anfrage", http.StatusBadRequest)
		return
	}

	err = c.service.UpdateUser(uid, user)
	if err != nil {
		c.Error(w, "Fehler beim Aktualisieren des Benutzers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"id": user.ID.String()})
	if err != nil {
		c.Error(w, "Fehler beim Kodieren der Antwort", http.StatusInternalServerError)
	}
}

func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get(UserIdHeader)
	if id == "" {
		c.Error(w, "ID fehlt oder ungültig", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		c.Error(w, "Fehler beim Parsen der ID", http.StatusBadRequest)
		return
	}

	err = c.service.DeleteUser(uid)
	if err != nil {
		c.Error(w, "Fehler beim Löschen des Benutzers", http.StatusInternalServerError)
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
		c.Error(w, "Fehler beim Lesen der Anfrage", http.StatusBadRequest)
		return
	}

	user, err := c.service.repo.GetUserByEmail(credentials.Email)
	if err != nil {
		c.Error(w, fmt.Sprintf("Nutzer nicht gefunden %v", err), http.StatusInternalServerError)
		return
	}

	// Verify the password
	if err := hasher.VerifyPassword(user.Password, credentials.Password); err != nil {
		c.Error(w, fmt.Sprintf("Falsches Passwort [%v]", err), http.StatusUnauthorized)
		return
	}

	token, err := c.Encoder.EncodeUUID(user.ID, time.Duration(24*time.Hour))
	if err != nil {
		c.Error(w, "Fehler beim Generieren des Tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"token": token})
	if err != nil {
		c.Error(w, "Fehler beim Kodieren der Antwort", http.StatusInternalServerError)
	}
}
