package service

import (
	"encoding/json"
	"time"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/chatservice/service/repo"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type Service struct {
	repo     repo.Repository
	producer *nats.Conn
}

func New(repo repo.Repository, natsUrl string) *Service {
	producer, err := nats.Connect(natsUrl)
	if err != nil {
		panic(err)
	}
	return &Service{
		producer: producer,
		repo:     repo,
	}
}

func (s *Service) GetChat(chatId, userId uuid.UUID) ([]repo.Message, error) {
	return s.repo.GetHistory(chatId)
}

func (s *Service) AddUserToChat(userId uuid.UUID, chatId uuid.UUID) error {
	return s.repo.AddUserToChat(userId, chatId)
}

func (s *Service) GetChats(userId uuid.UUID) ([]repo.Chat, error) {
	return s.repo.GetChats(userId)
}

func (s *Service) CreateChat(users ...uuid.UUID) (uuid.UUID, error) {
	return s.repo.CreateChat(users...)
}

func (s *Service) SendMessage(senderID, chatId uuid.UUID, content string) error {
	message := repo.Message{
		ID:        uuid.New(),
		SenderID:  senderID,
		Content:   content,
		ChatId:    chatId,
		CreatedAt: time.Now(),
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		return err
	}
	err = s.producer.Publish("chat.message."+chatId.String(), messageData)
	if err != nil {
		return err
	}
	return s.repo.SendMessage(message, chatId)
}
