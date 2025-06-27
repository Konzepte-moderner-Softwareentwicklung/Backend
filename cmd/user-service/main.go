package main

import (
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/version"
	"log"
	"os"
	"strconv"

	_ "embed"

	_ "github.com/Konzepte-moderner-Softwareentwicklung/Backend/cmd/user-service/docs"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/userservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/userservice/repo"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	DEFAULT_PORT = 8080
)

var (
	isVerbose bool

	mongoURL string
	jwtKey   string
)

//go:embed version.json
var versionJSON string

// @title User Service API
// @version 1.0
// @description This is the API for the User Service
//
//go:generate go run ../version/main.go
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

	mongoURL = os.Getenv("MONGO_URL")
	jwtKey = os.Getenv("JWT_SECRET")

	// Set log level based on verbose flag
	var loglevel = zerolog.ErrorLevel
	if isVerbose {
		loglevel = zerolog.DebugLevel
	}

	// Initialize logger
	logger := logstash.NewZerologLogger("user-service", loglevel)

	logger = version.LoggerWithVersion(versionJSON, logger)

	repository, err := repo.NewMongoRepo(mongoURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create MongoDB repository")
	}

	// Create the user service
	service := userservice.NewUserService(repository)
	api := userservice.New(service, []byte(jwtKey))
	var isSwagger = os.Getenv("SWAGGER") == "true"
	if isSwagger {
		api.Router.PathPrefix(server.SWAGGER_PATH).Handler(httpSwagger.WrapHandler)
	}

	// Start the rest service
	api.
		WithLogger(logger).
		WithLogRequest().
		WithVersion("1.0.0").
		WithPort(port).
		ListenAndServe()
}
