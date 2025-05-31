package gateway

import (
	"fmt"
	"net/http"
)

func (s *Service) HealthCheck(w http.ResponseWriter, r *http.Request) {
	s.GetLogger().Info().Msg("Health check")
	fmt.Fprintf(w, "Authenticated with id {%s}", r.Header.Get("UserId"))
	w.WriteHeader(http.StatusOK)
}
