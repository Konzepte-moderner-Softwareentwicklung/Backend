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

type CreateChatRequest struct {
	UserIds []uuid.UUID `json:"userIds"`
}
type ErrorResponse struct {
	Message string `json:"message"`
}
type SendMessageRequest struct {
	Content string `json:"content"`
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

// HandleGetChats godoc
// @Summary      Get user chats
// @Description  Retrieves all chat conversations for a specific user.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Auth         JWT
// @Success      200  {array}  []repo.Chat  "List of chats"
// @Failure      400  {object}  string  "Invalid user ID"
// @Failure      500  {object}  string  "Server error"
// @Router       /chat [get]
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

// CreateChat godoc
// @Summary      Create a new chat
// @Description  Creates a new chat between the authenticated user and the specified list of users.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "JWT token"
// @Param        body body CreateChatRequest true "List of user IDs to start chat with"
// @Success      200  {string}  string  "ID of the newly created chat"
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /chat [post]
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

	err = json.NewEncoder(w).Encode(chatId)
	if err != nil {
		c.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleGetChat godoc
// @Summary      Get chat messages
// @Description  Retrieves all messages in a specific chat that the user is part of.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "JWT token"
// @Param        chatId path string true "Chat ID (UUID)"
// @Success      200  {array}  repo.Message  "List of messages in the chat"
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /chat/{chatId} [get]
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

// HandleSendMessage godoc
// @Summary      Send message
// @Description  Sends a message to a specific chat.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "JWT token"
// @Param        chatId path string true "Chat ID (UUID)"
// @Param        body body SendMessageRequest true "Message content"
// @Success      201
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /chat/{chatId}/messages [post]
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
