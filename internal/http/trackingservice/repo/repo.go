package repo

import (
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/gateway"
	"github.com/google/uuid"
)

type Tracking struct {
	UserID   uuid.UUID               `json:"user_id" bson:"user_id"`
	Tracking gateway.TrackingRequest `json:"tracking" bson:"tracking"`
}

type TrackingRepo interface {
	SaveTracking(Tracking) error
}
