package gateway

import (
	"net/http"
	"net/url"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/jwt"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/middleware/auth"
	natsreciver "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/nats-receiver"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"github.com/nats-io/nats.go"
)

type Service struct {
	*server.Server
	NR *natsreciver.Receiver
	*auth.AuthMiddleware
	*jwt.Decoder
}

func New(natsurl string, jwtSecret []byte, proxyEndpoints map[string]url.URL) *Service {
	reciver, err := natsreciver.New(natsurl)
	if err != nil {
		panic(err)
	}
	svr := &Service{
		server.NewServer(),
		reciver,
		auth.NewAuthMiddleware(jwtSecret),
		jwt.NewDecoder(jwtSecret),
	}
	setupRoutes(svr)
	setupProxy(svr, proxyEndpoints)
	svr.LogNats()
	return svr
}

func (s *Service) Close() error {
	return s.NR.Close()
}

func (s *Service) LogNats() {
	subject := "*"
	msg := make(chan []byte)

	_, err := s.NR.Subscribe(subject, func(m *nats.Msg) {
		msg <- m.Data
	})
	if err != nil {
		s.GetLogger().Err(err).Msg("Failed to subscribe to NATS subject")
		return
	}

	go func() {
		for data := range msg {
			s.GetLogger().Info().Msgf("Received NATS message on subject '%s': %s", subject, string(data))
		}
	}()
	s.GetLogger().Info().Msgf("Listening for NATS messages on subject '%s'", subject)
}

func setupRoutes(s *Service) {
	s.WithHandlerFunc("/health", s.HealthCheck, http.MethodGet)
	s.WithHandlerFunc("/ws", s.WSHandler, http.MethodGet)
	s.WithHandlerFunc("/tracking", s.HandleTracking, http.MethodGet)
	s.WithHandlerFunc("/ws/chat/{chatId}", s.HandleChatWS, http.MethodGet)
}
