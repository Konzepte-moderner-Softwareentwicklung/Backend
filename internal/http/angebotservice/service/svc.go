package service

import (
	"time"

	repoangebot "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service/repo_angebot"
	"github.com/google/uuid"
)

type Service struct {
	repo repoangebot.Repo
}

func New(repo repoangebot.Repo) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GetOffer(id uuid.UUID) (*repoangebot.Offer, error) {
	return s.repo.GetOffer(id)
}

func (s *Service) CreateOffer(offer *repoangebot.Offer, url string) (uuid.UUID, error) {
	offer.CreatedAt = time.Now()
	offer.ImageURL = url
	offer.ID = uuid.New()
	return offer.ID, s.repo.CreateOffer(offer)
}

func (s *Service) OccupieOffer(offerId uuid.UUID, userId uuid.UUID) error {
	return s.repo.OccupieOffer(offerId, userId)
}

func (s *Service) GetOffersByFilter(filter repoangebot.Filter) ([]*repoangebot.Offer, error) {
	return s.repo.GetOffersByFilter(filter)
}
