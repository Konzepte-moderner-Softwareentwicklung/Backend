package service

import (
	"errors"
	"slices"
	"time"

	repoangebot "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service/repo_angebot"
	"github.com/google/uuid"
)

type OfferService interface {
	GetOffer(id uuid.UUID) (*repoangebot.Offer, error)
	CreateOffer(offer *repoangebot.Offer, url string) (uuid.UUID, error)
	OccupieOffer(offerId uuid.UUID, userId uuid.UUID, space repoangebot.Space) error
	PayOffer(offerId uuid.UUID, userId uuid.UUID) error
	GetOffersByFilter(filter repoangebot.Filter) ([]*repoangebot.Offer, error)
	EditOffer(offerId uuid.UUID, userId uuid.UUID, offer *repoangebot.Offer) error
	DeleteOffer(offerId uuid.UUID) error
}

type Service struct {
	repo repoangebot.Repo
}

func New(repo repoangebot.Repo) OfferService {
	return &Service{
		repo: repo,
	}
}

func (s *Service) DeleteOffer(offerId uuid.UUID) error {
	return s.repo.DeleteOffer(offerId)
}

func (s *Service) GetOffer(id uuid.UUID) (*repoangebot.Offer, error) {
	return s.repo.GetOffer(id)
}

func (s *Service) EditOffer(offerId uuid.UUID, userId uuid.UUID, offer *repoangebot.Offer) error {
	err := s.repo.EditOffer(offerId, userId, offer)
	return err
}

func (s *Service) CreateOffer(offer *repoangebot.Offer, url string) (uuid.UUID, error) {
	offer.CreatedAt = time.Now()
	offer.ImageURL = url
	offer.ID = uuid.New()
	return offer.ID, s.repo.CreateOffer(offer)
}

func (s *Service) OccupieOffer(offerId uuid.UUID, userId uuid.UUID, space repoangebot.Space) error {

	space.Occupier = userId
	return s.repo.OccupieOffer(offerId, userId, space)
}

func (s *Service) GetOffersByFilter(filter repoangebot.Filter) ([]*repoangebot.Offer, error) {
	return s.repo.GetOffersByFilter(filter)
}

func (s *Service) PayOffer(offerId uuid.UUID, userId uuid.UUID) error {
	offers, err := s.GetOffersByFilter(repoangebot.Filter{ID: offerId})
	if err != nil {
		return err
	}
	if !slices.Contains(offers[0].OccupiedSpace.Users(), userId) {
		return errors.New("user is not occupied")
	}
	var idx int
	for idx = 0; idx < len(offers[0].OccupiedSpace); idx++ {
		if offers[0].OccupiedSpace[idx].Occupier == userId {
			break
		}
	}

	if slices.Contains(offers[0].PaidSpaces.Users(), userId) {
		return errors.New("user is already paid")
	}

	offers[0].PaidSpaces = append(offers[0].PaidSpaces, offers[0].OccupiedSpace[idx])
	return s.repo.UpdateOffer(offers[0].ID, offers[0])
}
