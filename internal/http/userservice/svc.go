package userservice

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/hasher"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/userservice/repo"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type UserService struct {
	repo    repo.Repo
	webauth *webauthn.WebAuthn
}

func NewUserService(repo repo.Repo) *UserService {
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env file:", err)
	}

	// Load BASE_URL from environment variable or .env file if it exists
	BASE_URL := os.Getenv("BASE_URL")
	BASE_URL = strings.TrimSpace(BASE_URL)
	BASE_URL = strings.TrimSuffix(BASE_URL, "/")
	BASE_URL = strings.ReplaceAll(BASE_URL, "\n", "")

	wauth, err := webauthn.New(
		&webauthn.Config{
			RPID:          "localhost",
			RPDisplayName: "Example",
			RPOrigins:     []string{BASE_URL},
		},
	)
	if err != nil {
		panic("failed to create webauthn instance: " + err.Error())
	}

	return &UserService{
		repo:    repo,
		webauth: wauth,
	}
}

func (s *UserService) SaveUser(user repo.User) error {
	hashedPassword, err := hasher.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	return s.repo.CreateUser(user)
}

func (s *UserService) GetUserByID(id uuid.UUID) (repo.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *UserService) UpdateUser(userid uuid.UUID, user repo.User) error {
	user.ID = userid
	hashedPassword, err := hasher.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	return s.repo.UpdateUser(user)
}

func (s *UserService) DeleteUser(id uuid.UUID) error {
	return s.repo.DeleteUser(id)
}

func (s *UserService) CreateUser(user repo.User) error {
	temp, _ := s.repo.GetUserByEmail(user.Email)
	if temp.ID != uuid.Nil {
		return errors.New("email already exists")
	}

	hashedPassword, err := hasher.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	return s.repo.CreateUser(user)
}

func (s *UserService) GetUsers() ([]repo.User, error) {
	users, err := s.repo.GetUsers()
	for i := range len(users) {
		users[i].Password = ""
	}
	return users, err
}
