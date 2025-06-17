package main

import (
	"flag"
	"os"

	"github.com/nats-io/nats.go"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/ratingservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/ratingservice/repo"
	"github.com/rs/zerolog"
)

const (
	DEFAULT_PORT = 8080
)

var (
	isVerbose bool
	port      int
	natsURL   string
	mongoURL  string
	jwtKey    string
)

func main() {
	// Initialize flags
	flag.IntVar(&port, "port", DEFAULT_PORT, "Port to listen on")
	flag.StringVar(&natsURL, "nats", nats.DefaultURL, "NATS URL")
	flag.StringVar(&mongoURL, "mongo", "mongodb://mongo:27017", "MongoDB URL")
	flag.StringVar(&jwtKey, "jwt", "some jwt key", "JWT key")
	flag.BoolVar(&isVerbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	// Set log level based on verbose flag
	var loglevel = zerolog.ErrorLevel
	if isVerbose {
		loglevel = zerolog.DebugLevel
	}

	// Initialize logger
	logger := zerolog.New(os.Stdout).Level(loglevel)

	// Initialize repository
	repository, err := repo.NewMongoRepo(mongoURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create MongoDB repository")
	}

	// Create the rating service
	service := ratingservice.NewRatingService(repository)

	// Start the REST service
	ratingservice.New(service, []byte(jwtKey)).
		WithLogger(logger).
		WithLogRequest().
		WithVersion("1.0.0").
		WithPort(port).
		ListenAndServe("rating-service")
}
