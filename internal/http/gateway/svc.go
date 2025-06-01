package gateway

import (
	"net/http"
	"net/url"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/middleware/auth"
	natsreciver "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/nats-reciver"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
)

type Service struct {
	*server.Server
	*natsreciver.Receiver
	*auth.AuthMiddleware
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
	}
	setupRoutes(svr)
	setupProxy(svr, proxyEndpoints)

	return svr
}

func setupRoutes(s *Service) {
	s.WithHandlerFunc("/health", s.HealthCheck, http.MethodGet)
	s.WithHandlerFunc("/ws", s.WSHandler, http.MethodGet)
}
