package gateway

import (
	"net/http"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/middleware/auth"
	natsreciver "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/nats-reciver"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
)

type Service struct {
	*server.Server
	*natsreciver.Receiver
	*auth.AuthMiddleware
}

func New(natsurl string, jwtSecret []byte) *Service {
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
	return svr
}

func setupRoutes(s *Service) {
	s.WithHandlerFunc("/health", s.EnsureJWT(s.HealthCheck), http.MethodGet)
	s.WithHandlerFunc("/ws", s.WSHandler, http.MethodGet)
}
