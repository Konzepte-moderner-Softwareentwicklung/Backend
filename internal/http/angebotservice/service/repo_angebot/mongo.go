package repoangebot

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
	offerCollection *mongo.Collection
}

const (
	CollectionName = "offers"
	DBName         = "MyCargonaut"
)

func NewMongoRepo(mongoUri string) (*MongoRepo, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, err
	}
	return &MongoRepo{
		offerCollection: client.Database(DBName).Collection(CollectionName),
	}, nil
}

func (r *MongoRepo) Close() error {
	return r.offerCollection.Database().Client().Disconnect(context.Background())
}

func (r *MongoRepo) CreateOffer(offer *Offer) error {
	_, err := r.offerCollection.InsertOne(context.Background(), offer)
	return err
}

func (r *MongoRepo) GetOffer(id uuid.UUID) (*Offer, error) {
	var offer Offer
	err := r.offerCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&offer)
	return &offer, err
}

func (r *MongoRepo) GetOffersByFilter(ft Filter) ([]*Offer, error) {
	var offers []*Offer
	cur, err := r.offerCollection.Find(context.Background(), bson.M{
		"occupied": false,
	})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		var offer Offer
		if err := cur.Decode(&offer); err != nil {
			return nil, err
		}
		if !strings.HasPrefix(strings.ToLower(offer.Title), strings.ToLower(ft.NameStartsWith)) {
			continue
		}
		if !offer.CanTransport.Fits(ft.SpaceNeeded.Items...) {
			continue
		}
		if offer.CanTransport.Seats < ft.SpaceNeeded.Seats {
			continue
		}

		if offer.LocationFrom.IsInRadius(ft.LocationFromDiff, ft.LocationFrom) && offer.LocationTo.IsInRadius(ft.LocationToDiff, ft.LocationTo) {
			offers = append(offers, &offer)
		}
	}
	return offers, nil
}

func (r *MongoRepo) OccupieOffer(offerId uuid.UUID, userId uuid.UUID) error {
	_, err := r.offerCollection.UpdateOne(context.Background(), bson.M{"_id": offerId}, bson.M{"$set": bson.M{"occupied": true, "occupiedBy": userId}})
	if err != nil {
		return err
	}
	return nil
}

func (r *MongoRepo) ReleaseOffer(offerId uuid.UUID) error {
	_, err := r.offerCollection.UpdateOne(context.Background(), bson.M{"_id": offerId}, bson.M{"$set": bson.M{"occupied": false, "occupiedBy": nil}})
	if err != nil {
		return err
	}
	return nil
}
