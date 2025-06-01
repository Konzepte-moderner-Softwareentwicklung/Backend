package gateway

import (
	"net/http"
)

func (s *Service) HealthCheck(w http.ResponseWriter, r *http.Request) {
	s.GetLogger().Info().Msg("Health check")
	w.WriteHeader(http.StatusOK)
}
