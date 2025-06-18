package main

import (
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/ratingservice"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"log"
	"os"
	"strconv"
)

const (
	DEFAULT_PORT = 8080
)

var (
	isVerbose bool
	port      int
	natsURL   string
	mongoURL  string
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Println("Failed to parse PORT environment variable")
		panic(err)
	}

	if os.Getenv("VERBOSE") == "true" {
		isVerbose = true
	}

	natsURL = os.Getenv("NATS_URL")
	mongoURL = os.Getenv("MONGO_URL")

	// Set log level based on verbose flag
	var loglevel = zerolog.ErrorLevel
	if isVerbose {
		loglevel = zerolog.DebugLevel
	}

	// Initialize logger
	logger := logstash.NewZerologLogger("rating-service", loglevel)

	// Create the user service

	service := ratingservice.NewService(natsURL, ratingservice.NewMongoRepo(mongoURL))
	service.WithLogger(logger)
	done := make(chan struct{})

	go service.StartNats(done)

	service.WithLogger(logger).WithLogRequest().WithPort(port).ListenAndServe()

}
