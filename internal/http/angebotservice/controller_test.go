package angebotservice

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service/repo_angebot"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockService implements the service.Service interface
type MockService struct {
	CreateOfferFunc     func(*repoangebot.Offer, string) (uuid.UUID, error)
	GetOfferFunc        func(uuid.UUID) (*repoangebot.Offer, error)
	PayOfferFunc        func(uuid.UUID, uuid.UUID) error
	OccupieOfferFunc    func(uuid.UUID, uuid.UUID, repoangebot.Space) error
	GetOffersByFilterFn func(repoangebot.Filter) ([]repoangebot.Offer, error)
}

func (m *MockService) CreateOffer(offer *repoangebot.Offer, imageURL string) (uuid.UUID, error) {
	return m.CreateOfferFunc(offer, imageURL)
}
func (m *MockService) GetOffer(id uuid.UUID) (*repoangebot.Offer, error) {
	return m.GetOfferFunc(id)
}
func (m *MockService) PayOffer(offerID, userID uuid.UUID) error {
	return m.PayOfferFunc(offerID, userID)
}
func (m *MockService) OccupieOffer(id, userID uuid.UUID, space repoangebot.Space) error {
	return m.OccupieOfferFunc(id, userID, space)
}
func (m *MockService) GetOffersByFilter(filter repoangebot.Filter) ([]*repoangebot.Offer, error) {

	offer, err := m.GetOffersByFilter(filter)
	return offer, err
}

func TestHandleCreateOffer_Success(t *testing.T) {
	mockSvc := &MockService{
		CreateOfferFunc: func(offer *repoangebot.Offer, imageURL string) (uuid.UUID, error) {
			return uuid.New(), nil
		},
	}
	controller := &OfferController{
		service: mockSvc,
	}
	offer := repoangebot.Offer{
		Title: "Test Angebot",
	}
	body, _ := json.Marshal(offer)
	req := httptest.NewRequest(http.MethodPost, "/angebot", bytes.NewBuffer(body))
	req.Header.Set(UserIdHeader, uuid.New().String())
	rec := httptest.NewRecorder()

	controller.handleCreateOffer(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rec.Code)
	}
}

func TestPayOffer_Success(t *testing.T) {
	offerID := uuid.New()
	userID := uuid.New()
	mockSvc := &MockService{
		PayOfferFunc: func(oid, uid uuid.UUID) error {
			if oid != offerID || uid != userID {
				return errors.New("IDs mismatch")
			}
			return nil
		},
	}
	controller := &OfferController{
		service: mockSvc,
	}
	req := httptest.NewRequest(http.MethodPost, "/angebot/"+offerID.String()+"/pay", nil)
	req = mux.SetURLVars(req, map[string]string{"id": offerID.String()})
	req.Header.Set(UserIdHeader, userID.String())
	rec := httptest.NewRecorder()

	controller.PayOffer(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rec.Code)
	}
}
