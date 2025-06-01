package jwt

import "github.com/google/uuid"

// mockDecoder implements jwt.Decodable for testing
type MockDecoder struct {
	DecodeFunc func(token string) (uuid.UUID, error)
}

func (m *MockDecoder) DecodeUUID(token string) (uuid.UUID, error) {
	return m.DecodeFunc(token)
}
