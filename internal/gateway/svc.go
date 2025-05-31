package gateway

import (
	"net/http"

	natsreciver "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/nats-reciver"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
)

type Service struct {
	*server.Server
	*natsreciver.Receiver
}

func New(natsurl string) *Service {
	reciver, err := natsreciver.New(natsurl)
	if err != nil {
		panic(err)
	}
	svr := &Service{server.NewServer(), reciver}
	setupRoutes(svr)
	return svr
}

func setupRoutes(s *Service) {
	s.WithHandlerFunc("/health", s.HealthCheck, http.MethodGet)
	s.WithHandlerFunc("/ws", s.WSHandler, http.MethodGet)
}
