package repo

import "github.com/google/uuid"

type Repo interface {
	CreateUser(user User) error
	GetUserByID(id uuid.UUID) (User, error)
	UpdateUser(user User) error
	DeleteUser(id uuid.UUID) error
	GetUsers() ([]User, error)
	GetUserByEmail(email string) (*User, error)
}
