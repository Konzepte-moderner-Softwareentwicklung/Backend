package ratingservice

import "github.com/google/uuid"

type Repository interface {
	CreateRating(rating *Rating) error
	GetRatings(userID uuid.UUID) ([]Rating, error)
}
