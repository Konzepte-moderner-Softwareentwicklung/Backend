package userservice

import (
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/hasher"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/userservice/repo"
	"github.com/google/uuid"
)

type UserService struct {
	repo repo.Repo
}

func NewUserService(repo repo.Repo) *UserService {
	return &UserService{
		repo: repo,
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
	return s.repo.UpdateUser(user)
}

func (s *UserService) DeleteUser(id uuid.UUID) error {
	return s.repo.DeleteUser(id)
}

func (s *UserService) CreateUser(user repo.User) error {
	hashedPassword, err := hasher.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	return s.repo.CreateUser(user)
}

func (s *UserService) GetUsers() ([]repo.User, error) {
	return s.repo.GetUsers()
}
