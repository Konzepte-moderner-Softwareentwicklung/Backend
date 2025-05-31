package userservice

import (
	"encoding/json"
	"net/http"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
)

type UserController struct {
	*server.Server
	service *UserService
}

func New(svc *UserService) *UserController {
	svr := &UserController{
		Server:  server.NewServer(),
		service: svc,
	}

	svr.setupRoutes()

	return svr
}

func (c *UserController) setupRoutes() {
	c.WithHandlerFunc("/users", c.GetUsers, http.MethodGet)
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
