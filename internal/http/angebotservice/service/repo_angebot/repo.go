package repoangebot

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

func (l *Location) IsInRadius(radius float64, location Location) bool {
	if location.Longitude == 0 && location.Latitude == 0 {
		return true
	}
	if radius <= 0 {
		return false
	}
	return l.DistanceTo(location) <= radius
}

func (l *Location) DistanceTo(location Location) float64 {
	return math.Sqrt(math.Pow(l.Longitude-location.Longitude, 2) + math.Pow(l.Latitude-location.Latitude, 2))
}

func (s *Size) Volume() float64 {
	return s.Width * s.Height * s.Depth
}

func (s *Space) TotalVolume() float64 {
	total := 0.0
	for _, item := range s.Items {
		total += item.Volume()
	}
	return total
}

func (i *Item) Volume() float64 {
	return i.Size.Volume()
}

func (s *Space) Fits(items ...Item) bool {
	// Check total volume
	totalNewVolume := 0.0
	for _, item := range items {
		totalNewVolume += item.Volume()
	}

	if totalNewVolume > s.TotalVolume() {
		return false
	}

	return true
}

type Space struct {
	Items []Item `json:"items"`
	Seats int    `json:"seats"`
}

type Item struct {
	Size   Size `json:"size"`
	Weight int  `json:"weight"`
}

type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Depth  float64 `json:"depth"`
}

type Offer struct {
	ID            uuid.UUID `json:"id" bson:"_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	LocationFrom  Location  `json:"locationFrom"`
	LocationTo    Location  `json:"locationTo"`
	Creator       uuid.UUID `json:"creator"`
	CreatedAt     time.Time `json:"createdAt"`
	IsChat        bool      `json:"isChat"`
	ChatId        uuid.UUID `json:"chatId"`
	IsPhone       bool      `json:"isPhone"`
	IsEmail       bool      `json:"isEmail"`
	StartDateTime time.Time `json:"startDateTime"`
	EndDateTime   time.Time `json:"endDateTime"`
	CanTransport  Space     `json:"canTransport"`
	Occupied      bool      `json:"occupied"`
	OccupiedBy    uuid.UUID `json:"occupiedBy"`
	Restrictions  []string  `json:"restrictions"`
	Info          []string  `json:"info"`
	InfoCar       []string  `json:"infoCar"`
	ImageURL      string    `json:"imageURL"`
}

type Filter struct {
	NameStartsWith   string   `json:"nameStartsWith"`
	SpaceNeeded      Space    `json:"spaceNeeded"`
	LocationFrom     Location `json:"locationFrom"`
	LocationTo       Location `json:"locationTo"`
	LocationFromDiff float64  `json:"locationFromDiff"`
	LocationToDiff   float64  `json:"locationToDiff"`
}

type Repo interface {
	GetOffer(id uuid.UUID) (*Offer, error)
	GetOffersByFilter(filter Filter) ([]*Offer, error)
	CreateOffer(offer *Offer) error
	OccupieOffer(offerId uuid.UUID, userId uuid.UUID) error
	ReleaseOffer(offerId uuid.UUID) error
}
