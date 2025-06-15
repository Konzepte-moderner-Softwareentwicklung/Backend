package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoTrackingRepo struct {
	trackingCollection *mongo.Collection
}

func NewMongoTrackingRepo(mongoURI string) TrackingRepo {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		panic(err)
	}
	return &mongoTrackingRepo{
		trackingCollection: client.Database("MyCargonaut").Collection("tracking"),
	}
}

func (m *mongoTrackingRepo) SaveTracking(tracking Tracking) error {
	_, err := m.trackingCollection.InsertOne(context.Background(), tracking)
	return err
}
