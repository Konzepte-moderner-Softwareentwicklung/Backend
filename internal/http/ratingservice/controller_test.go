package ratingservice

import (
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockRepository implementiert nur GetRatings für Tests
type MockRepository struct {
	GetRatingsFunc func(userID uuid.UUID) ([]Rating, error)
}

func (m *MockRepository) CreateRating(rating *Rating) error {
	return nil
}

func (m *MockRepository) GetRatings(userID uuid.UUID) ([]Rating, error) {
	return m.GetRatingsFunc(userID)
}

func newTestService(repo Repository) *Service {
	svc := &Service{
		Server:     *server.NewServer(),
		Repository: repo,
	}
	svc.setupRoutes()
	return svc
}

func TestHandleGetRatings(t *testing.T) {
	validUUID := uuid.New()

	mockRepo := &MockRepository{
		GetRatingsFunc: func(userID uuid.UUID) ([]Rating, error) {
			return []Rating{
				{UserIDFrom: uuid.New(), UserIDTo: userID, Value: 5, Content: "Great"},
			}, nil
		},
	}

	svc := newTestService(mockRepo)
	router := svc.Router // <- Wichtig: den Router aus dem Service verwenden

	t.Run("valid_request_returns_ratings", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/"+validUUID.String(), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		// Weitere Assertions: z.B. JSON Inhalt prüfen
	})

	// Weitere Subtests z.B. invalid UUID, Repo Fehler, JSON Fehler ...
}
