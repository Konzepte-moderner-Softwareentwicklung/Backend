package repoangebot

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
	offerCollection *mongo.Collection
}

func (r *MongoRepo) ReleaseOffer(offerId uuid.UUID) error {
	//TODO implement me
	panic("implement me")
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

func (r *MongoRepo) DeleteOffer(offerId uuid.UUID) error {
	_, err := r.offerCollection.DeleteOne(context.Background(), bson.M{"_id": offerId})
	return err
}

func (r *MongoRepo) EditOffer(offerId uuid.UUID, userId uuid.UUID, offer *Offer) error {
	_, err := r.offerCollection.UpdateOne(context.Background(), bson.M{"_id": offerId}, bson.M{"$set": offer})
	return err
}

func (r *MongoRepo) Close() error {
	return r.offerCollection.Database().Client().Disconnect(context.Background())
}

func (r *MongoRepo) CreateOffer(offer *Offer) error {
	_, err := r.offerCollection.InsertOne(context.Background(), offer)
	return err
}

func (r *MongoRepo) UpdateOffer(offerId uuid.UUID, offer *Offer) error {
	_, err := r.offerCollection.UpdateOne(context.Background(), bson.M{"_id": offerId}, bson.M{"$set": offer})
	return err
}

func (r *MongoRepo) GetOffer(id uuid.UUID) (*Offer, error) {
	var offer Offer
	err := r.offerCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&offer)
	return &offer, err
}

func (r *MongoRepo) GetOffersByFilter(ft Filter) ([]*Offer, error) {
	var offers []*Offer

	// Nur Angebote laden, die nicht belegt sind
	cur, err := r.offerCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cur.Close(context.Background()); err != nil {
			log.Printf("Fehler beim Schließen der Cursor: %v", err)
		}
	}()

	for cur.Next(context.Background()) {
		var offer Offer
		if err := cur.Decode(&offer); err != nil {
			return nil, err
		}

		if !ft.IncludePassed {
			if offer.EndDateTime.Before(time.Now()) {
				continue
			}
		}

		if ft.Price != 0 && offer.Price >= ft.Price {
			continue
		}
		var nullTime = time.Time{}
		if ft.DateTime != nullTime && offer.StartDateTime.Day() == ft.DateTime.Day() {
			continue
		}

		// Titel-Filter
		if !strings.HasPrefix(strings.ToLower(offer.Title), strings.ToLower(ft.NameStartsWith)) {
			continue
		}

		// Transportgröße prüfen
		if !offer.HasEnoughFreeSpace(ft.SpaceNeeded) {
			continue
		}

		// Sitzanzahl prüfen
		if offer.CanTransport.Seats < ft.SpaceNeeded.Seats {
			continue
		}

		// Creator-Filter
		if ft.Creator != uuid.Nil && offer.Creator != ft.Creator {
			continue
		}

		// ID-Filter
		if ft.ID != uuid.Nil && offer.ID != ft.ID {
			continue
		}

		// Nutzerbezogene Filter (z. B. für eigene oder belegte Angebote)
		if ft.User != uuid.Nil {
			if !slices.Contains(offer.OccupiedSpace.Users(), ft.User) && offer.Creator != ft.User {
				continue
			}
		}

		// Zeitfilter: ft.CurrentTime muss innerhalb von Start–Ende liegen
		if !ft.CurrentTime.IsZero() {
			if ft.CurrentTime.Before(offer.StartDateTime) || ft.CurrentTime.After(offer.EndDateTime) {
				continue
			}
		}

		// Location-Filter
		if ft.LocationFrom != emptyLocation && !offer.LocationFrom.IsInRadius(ft.LocationFromDiff, ft.LocationFrom) {
			continue
		}
		if ft.LocationTo != emptyLocation && !offer.LocationTo.IsInRadius(ft.LocationToDiff, ft.LocationTo) {
			continue
		}

		// Angebot passt zu allen Filtern
		offers = append(offers, &offer)
	}

	// Cursor-Fehler nach Iteration prüfen
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return offers, nil
}

func (r *MongoRepo) OccupieOffer(offerId, userId uuid.UUID, space Space) error {
	var offer Offer
	err := r.offerCollection.FindOne(context.Background(), bson.M{"_id": offerId}).Decode(&offer)
	if err != nil {
		return err
	}

	// Prüfen, ob der Platz ausreicht
	if !offer.HasEnoughFreeSpace(space) {
		return fmt.Errorf("nicht genug freier Platz im Angebot")
	}

	spc := Space{
		Occupier: userId,
		Items:    space.Items,
		Seats:    space.Seats,
	}

	// Space und User zu den belegten hinzufügen
	offer.OccupiedSpace = append(offer.OccupiedSpace, spc)

	update := bson.M{
		"$set": bson.M{
			"occupiedSpace": offer.OccupiedSpace,
		},
	}

	_, err = r.offerCollection.UpdateOne(context.Background(), bson.M{"_id": offerId}, update)
	return err
}
