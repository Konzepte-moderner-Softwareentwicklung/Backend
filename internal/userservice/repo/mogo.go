package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID             uuid.UUID `bson:"_id"`
	BirthDate      time.Time `bson:"birthDate"`
	FirstName      string    `bson:"firstName"`
	LastName       string    `bson:"lastName"`
	Email          string    `bson:"email"`
	Password       string    `bson:"password"`
	PhoneNumber    string    `bson:"phoneNumber"`
	ProfilePicture string    `bson:"profilePicture"`
}

type MongoRepo struct {
	userCollection *mongo.Collection
}

const (
	DBName         = "MyCargonaut"
	CollectionUser = "users"
)

func (r *MongoRepo) GetUsers() ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.userCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var users []User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func NewMongoRepo(mongoUri string) (*MongoRepo, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, err
	}
	return &MongoRepo{
		userCollection: client.Database(DBName).Collection(CollectionUser),
	}, nil
}

func (r *MongoRepo) CreateUser(user User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.userCollection.InsertOne(ctx, user)
	return err
}

func (r *MongoRepo) GetUserByID(id uuid.UUID) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user User
	err := r.userCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *MongoRepo) UpdateUser(user User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.userCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": user})
	return err
}

func (r *MongoRepo) DeleteUser(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.userCollection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *MongoRepo) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user User
	err := r.userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
