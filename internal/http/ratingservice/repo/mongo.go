package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RatingMeta struct {
	ID        uuid.UUID `bson:"_id" json:"id"`
	RideID    uuid.UUID `bson:"rideId" json:"rideId"`
	RaterID   uuid.UUID `bson:"raterId" json:"raterId"`
	TargetID  uuid.UUID `bson:"targetId" json:"targetId"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
type DriverRating struct {
	RatingMeta `bson:",inline"`

	Punctuality    int `bson:"punctuality" json:"punctuality"`       // War der Fahrer pünktlich?
	AgreementsKept int `bson:"agreementsKept" json:"agreementsKept"` // Abmachungen eingehalten?
	Comfort        int `bson:"comfort" json:"comfort"`               // Wohlgefühlt?
	CargoSafe      int `bson:"cargoSafe" json:"cargoSafe"`           // Fracht intakt?
}
type PassengerRating struct {
	RatingMeta `bson:",inline"`

	Punctuality    int `bson:"punctuality" json:"punctuality"`       // War der Mitfahrer pünktlich?
	AgreementsKept int `bson:"agreementsKept" json:"agreementsKept"` // Abmachungen eingehalten?
	EnjoyedRide    int `bson:"enjoyedRide" json:"enjoyedRide"`       // Gerne mitgenommen?
}

type MongoRepo struct {
	driverRatingColl    *mongo.Collection
	passengerRatingColl *mongo.Collection
}

const (
	DBName                    = "MyCargonaut"
	CollectionDriverRating    = "driverRatings"
	CollectionPassengerRating = "passengerRatings"
)

func NewMongoRepo(mongoUri string) (*MongoRepo, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, err
	}
	db := client.Database(DBName)
	return &MongoRepo{
		driverRatingColl:    db.Collection(CollectionDriverRating),
		passengerRatingColl: db.Collection(CollectionPassengerRating),
	}, nil
}

func (r *MongoRepo) CreateDriverRating(rating DriverRating) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.driverRatingColl.InsertOne(ctx, rating)
	return err
}

func (r *MongoRepo) CreatePassengerRating(rating PassengerRating) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.passengerRatingColl.InsertOne(ctx, rating)
	return err
}

func (r *MongoRepo) GetDriverRatingByID(id uuid.UUID) (DriverRating, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var rating DriverRating
	err := r.driverRatingColl.FindOne(ctx, bson.M{"_id": id}).Decode(&rating)
	return rating, err
}

func (r *MongoRepo) GetPassengerRatingByID(id uuid.UUID) (PassengerRating, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var rating PassengerRating
	err := r.passengerRatingColl.FindOne(ctx, bson.M{"_id": id}).Decode(&rating)
	return rating, err
}

func (r *MongoRepo) UpdateDriverRating(rating DriverRating) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.driverRatingColl.UpdateOne(ctx, bson.M{"_id": rating.ID}, bson.M{"$set": rating})
	return err
}

func (r *MongoRepo) UpdatePassengerRating(rating PassengerRating) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.passengerRatingColl.UpdateOne(ctx, bson.M{"_id": rating.ID}, bson.M{"$set": rating})
	return err
}

func (r *MongoRepo) DeleteDriverRating(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.driverRatingColl.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *MongoRepo) DeletePassengerRating(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.passengerRatingColl.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *MongoRepo) GetDriverRatingsByTarget(targetID uuid.UUID) ([]DriverRating, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.driverRatingColl.Find(ctx, bson.M{"targetId": targetID})
	if err != nil {
		return nil, err
	}
	var ratings []DriverRating
	err = cursor.All(ctx, &ratings)
	return ratings, err
}

func (r *MongoRepo) GetPassengerRatingsByTarget(targetID uuid.UUID) ([]PassengerRating, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.passengerRatingColl.Find(ctx, bson.M{"targetId": targetID})
	if err != nil {
		return nil, err
	}
	var ratings []PassengerRating
	err = cursor.All(ctx, &ratings)
	return ratings, err
}
