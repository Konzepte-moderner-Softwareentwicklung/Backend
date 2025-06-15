package main

import (
	"flag"
	"net/url"
	"os"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/gateway"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

const (
	DEFAULT_PORT = 8080
)

var (
	port           int
	natsURL        string
	jwtKey         string
	userService    string
	mediaService   string
	angebotService string
	chatService    string
	isVerbose      bool
)

func main() {
	flag.IntVar(&port, "port", DEFAULT_PORT, "Port to listen on")
	flag.StringVar(&natsURL, "nats", nats.DefaultURL, "NATS URL")
	flag.StringVar(&jwtKey, "jwt", "some jwt key", "JWT key")
	flag.StringVar(&userService, "user-service", "http://user-service:8080", "User service URL")
	flag.StringVar(&mediaService, "media-service", "http://media-service:8080", "Media service URL")
	flag.StringVar(&angebotService, "angebot-service", "http://angebot-service:8080", "Angebot service URL")
	flag.StringVar(&chatService, "chat-service", "http://chat-service:8080", "Chat service URL")
	flag.BoolVar(&isVerbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	var loglevel zerolog.Level = zerolog.InfoLevel
	if isVerbose {
		loglevel = zerolog.DebugLevel
	}
	logger := logstash.NewZerologLogger("gateway", loglevel)

	// parse URLs
	userServiceURL, err := url.Parse(userService)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse user service URL")
		os.Exit(1)
	}
	mediaServiceURL, err := url.Parse(mediaService)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse media service URL")
		os.Exit(1)
	}
	angebotServiceURL, err := url.Parse(angebotService)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse angebot service URL")
		os.Exit(1)
	}
	chatServiceURL, err := url.Parse(chatService)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse chat service URL")
		os.Exit(1)
	}

	var gw = gateway.New(natsURL, []byte(jwtKey), map[string]url.URL{
		"user":    *userServiceURL,
		"media":   *mediaServiceURL,
		"angebot": *angebotServiceURL,
		"chat":    *chatServiceURL,
	})

	gw.
		WithLogger(logger).
		WithLogRequest().
		WithVersion("1.0.0").
		WithPort(port)
	defer gw.Close()
	gw.ListenAndServe()
}
