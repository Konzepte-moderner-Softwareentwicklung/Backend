package repoangebot

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type Location struct {
	Longitude float64 `json:"longitude" bson:"longitude"`
	Latitude  float64 `json:"latitude" bson:"latitude"`
}

var emptyLocation = Location{}

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
	dx := l.Longitude - location.Longitude
	dy := l.Latitude - location.Latitude
	return math.Sqrt(dx*dx + dy*dy)
}

type SpaceSlice []Space

func (s SpaceSlice) Sum() Space {
	var result Space
	for _, space := range s {
		result = result.Add(space)
	}
	return result
}

func (s SpaceSlice) Users() []uuid.UUID {
	var result []uuid.UUID
	for _, space := range s {
		result = append(result, space.Occupier)
	}
	return result
}

type Space struct {
	Occupier uuid.UUID `json:"occupiedBy"`
	Items    []Item    `json:"items"`
	Seats    int       `json:"seats"`
}

func (s Space) Add(other Space) Space {
	return Space{
		Items: append(s.Items, other.Items...),
		Seats: s.Seats + other.Seats,
	}
}

func (s Space) Fits(max Space) bool {
	return len(s.Items) <= len(max.Items) && s.Seats <= max.Seats
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
	ID            uuid.UUID  `json:"id" bson:"_id"`
	Driver        uuid.UUID  `json:"driver"`
	IsGesuch      bool       `json:"isGesuch" bson:"isGesuch"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Price         float64    `json:"price"`
	LocationFrom  Location   `json:"locationFrom"`
	LocationTo    Location   `json:"locationTo"`
	Creator       uuid.UUID  `json:"creator"`
	CreatedAt     time.Time  `json:"createdAt"`
	IsChat        bool       `json:"isChat"`
	IsPhone       bool       `json:"isPhone"`
	IsEmail       bool       `json:"isEmail"`
	StartDateTime time.Time  `json:"startDateTime"`
	EndDateTime   time.Time  `json:"endDateTime"`
	CanTransport  Space      `json:"canTransport"`
	OccupiedSpace SpaceSlice `json:"occupiedSpace"`
	PaidSpaces    SpaceSlice `json:"paidSpaces"`
	Restrictions  []string   `json:"restrictions"`
	Info          []string   `json:"info"`
	InfoCar       []string   `json:"infoCar"`
	ImageURL      string     `json:"imageURL"`
}

func (o *Offer) HasEnoughFreeSpace(space Space) bool {
	return o.OccupiedSpace.Sum().Add(space).Fits(o.CanTransport)
}

type Filter struct {
	Price            float64   `json:"price"`
	IncludePassed    bool      `json:"includePassed"`
	DateTime         time.Time `json:"dateTime"`
	NameStartsWith   string    `json:"nameStartsWith"`
	SpaceNeeded      Space     `json:"spaceNeeded"`
	LocationFrom     Location  `json:"locationFrom"`
	LocationTo       Location  `json:"locationTo"`
	LocationFromDiff float64   `json:"locationFromDiff"`
	LocationToDiff   float64   `json:"locationToDiff"`
	User             uuid.UUID `json:"user"`
	Creator          uuid.UUID `json:"creator"`
	CurrentTime      time.Time `json:"currentTime"`
	ID               uuid.UUID `json:"id"`
}

type Repo interface {
	GetOffer(id uuid.UUID) (*Offer, error)
	GetOffersByFilter(filter Filter) ([]*Offer, error)
	CreateOffer(offer *Offer) error
	OccupieOffer(offerId uuid.UUID, userId uuid.UUID, space Space) error
	ReleaseOffer(offerId uuid.UUID) error
	UpdateOffer(offerId uuid.UUID, offer *Offer) error
	EditOffer(offerId uuid.UUID, userId uuid.UUID, offer *Offer) error
	DeleteOffer(offerId uuid.UUID) error
}
