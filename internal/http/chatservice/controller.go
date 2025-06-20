package chatservice

import (
	"encoding/json"
	"net/http"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/chatservice/service"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/chatservice/service/repo"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/middleware/auth"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	UserIdHeader = "UserId"
)

type ChatController struct {
	*server.Server
	service service.Service
	*auth.AuthMiddleware
}

func New(secret []byte, repo repo.Repository, natsUrl string) *ChatController {
	svc := &ChatController{
		Server:         server.NewServer(),
		service:        *service.New(repo, natsUrl),
		AuthMiddleware: auth.NewAuthMiddleware(secret),
	}
	svc.setupRoutes()
	return svc
}

func (c *ChatController) setupRoutes() {
	c.WithHandlerFunc("/", c.EnsureJWT(c.HandleGetChats), http.MethodGet)
	c.WithHandlerFunc("/", c.EnsureJWT(c.CreateChat), http.MethodPost)
	c.WithHandlerFunc("/{chatId}", c.EnsureJWT(c.HandleGetChat), http.MethodGet)
	c.WithHandlerFunc("/{chatId}", c.EnsureJWT(c.HandleSendMessage), http.MethodPost)
}

func (c *ChatController) HandleGetChats(w http.ResponseWriter, r *http.Request) {
	var (
		userId uuid.UUID
		err    error
	)

	userId, err = uuid.Parse(r.Header.Get(UserIdHeader))
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	chats, err := c.service.GetChats(userId)
	if err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(chats); err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *ChatController) CreateChat(w http.ResponseWriter, r *http.Request) {
	var (
		userId uuid.UUID
		err    error
	)

	userId, err = uuid.Parse(r.Header.Get(UserIdHeader))
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var usersRequest = struct {
		UserIds []uuid.UUID `json:"userIds"`
	}{
		UserIds: []uuid.UUID{},
	}
	if err := json.NewDecoder(r.Body).Decode(&usersRequest); err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	chatId, err := c.service.CreateChat(append(usersRequest.UserIds, userId)...)
	if err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(chatId)
}

func (c *ChatController) HandleGetChat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatId, err := uuid.Parse(vars["chatId"])
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(r.Header.Get(UserIdHeader))
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	messages, err := c.service.GetChat(chatId, userId)
	if err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(messages)
	if err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *ChatController) HandleSendMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatId, err := uuid.Parse(vars["chatId"])
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(r.Header.Get(UserIdHeader))
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var content map[string]string
	err = json.NewDecoder(r.Body).Decode(&content)

	c.GetLogger().Debug().Str("message", content["content"]).Msg("Received message")
	if err != nil {
		c.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = c.service.SendMessage(userId, chatId, content["content"])
	if err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
