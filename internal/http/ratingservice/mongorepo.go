package ratingservice

import (
	"context"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
	ratingCollection *mongo.Collection
}

func NewMongoRepo(mongoURL string) Repository {

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		panic(err)
	}

	return &MongoRepo{
		ratingCollection: client.Database("MyCargonaut").Collection("ratings"),
	}

}

func (mr *MongoRepo) CreateRating(rating *Rating) error {
	_, err := mr.ratingCollection.InsertOne(context.Background(), rating)
	return err
}

func (mr *MongoRepo) GetRatings(userID uuid.UUID) ([]Rating, error) {
	var ratings []Rating
	cursor, err := mr.ratingCollection.Find(context.Background(), bson.M{"user_id_to": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var rating Rating
		err := cursor.Decode(&rating)
		if err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}
	return ratings, nil
}
