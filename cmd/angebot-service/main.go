package main

import (
	"log"
	"os"
	"strconv"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service"
	repoangebot "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service/repo_angebot"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

const (
	DEFAULT_PORT = 8080
)

var (
	port      int  = DEFAULT_PORT
	isVerbose bool = false
	mongoUrl  string
	jwtSecret string
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

	mongoUrl = os.Getenv("MONGO_URL")
	jwtSecret = os.Getenv("JWT_SECRET")

	var loglevel = zerolog.InfoLevel
	if isVerbose {
		loglevel = zerolog.DebugLevel
	}
	logger := logstash.NewZerologLogger("angebot-service", loglevel)

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

	angebotservice.New(*svc, []byte(jwtSecret)).
		WithPort(port).
		WithLogger(logger).
		WithLogRequest().
		ListenAndServe()
}
