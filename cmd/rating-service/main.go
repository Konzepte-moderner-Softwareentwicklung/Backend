package main

import (
	"flag"
	"os"

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
	mongoURL  string
)

func main() {
	// Kommandozeilen-Parameter definieren
	flag.IntVar(&port, "port", DEFAULT_PORT, "Port to listen on")
	flag.StringVar(&mongoURL, "mongo", "mongodb://mongo:27017", "MongoDB URL")
	flag.BoolVar(&isVerbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	// Logging-Konfiguration
	var loglevel zerolog.Level = zerolog.ErrorLevel
	if isVerbose {
		loglevel = zerolog.DebugLevel
	}
	logger := zerolog.New(os.Stdout).Level(loglevel)

	// Repository initialisieren
	repository, err := repo.NewMongoRepo(mongoURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create MongoDB repository")
	}

	// Service initialisieren
	service := ratingservice.NewRatingService(repository)

	// HTTP-Service starten
	ratingservice.New(service).
		WithLogger(logger).
		WithLogRequest().
		WithVersion("1.0.0").
		WithPort(port).
		ListenAndServe()
}
