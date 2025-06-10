package repo

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
	messageCollection *mongo.Collection
}

func NewMongoRepo(uri string) Repository {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	return &MongoRepo{
		messageCollection: client.Database("MyCargonaut").Collection("chat_messages"),
	}
}

func (r *MongoRepo) GetHistory(id uuid.UUID) ([]Message, error) {
	filter := bson.M{"chat_id": id}
	cursor, err := r.messageCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

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

func (r *MongoRepo) SendMessage(message Message) error {
	_, err := r.messageCollection.InsertOne(context.Background(), message)
	return err
}
