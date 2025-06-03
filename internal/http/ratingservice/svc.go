package ratingservice

import (
	"errors"
	"time"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/ratingservice/repo"
	"github.com/google/uuid"
)

type RatingService struct {
	repo repo.RatingRepo
}

func NewRatingService(r repo.RatingRepo) *RatingService {
	return &RatingService{repo: r}
}

// --- DRIVER RATINGS ---

func (s *RatingService) CreateDriverRating(r repo.DriverRating) error {
	if err := validateDriverRating(r); err != nil {
		return err
	}
	r.ID = uuid.New()
	r.CreatedAt = time.Now()
	return s.repo.CreateDriverRating(r)
}

func (s *RatingService) GetDriverRatingByID(id uuid.UUID) (repo.DriverRating, error) {
	return s.repo.GetDriverRatingByID(id)
}

func (s *RatingService) UpdateDriverRating(r repo.DriverRating) error {
	if err := validateDriverRating(r); err != nil {
		return err
	}
	return s.repo.UpdateDriverRating(r)
}

func (s *RatingService) DeleteDriverRating(id uuid.UUID) error {
	return s.repo.DeleteDriverRating(id)
}

func (s *RatingService) GetDriverRatingsByTarget(targetID uuid.UUID) ([]repo.DriverRating, error) {
	return s.repo.GetDriverRatingsByTarget(targetID)
}

func (s *RatingService) GetDriverRatingsByRater(raterID uuid.UUID) ([]repo.DriverRating, error) {
	return s.repo.GetDriverRatingsByRater(raterID)
}

// --- PASSENGER RATINGS ---

func (s *RatingService) CreatePassengerRating(r repo.PassengerRating) error {
	if err := validatePassengerRating(r); err != nil {
		return err
	}
	r.ID = uuid.New()
	r.CreatedAt = time.Now()
	return s.repo.CreatePassengerRating(r)
}

func (s *RatingService) GetPassengerRatingByID(id uuid.UUID) (repo.PassengerRating, error) {
	return s.repo.GetPassengerRatingByID(id)
}

func (s *RatingService) UpdatePassengerRating(r repo.PassengerRating) error {
	if err := validatePassengerRating(r); err != nil {
		return err
	}
	return s.repo.UpdatePassengerRating(r)
}

func (s *RatingService) DeletePassengerRating(id uuid.UUID) error {
	return s.repo.DeletePassengerRating(id)
}

func (s *RatingService) GetPassengerRatingsByTarget(targetID uuid.UUID) ([]repo.PassengerRating, error) {
	return s.repo.GetPassengerRatingsByTarget(targetID)
}

func (s *RatingService) GetPassengerRatingsByRater(raterID uuid.UUID) ([]repo.PassengerRating, error) {
	return s.repo.GetPassengerRatingsByRater(raterID)
}

// --- VALIDATION ---

func validateDriverRating(r repo.DriverRating) error {
	if !inRange(r.Punctuality) || !inRange(r.Comfort) || !inRange(r.AgreementsKept) || !inRange(r.CargoSafe) {
		return errors.New("driver rating values must be between 1 and 5")
	}
	return nil
}

func validatePassengerRating(r repo.PassengerRating) error {
	if !inRange(r.Punctuality) || !inRange(r.AgreementsKept) || !inRange(r.EnjoyedRide) {
		return errors.New("passenger rating values must be between 1 and 5")
	}
	return nil
}

func inRange(val int) bool {
	return val >= 1 && val <= 5
}
