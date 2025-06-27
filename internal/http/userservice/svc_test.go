package userservice

import (
	"os"
	"testing"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/hasher"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/userservice/repo"
	"github.com/google/uuid"
)

var svc *UserService

func TestMain(m *testing.M) {
	repo := repo.NewMockRepo()
	svc = NewUserService(repo)
	os.Exit(m.Run())
}

func TestUserService_CreateUser(t *testing.T) {
	id := uuid.New()

	user := repo.User{
		Email:    uuid.New().String() + "@example.com",
		ID:       id,
		Password: "some password",
	}

	err := svc.CreateUser(user)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUserService_GetUser(t *testing.T) {
	id := uuid.New()

	password := "some password"
	user := repo.User{
		Email:    uuid.New().String() + "@example.com",
		ID:       id,
		Password: password,
	}

	err := svc.CreateUser(user)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	got, err := svc.GetUserByID(id)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if got.ID != id {
		t.Errorf("unexpected ID: %v", got.ID)
	}

	if got.Password == user.Password {
		t.Errorf("passwords should be hashed: %v", got.Password)
	}

	if err := hasher.VerifyPassword(got.Password, password); err != nil {
		t.Errorf("passwords should match but verification failed: %v", err)
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	id := uuid.New()

	password := "some password"
	user := repo.User{
		Email:    uuid.New().String() + "@example.com",
		ID:       id,
		Password: password,
	}

	err := svc.CreateUser(user)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	got, err := svc.GetUserByID(id)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if got.ID != id {
		t.Errorf("unexpected ID: %v", got.ID)
	}

	if got.Password == user.Password {
		t.Errorf("passwords should be hashed: %v", got.Password)
	}

	if err := hasher.VerifyPassword(got.Password, password); err != nil {
		t.Errorf("passwords should match but verification failed: %v", err)
	}

	newPassword := "new password"
	user.Password = newPassword

	err = svc.UpdateUser(user.ID, user)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	got, err = svc.GetUserByID(id)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if got.ID != id {
		t.Errorf("unexpected ID: %v", got.ID)
	}

	if got.Password == user.Password {
		t.Errorf("passwords should be hashed: %v", got.Password)
	}

	if err := hasher.VerifyPassword(got.Password, newPassword); err != nil {
		t.Errorf("passwords should match but verification failed: %v", err)
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	id := uuid.New()

	password := "some password"
	user := repo.User{
		Email:    uuid.New().String() + "@example.com",
		ID:       id,
		Password: password,
	}

	err := svc.CreateUser(user)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = svc.DeleteUser(id)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = svc.GetUserByID(id)
	if err == nil {
		t.Errorf("expected error but got none")
	}
}
