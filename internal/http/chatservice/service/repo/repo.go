package repo

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID `json:"id" bson:"_id"`
	ChatId    uuid.UUID `json:"chat_id" bson:"chat_id"`
	Content   string    `json:"content" bson:"content"`
	SenderID  uuid.UUID `json:"sender_id" bson:"sender_id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type Chat struct {
	ID        uuid.UUID   `json:"id" bson:"_id"`
	UserIds   []uuid.UUID `json:"user_ids" bson:"user_ids"`
	CreatedAt time.Time   `json:"created_at" bson:"created_at"`
}

type Repository interface {
	GetHistory(chatId uuid.UUID) ([]Message, error)
	SendMessage(message Message, chatId uuid.UUID) error
	CreateChat(user ...uuid.UUID) (uuid.UUID, error)
	GetChats(userId uuid.UUID) ([]Chat, error)
}
