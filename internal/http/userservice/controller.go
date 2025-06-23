package userservice

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"
	"time"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/ratingclient"
	"github.com/joho/godotenv"

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
	service      *UserService
	ratingClient *ratingclient.RatingClient
}

type ErrorResponse struct {
	Message string `json:"message"`
}
type UserCredentials struct {
	Email    string
	Password string
}

func New(svc *UserService, secret []byte) *UserController {
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env file:", err)
	}
	RATING_URL := os.Getenv("RATING_SERVICE_URL")

	svr := &UserController{
		Server:         server.NewServer(),
		AuthMiddleware: auth.NewAuthMiddleware(secret),
		Encoder:        jwt.NewEncoder(secret),
		service:        svc,
		ratingClient:   ratingclient.NewRatingClient(RATING_URL),
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
		})

	// user
	c.WithHandlerFunc("/self", c.EnsureJWT(c.GetSelfId), http.MethodGet)
	c.WithHandlerFunc("/", c.GetUsers, http.MethodGet)
	c.WithHandlerFunc("/", c.CreateUser, http.MethodPost)
	c.WithHandlerFunc("/{id}", c.EnsureJWT(c.UpdateUser), http.MethodPut)
	c.WithHandlerFunc("/{id}", c.EnsureJWT(c.DeleteUser), http.MethodDelete)
	c.WithHandlerFunc("/email", c.GetUserByEmail, http.MethodGet)
	c.WithHandlerFunc("/{id}", c.GetUser, http.MethodGet)

	// login
	c.WithHandlerFunc("/login", c.GetLoginToken, http.MethodPost)

	// passkey
	c.WithHandlerFunc("/webauthn/register/options", c.EnsureJWT(c.beginRegistration), http.MethodGet)
	c.WithHandlerFunc("/webauthn/register", c.EnsureJWT(c.finishRegistration), http.MethodPost)
	c.WithHandlerFunc("/webauthn/login/options", c.beginLogin, http.MethodGet)
	c.WithHandlerFunc("/webauthn/login", c.finishLogin, http.MethodPost)

	// rating
	c.WithHandlerFunc("/{id}/rating", c.HandleGetRating, http.MethodGet)
}

func (c *UserController) GetSelfId(w http.ResponseWriter, r *http.Request) {
	var (
		userID uuid.UUID
		err    error
	)
	if userID, err = uuid.Parse(r.Header.Get(UserIdHeader)); err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, userID)
}

// HandleGetRating godoc
// @Summary      Get ratings for a user
// @Description  Retrieves ratings by user ID.
// @Tags         ratings
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID (UUID)"
// @Success      200  {array}  []ratingservice.Rating  "User ratings"
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /users/{id}/ratings [get]
func (c *UserController) HandleGetRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
	}
	ratings, err := c.ratingClient.GetRatingsByUserID(id)
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
	}
	if err := json.NewEncoder(w).Encode(ratings); err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// beginRegistration godoc
// @Summary      Begin WebAuthn registration
// @Description  Starts the WebAuthn registration process for the authenticated user.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "User JWT token"
// @Success      200  {object}  protocol.CredentialCreation "Registration options"
// @Failure      400  {object}  ErrorResponse "Invalid user ID"
// @Failure      404  {object}  ErrorResponse "User not found"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /users/webauthn/register/options [post]
func (c *UserController) beginRegistration(w http.ResponseWriter, r *http.Request) {
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
	if err := json.NewEncoder(w).Encode(options); err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// finishRegistration godoc
// @Summary      Finish WebAuthn registration
// @Description  Completes the WebAuthn registration process for the authenticated user.
// @Tags         users
// @Accept       json
// @Produce      plain
// @Param        Authorization header string true "User JWT token"
// @Success      200  {string}  string  "Registrierung erfolgreich"
// @Failure      400  {object}  ErrorResponse "Ungültige Anfrage oder Registrierung fehlgeschlagen"
// @Failure      404  {object}  ErrorResponse "Benutzer nicht gefunden"
// @Failure      500  {object}  ErrorResponse "Interner Serverfehler"
// @Router       /users/webauthn/register [post]
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

	if _, err := w.Write([]byte("Registrierung erfolgreich")); err != nil {
		c.GetLogger().Error().Str("Fehler beim schreiben der Antwort", err.Error())
		return
	}
}

// beginLogin godoc
// @Summary      Begin WebAuthn login
// @Description  Starts the WebAuthn login process by generating login options for the user.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        email query string true "User email address"
// @Success      200  {object}  protocol.CredentialAssertion "Login options"
// @Failure      400  {object}  ErrorResponse "Ungültige E-Mail-Adresse"
// @Failure      404  {object}  ErrorResponse "Benutzer nicht gefunden"
// @Failure      500  {object}  ErrorResponse "Interner Serverfehler"
// @Router       /users/webauthn/login/options [get]
func (c *UserController) beginLogin(w http.ResponseWriter, r *http.Request) {
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
	if err := c.service.repo.UpdateUser(user); err != nil {
		log.Println("Failed to update user:", err)
	}
	if err := json.NewEncoder(w).Encode(options); err != nil {
		log.Println("Failed to encode options:", err)
	}
}

// finishLogin godoc
// @Summary      Complete WebAuthn login
// @Description  Verifies the WebAuthn login and returns a JWT token upon success.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        email query string true "User email address"
// @Success      200  {object}  map[string]string "JWT token"
// @Failure      400  {object}  ErrorResponse "Ungültige E-Mail-Adresse"
// @Failure      401  {object}  ErrorResponse "Authentifizierung fehlgeschlagen"
// @Failure      404  {object}  ErrorResponse "Benutzer nicht gefunden"
// @Failure      500  {object}  ErrorResponse "Interner Serverfehler"
// @Router       /users/webauthn/login [post]
func (c *UserController) finishLogin(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if _, err := mail.ParseAddress(email); email == "" || err != nil {
		http.Error(w, "ungültige E-Mail-Adresse", http.StatusBadRequest)
		return
	}

	user, err := c.service.repo.GetUserByEmail(email)
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
	token, err := c.EncodeUUID(user.ID, time.Duration(24*time.Hour))
	if err != nil {
		c.Error(w, "Fehler beim Generieren des Tokens", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		c.GetLogger().Err(err)
	}
}

// GetUsers godoc
// @Summary      Get all users
// @Description  Retrieves a list of all users.
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {array}   []repo.User  "List of users"
// @Failure      500  {string}  string     "Server error"
// @Router       /users [get]
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

// GetUserByEmail godoc
// @Summary      Get user by email
// @Description  Retrieves a user by their email address provided as a query parameter.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        email  query  string  true  "User email address"
// @Success      200  {object}  repo.User  "User data"
// @Failure      400  {string}  string  "Invalid email address"
// @Failure      404  {string}  string  "User not found"
// @Failure      500  {string}  string  "Server error"
// @Router       /users/email [get]
func (c *UserController) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	user, err := c.service.repo.GetUserByEmail(email)
	if err != nil {
		http.Error(w, "Fehler beim Laden des Benutzers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, "Fehler beim Kodieren der Antwort", http.StatusInternalServerError)
	}
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Creates a new user with the provided JSON payload.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      repo.User  true  "User data"
// @Success      200   {object}  map[string]string  "ID of the created user"
// @Failure      400   {string}  string  "Invalid request payload"
// @Failure      500   {string}  string  "Server error"
// @Router       /users [post]
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

// GetUser godoc
// @Summary      Get user by ID
// @Description  Retrieves a user by their unique ID provided as a path parameter. Password is omitted in the response.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  repo.User  "User data without password"
// @Failure      400  {string}  string  "Invalid or missing ID"
// @Failure      404  {string}  string  "User not found"
// @Failure      500  {string}  string  "Server error"
// @Router       /users/{id} [get]
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

// UpdateUser godoc
// @Summary      Update user
// @Description  Updates the user identified by the ID provided in the request header. The user data is passed as JSON in the request body.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        UserIdHeader  header  string     true  "User ID (UUID)"
// @Param        user          body    repo.User  true  "User data to update"
// @Success      200           {object} map[string]string  "Returns the updated user ID"
// @Failure      400           {string} string  "Invalid or missing ID / Bad request"
// @Failure      500           {string} string  "Server error updating user"
// @Router       /users/{id} [put]
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

// DeleteUser godoc
// @Summary      Delete user
// @Description  Deletes the user identified by the ID provided in the request header.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        UserIdHeader  header  string  true  "User ID (UUID)"
// @Success      204  "User deleted successfully, no content"
// @Failure      400  {string}  string  "Invalid or missing ID"
// @Failure      500  {string}  string  "Server error deleting user"
// @Router       /users/{id} [delete]
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

// GetLoginToken godoc
// @Summary      Get login token
// @Description  Authenticates a user with email and password and returns a JWT token valid for 24 hours.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        credentials  body  UserCredentials  true  "User credentials"
// @Success      200  {object}  map[string]string  "JWT token"
// @Failure      400  {string}  string  "Fehler beim Lesen der Anfrage"
// @Failure      401  {string}  string  "Falsches Passwort"
// @Failure      500  {string}  string  "Fehler beim Generieren des Tokens oder Benutzer nicht gefunden"
// @Router       /users/login [post]
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

	token, err := c.EncodeUUID(user.ID, time.Duration(24*time.Hour))
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
