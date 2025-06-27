package chatservice

import (
	"bytes"
	"encoding/json"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/chatservice/service/mocks"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/chatservice/service/repo"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setupTestController() (*ChatController, *mocks.MockService) {
	mockService := mocks.NewMockService()

	controller := &ChatController{
		service: mockService,
		Server:  server.NewServer(),
	}
	controller.setupRoutes()

	return controller, mockService
}

func TestHandleGetChats(t *testing.T) {
	controller, service := setupTestController()

	userId := uuid.New()
	expectedChats := []repo.Chat{{ID: uuid.New()}}
	service.Chats[userId] = expectedChats

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("UserId", userId.String())
	w := httptest.NewRecorder()

	controller.HandleGetChats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var actual []repo.Chat
	err := json.Unmarshal(w.Body.Bytes(), &actual)
	assert.NoError(t, err)
	assert.Equal(t, expectedChats, actual)
}

func TestHandleGetChat(t *testing.T) {
	controller, service := setupTestController()

	chatId := uuid.New()
	userId := uuid.New()

	expectedMessages := []repo.Message{
		{Content: "Hello"},
	}
	service.Messages[chatId] = expectedMessages

	req := httptest.NewRequest(http.MethodGet, "/"+chatId.String(), nil)
	req = mux.SetURLVars(req, map[string]string{"chatId": chatId.String()})
	req.Header.Set("UserId", userId.String())
	w := httptest.NewRecorder()

	controller.HandleGetChat(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var actual []repo.Message
	err := json.Unmarshal(w.Body.Bytes(), &actual)
	assert.NoError(t, err)
	assert.Equal(t, expectedMessages, actual)
}

func TestHandleSendMessage(t *testing.T) {
	controller, service := setupTestController()

	chatId := uuid.New()
	userId := uuid.New()
	content := "Test message"

	body, _ := json.Marshal(map[string]string{"content": content})
	req := httptest.NewRequest(http.MethodPost, "/"+chatId.String(), bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"chatId": chatId.String()})
	req.Header.Set("UserId", userId.String())
	w := httptest.NewRecorder()

	controller.HandleSendMessage(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	messages := service.Messages[chatId]
	assert.Len(t, messages, 1)
	assert.Equal(t, content, messages[0].Content)
}
