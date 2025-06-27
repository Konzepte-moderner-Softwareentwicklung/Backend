package main

import (
	"log"
	"os"
	"strconv"

	_ "embed"

	_ "github.com/Konzepte-moderner-Softwareentwicklung/Backend/cmd/angebot-service/docs"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service"
	repoangebot "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service/repo_angebot"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/version"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	DEFAULT_PORT = 8080
)

var (
	isVerbose bool = false
	mongoUrl  string
	jwtSecret string
)

//go:embed version.json
var versionJSON string

// @title Angebot Service API
// @version 1.0
// @description This is the API for the Angebot Service
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

	mongoUrl = os.Getenv("MONGO_URL")
	jwtSecret = os.Getenv("JWT_SECRET")

	var loglevel = zerolog.InfoLevel
	if isVerbose {
		loglevel = zerolog.DebugLevel
	}
	logger := logstash.NewZerologLogger("angebot-service", loglevel)
	logger = version.LoggerWithVersion(versionJSON, logger)
	repo, err := repoangebot.NewMongoRepo(mongoUrl)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create repository")
		os.Exit(1)
	}

	svc := service.New(repo)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create service")
		os.Exit(1)
	}

	api := angebotservice.New(*svc, []byte(jwtSecret))

	var isSwagger = os.Getenv("SWAGGER") == "true"
	if isSwagger {
		api.Router.PathPrefix(server.SWAGGER_PATH).Handler(httpSwagger.WrapHandler)
	}

	api.
		WithPort(port).
		WithLogger(logger).
		WithLogRequest().
		ListenAndServe()
}
