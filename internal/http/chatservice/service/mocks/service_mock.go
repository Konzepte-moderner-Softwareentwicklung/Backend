package mocks

import (
	"errors"
	"sync"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/chatservice/service/repo"
	"github.com/google/uuid"
)

type Service interface {
	GetChats(userID uuid.UUID) ([]repo.Chat, error)
	CreateChat(userIDs ...uuid.UUID) (uuid.UUID, error)
	GetChat(chatID uuid.UUID, userID uuid.UUID) ([]repo.Message, error)
	SendMessage(userID, chatID uuid.UUID, content string) error
}

type MockService struct {
	Chats        map[uuid.UUID][]repo.Chat
	Messages     map[uuid.UUID][]repo.Message
	SentMessages []string
	CreatedChats []uuid.UUID

	mu sync.Mutex
}

func NewMockService() *MockService {
	return &MockService{
		Chats:    make(map[uuid.UUID][]repo.Chat),
		Messages: make(map[uuid.UUID][]repo.Message),
	}
}

func (m *MockService) GetChats(userID uuid.UUID) ([]repo.Chat, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if chats, ok := m.Chats[userID]; ok {
		return chats, nil
	}
	return []repo.Chat{}, nil
}

func (m *MockService) CreateChat(userIDs ...uuid.UUID) (uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	newChatID := uuid.New()
	m.CreatedChats = append(m.CreatedChats, newChatID)
	return newChatID, nil
}

func (m *MockService) GetChat(chatID uuid.UUID, userID uuid.UUID) ([]repo.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if messages, ok := m.Messages[chatID]; ok {
		return messages, nil
	}
	return []repo.Message{}, nil
}

func (m *MockService) SendMessage(userID, chatID uuid.UUID, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if content == "" {
		return errors.New("empty content")
	}

	m.SentMessages = append(m.SentMessages, content)
	m.Messages[chatID] = append(m.Messages[chatID], repo.Message{

		Content: content,
	})

	return nil
}
