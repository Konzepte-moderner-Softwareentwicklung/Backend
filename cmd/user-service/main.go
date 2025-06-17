package main

import (
	"log"
	"os"
	"strconv"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/userservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/userservice/repo"
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
	jwtKey    string
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Println("Failed to parse PORT environment variable")
		panic(err)
	}

	if os.Getenv("VERBOSE") == "true" {
		isVerbose = true
	}

	natsURL = os.Getenv("NATS_URL")
	mongoURL = os.Getenv("MONGO_URL")
	jwtKey = os.Getenv("JWT_SECRET")

	// Set log level based on verbose flag
	var loglevel = zerolog.ErrorLevel
	if isVerbose {
		loglevel = zerolog.DebugLevel
	}

	// Initialize logger
	logger := logstash.NewZerologLogger("user-service", loglevel)
	repository, err := repo.NewMongoRepo(mongoURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create MongoDB repository")
	}

	// Create the user service
	service := userservice.NewUserService(repository)

	// Start the rest service
	userservice.New(service, []byte(jwtKey)).
		WithLogger(logger).
		WithLogRequest().
		WithVersion("1.0.0").
		WithPort(port).
		ListenAndServe()
}
