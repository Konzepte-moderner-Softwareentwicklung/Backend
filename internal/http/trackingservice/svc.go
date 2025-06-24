package trackingservice

import (
	"encoding/json"
	"os"
	"time"

	repoangebot "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service/repo_angebot"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/gateway"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/trackingservice/repo"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/offerclient"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

type TrackingService struct {
	queue       *nats.Conn
	logger      zerolog.Logger
	offerClient *offerclient.OfferClient
	mongoClient repo.TrackingRepo
}

func NewTrackingService(natsURL string, offerURL string, mongoURL string) *TrackingService {
	queue, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}
	svc := &TrackingService{
		queue:       queue,
		logger:      zerolog.New(os.Stdout),
		offerClient: offerclient.NewOfferClient(offerURL),
		mongoClient: repo.NewMongoTrackingRepo(mongoURL),
	}
	return svc
}

func (s *TrackingService) WithLogger(logger zerolog.Logger) *TrackingService {
	s.logger = logger
	return s
}

func (s *TrackingService) Start() {
	s.logger.Info().Msg("Starting TrackingService...")
	subjectPrefix := "tracking.user."
	_, err := s.queue.Subscribe(subjectPrefix+"*", func(msg *nats.Msg) {
		userID, err := uuid.Parse(msg.Subject[len(subjectPrefix):])
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to parse user ID from subject")
			return
		}
		offers, err := s.offerClient.GetOffersByFilter(repoangebot.Filter{
			User:        userID,
			CurrentTime: time.Now(),
		})
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to get offers by filter")
			return
		}
		if len(offers) == 0 {
			s.logger.Info().Msgf("No offers found for user %s", userID)
			return
		}
		offer := offers[0]

		var trackingRequest gateway.TrackingRequest
		err = json.Unmarshal(msg.Data, &trackingRequest)

		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to unmarshal tracking request")
			return
		}

		s.logger.Info().Dict("location", zerolog.Dict().
			Float64("lat", trackingRequest.Location.Latitude).
			Float64("lon", trackingRequest.Location.Longitude)).
			Msg("user tracking")

		if err := s.mongoClient.SaveTracking(repo.Tracking{
			UserID:   userID,
			Tracking: trackingRequest,
		}); err != nil {
			s.logger.Error().Err(err).Msg("Failed to save tracking")
			return
		}
		for _, occupied := range offer.OccupiedSpace.Users() {
			if err := s.queue.Publish("user."+occupied.String(), msg.Data); err != nil {
				s.logger.Error().Err(err).Msg("Failed to publish tracking request")
				return
			}
			s.logger.Info().Msgf("Published tracking request for user %s to offer %s, %s", userID, offer.ID, string(msg.Data))
		}

	})
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to subscribe to tracking requests")
		return
	}
	select {} // Block forever
}
