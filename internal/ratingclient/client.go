package ratingclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/ratingservice"
	"github.com/google/uuid"
	"net/http"
)

type RatingClient string

func NewRatingClient(url string) *RatingClient {
	client := RatingClient(url)
	return &client
}

func (c *RatingClient) GetRatingsByUserID(userID uuid.UUID) ([]*ratingservice.Rating, error) {

	req, err := http.NewRequest(http.MethodGet, string(*c)+"/"+userID.String(), bytes.NewBuffer(nil))
	if err != nil {
		return []*ratingservice.Rating{}, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return []*ratingservice.Rating{}, err
	}

	if resp.StatusCode >= 300 {
		return []*ratingservice.Rating{}, fmt.Errorf("failed to get offers, status code: %d", resp.StatusCode)
	}

	var ratings []*ratingservice.Rating
	if err := json.NewDecoder(resp.Body).Decode(&ratings); err != nil {
		return []*ratingservice.Rating{}, fmt.Errorf("failed to decode response: %w", err)
	}
	return ratings, nil
}
