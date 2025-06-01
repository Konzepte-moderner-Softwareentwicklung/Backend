package repo

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

// MockRepo is a thread-safe in-memory implementation of Repo
type MockRepo struct {
	mu    sync.RWMutex
	users map[uuid.UUID]User
}

// NewMockRepo initializes a new MockRepo
func NewMockRepo() *MockRepo {
	return &MockRepo{
		users: make(map[uuid.UUID]User),
	}
}

func (m *MockRepo) CreateUser(user User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.ID]; exists {
		return errors.New("user already exists")
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockRepo) GetUserByID(id uuid.UUID) (User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return User{}, errors.New("user not found")
	}
	return user, nil
}

func (m *MockRepo) UpdateUser(user User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.ID]; !exists {
		return errors.New("user not found")
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockRepo) DeleteUser(id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[id]; !exists {
		return errors.New("user not found")
	}
	delete(m.users, id)
	return nil
}

func (m *MockRepo) GetUsers() ([]User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *MockRepo) GetUserByEmail(email string) (User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return User{}, errors.New("user not found")
}
