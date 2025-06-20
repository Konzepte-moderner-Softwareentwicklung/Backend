package main

import (
	_ "github.com/Konzepte-moderner-Softwareentwicklung/Backend/cmd/rating-service/docs"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/ratingservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	httpSwagger "github.com/swaggo/http-swagger"

	"log"
	"os"
	"strconv"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/joho/godotenv"
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
)

// @title Rating Service API
// @version 1.0
// @description This is the API for the Rating Service
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
	var isSwagger = os.Getenv("SWAGGER") == "true"
	if isSwagger {
		service.Router.PathPrefix(server.SWAGGER_PATH).Handler(httpSwagger.WrapHandler)
	}

	service.WithLogger(logger).WithLogRequest().WithPort(port).ListenAndServe()

}
