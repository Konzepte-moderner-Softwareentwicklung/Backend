package main

import (
	"log"
	"net/url"
	"os"
	"strconv"

	_ "embed"

	_ "github.com/Konzepte-moderner-Softwareentwicklung/Backend/cmd/gateway/docs"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/gateway"
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
	natsURL        string
	jwtKey         string
	userService    string
	mediaService   string
	angebotService string
	chatService    string
	isVerbose      bool = false
)

//go:embed version.json
var versionJSON string

// @title Gateway API
// @version 1.0
// @description This is the API for the Gateway Service
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

	natsURL = os.Getenv("NATS_URL")
	jwtKey = os.Getenv("JWT_SECRET")
	userService = os.Getenv("USER_SERVICE")
	mediaService = os.Getenv("MEDIA_SERVICE")
	angebotService = os.Getenv("ANGEBOT_SERVICE")
	chatService = os.Getenv("CHAT_SERVICE")

	var loglevel = zerolog.InfoLevel
	if isVerbose {
		loglevel = zerolog.DebugLevel
	}
	logger := logstash.NewZerologLogger("gateway", loglevel)
	logger = version.LoggerWithVersion(versionJSON, logger)
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

	var isSwagger = os.Getenv("SWAGGER") == "true"
	if isSwagger {
		gw.Router.PathPrefix(server.SWAGGER_PATH).Handler(httpSwagger.WrapHandler)
	}
	gw.
		WithLogger(logger).
		WithLogRequest().
		WithVersion("1.0.0").
		WithPort(port)
	defer func() {
		if err := gw.Close(); err != nil {
			logger.Error().Err(err).Msg("Failed to close gateway")
		}
	}()
	gw.ListenAndServe()
}
