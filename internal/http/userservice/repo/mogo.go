package repo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID             uuid.UUID             `bson:"_id"             json:"id"`
	BirthDate      time.Time             `bson:"birthDate"      json:"birthDate"`
	FirstName      string                `bson:"firstName"      json:"firstName"`
	LastName       string                `bson:"lastName"       json:"lastName"`
	Email          string                `bson:"email"          json:"email"`
	Password       string                `bson:"password"       json:"-"`
	PhoneNumber    string                `bson:"phoneNumber"    json:"phoneNumber"`
	ProfilePicture string                `bson:"profilePicture" json:"profilePicture"`
	SessionData    webauthn.SessionData  `bson:"sessionData"    json:"sessionData"`
	Credentials    []webauthn.Credential `bson:"credentials"   json:"credentials"`
}

func (u User) WebAuthnID() []byte {
	return u.ID[:]
}

func (u User) WebAuthnName() string {
	return u.FirstName + " " + u.LastName
}

func (u User) WebAuthnDisplayName() string {
	return u.FirstName
}

func (u User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

func (u *User) AddCredential(cred *webauthn.Credential) {
	u.Credentials = append(u.Credentials, *cred)
}

func (u *User) UnmarshalJSON(data []byte) error {
	type Alias User // alias ohne json:"-"
	aux := &struct {
		Password string `json:"password"` // wird beim Einlesen beachtet
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	u.Password = aux.Password
	return nil
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
