package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

const (
	DEFAULT_PORT = 8080
)

var (
	DEFAULT_LOGGER = zerolog.New(os.Stdout).With().Timestamp().Logger()
	DEFAULT_ROUTER = mux.NewRouter()
)

type Server struct {
	log    zerolog.Logger
	port   int
	Router *mux.Router
}

func NewServer() *Server {
	return &Server{
		log:    DEFAULT_LOGGER,
		port:   DEFAULT_PORT,
		Router: DEFAULT_ROUTER,
	}
}

func (s *Server) GetLogger() *zerolog.Logger {
	return &s.log
}

func (s *Server) WithLogger(logger zerolog.Logger) *Server {
	s.log = logger
	return s
}

func (s *Server) WithLogRequest() *Server {
	s.WithMiddleware(s.logRequestMiddleware)
	return s
}

func (s *Server) logRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		duration := time.Since(start)
		fullURL := fmt.Sprintf("%s://%s%s", GetScheme(r), r.Host, r.RequestURI)

		level := s.log.Info()
		next.ServeHTTP(w, r)
		level.
			Str("url", fullURL).
			Str("method", r.Method).
			Int("duration", int(duration.Milliseconds())).
			Msg("request")
	})
}

func (s *Server) WithPort(port int) *Server {
	s.port = port
	return s
}

func (s *Server) WithRouter(router *mux.Router) *Server {
	s.Router = router
	return s
}

func (s *Server) WithMiddleware(middleware func(http.Handler) http.Handler) *Server {
	s.Router.Use(middleware)
	return s
}

func (s *Server) WithHandlerFunc(path string, handler http.HandlerFunc, methods ...string) *Server {
	s.Router.HandleFunc(path, handler).Methods(methods...)
	return s
}

func (s *Server) Error(w http.ResponseWriter, message string, code int) {
	s.log.Error().Msgf("Error: %s", message)
	http.Error(w, message, code)
}

func (s *Server) WithVersion(version string) *Server {
	s.WithHandlerFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "Version: %s", version)
		if err != nil {
			s.log.Error().Err(err).Msg("Failed to write version response")
		}
	}, http.MethodGet)
	return s
}

func (s *Server) Info(message string) {
	s.log.Info().Msg(message)
}

func (s *Server) Warning(w http.ResponseWriter, code int, message string) {
	s.log.Warn().Msgf("Warning: %s", message)
	w.Header().Add("Warning", fmt.Sprintf(`%d - "%s"`, code, message))
}

func GetScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	// Optional: "X-Forwarded-Proto" für Reverse Proxies prüfen
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	return "http"
}

func (s *Server) ListenAndServe() {
	s.log.Print("Server started on port ", s.port)
	s.log.Error().AnErr("startup", http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.Router))
}
