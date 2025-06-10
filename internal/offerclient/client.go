package offerclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	repoangebot "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service/repo_angebot"
)

type OfferClient string

func NewOfferClient(url string) *OfferClient {
	client := OfferClient(url)
	return &client
}

func (c *OfferClient) GetOffersByFilter(filter repoangebot.Filter) ([]*repoangebot.Offer, error) {
	byteJson, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, string(*c)+"/filter", bytes.NewBuffer(byteJson))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to get offers, status code: %d", resp.StatusCode)
	}

	var offers []*repoangebot.Offer
	if err := json.NewDecoder(resp.Body).Decode(&offers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return offers, nil
}
