package userservice

import (
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/hasher"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/userservice/repo"
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
