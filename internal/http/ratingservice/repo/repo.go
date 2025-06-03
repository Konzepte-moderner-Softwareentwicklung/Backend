package repo

import "github.com/google/uuid"

// Repo definiert die Schnittstelle für das Speichern und Abrufen von Fahrer- und Mitfahrerbewertungen
type Repo interface {
	// Fahrer-Bewertungen (Mitfahrer → Fahrer)
	CreateDriverRating(rating DriverRating) error
	GetDriverRatingByID(id uuid.UUID) (DriverRating, error)
	UpdateDriverRating(rating DriverRating) error
	DeleteDriverRating(id uuid.UUID) error
	GetDriverRatingsByTarget(targetID uuid.UUID) ([]DriverRating, error)
	GetDriverRatingsByRater(raterID uuid.UUID) ([]DriverRating, error)

	// Mitfahrer-Bewertungen (Fahrer → Mitfahrer)
	CreatePassengerRating(rating PassengerRating) error
	GetPassengerRatingByID(id uuid.UUID) (PassengerRating, error)
	UpdatePassengerRating(rating PassengerRating) error
	DeletePassengerRating(id uuid.UUID) error
	GetPassengerRatingsByTarget(targetID uuid.UUID) ([]PassengerRating, error)
	GetPassengerRatingsByRater(raterID uuid.UUID) ([]PassengerRating, error)

	// Für Sichtbarkeitslogik oder Matching
	//HasUserRatedRide(rideID, raterID uuid.UUID) (bool, error)
	//SetRatingVisible(rideID uuid.UUID) error
}
