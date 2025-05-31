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
	log         zerolog.Logger
	port        int
	Router      *mux.Router
	logRequests bool
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

		// Wrap the response writer to capture the status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		fullURL := fmt.Sprintf("%s://%s%s", GetScheme(r), r.Host, r.RequestURI)

		level := s.log.Debug()
		if rw.statusCode >= 500 {
			level = s.log.Error()
		}

		level.
			Str("url", fullURL).
			Str("method", r.Method).
			Int("status code", rw.statusCode).
			Int("duration", int(duration.Milliseconds())).
			Str("content", string(rw.content)).
			Any("args", rw.Header()).
			Msg("request")
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	wroteHeader  bool
	wroteContent bool
	content      []byte
}

// WriteHeader sets the status code
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write writes the body and sets 200 if not already set
func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.wroteContent = true
	rw.content = append(rw.content, b...)

	if !rw.wroteHeader {
		// Status code not explicitly set, default to 200
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
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

func Error(w http.ResponseWriter, message string, code int) {
	http.Error(w, message, code)
}

func Warning(w http.ResponseWriter, code int, message string) {
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
