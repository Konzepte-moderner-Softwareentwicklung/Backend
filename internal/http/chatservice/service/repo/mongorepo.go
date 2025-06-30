package repo

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
	messageCollection *mongo.Collection
	chatCollection    *mongo.Collection
}

func NewMongoRepo(uri string) Repository {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	return &MongoRepo{
		messageCollection: client.Database("MyCargonaut").Collection("chat_messages"),
		chatCollection:    client.Database("MyCargonaut").Collection("chats"),
	}
}

func (r *MongoRepo) AddUserToChat(userId uuid.UUID, chatId uuid.UUID) error {
	filter := bson.M{"_id": chatId}
	update := bson.M{"$push": bson.M{"user_ids": userId}}
	_, err := r.chatCollection.UpdateOne(context.Background(), filter, update)
	return err
}

func (r *MongoRepo) CreateChat(userIds ...uuid.UUID) (uuid.UUID, error) {
	id := uuid.New()
	chat := Chat{
		ID:        id,
		UserIds:   userIds,
		CreatedAt: time.Now(),
	}
	_, err := r.chatCollection.InsertOne(context.Background(), chat)
	return id, err
}

func (r *MongoRepo) GetChats(userId uuid.UUID) ([]Chat, error) {
	filter := bson.M{"user_ids": bson.M{"$in": []uuid.UUID{userId}}}
	cursor, err := r.chatCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			log.Println(err)
		}
	}()

	var chats []Chat
	for cursor.Next(context.Background()) {
		var chat Chat
		if err := cursor.Decode(&chat); err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *MongoRepo) GetHistory(id uuid.UUID) ([]Message, error) {
	filter := bson.M{"chat_id": id}
	cursor, err := r.messageCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			log.Println(err)
		}
	}()

	var messages []Message
	for cursor.Next(context.Background()) {
		var message Message
		if err := cursor.Decode(&message); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// Sort messages by creation time to ensure they are in chronological order
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})

	return messages, nil
}

func (r *MongoRepo) SendMessage(message Message, chatId uuid.UUID) error {
	message.ChatId = chatId
	_, err := r.messageCollection.InsertOne(context.Background(), message)
	return err
}
