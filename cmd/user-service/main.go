package main

import (
	"flag"
	"os"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/userservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/userservice/repo"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

const (
	DEFAULT_PORT = 8080
)

var (
	port     int
	natsURL  string
	mongoURL string
)

func main() {
	flag.IntVar(&port, "port", DEFAULT_PORT, "Port to listen on")
	flag.StringVar(&natsURL, "nats", nats.DefaultURL, "NATS URL")
	flag.StringVar(&mongoURL, "mongo", "mongodb://mongo:27017", "MongoDB URL")
	flag.Parse()

	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	repository, err := repo.NewMongoRepo(mongoURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create MongoDB repository")
	}
	userservice.New(userservice.NewUserService(repository)).
		WithLogger(logger).
		WithLogRequest().
		WithPort(port).
		ListenAndServe()
}
