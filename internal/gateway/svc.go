package gateway

import (
	"net/http"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
)

type Service struct {
	*server.Server
}

func New() *Service {
	svr := &Service{server.NewServer()}
	setupRoutes(svr)
	return svr
}

func setupRoutes(s *Service) {
	s.WithHandlerFunc("/health", s.HealthCheck, http.MethodGet)
}
