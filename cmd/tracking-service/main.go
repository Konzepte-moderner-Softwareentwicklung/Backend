package main

import (
	_ "embed"
	"os"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/trackingservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/version"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

var (
	natsURL   string
	offerURL  string
	isVerbose = false
	mongoURL  string
)

//go:embed version.json
var versionJSON string

//go:generate go run ../version/main.go
func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	if os.Getenv("VERBOSE") == "true" {
		isVerbose = true
	}

	natsURL = os.Getenv("NATS_URL")
	offerURL = os.Getenv("ANGEBOT_SERVICE")
	mongoURL = os.Getenv("MONGO_URL")

	var loglevel = zerolog.InfoLevel
	if isVerbose {
		loglevel = zerolog.DebugLevel
	}
	logger := logstash.NewZerologLogger("tracking-service", loglevel)
	logger = version.LoggerWithVersion(versionJSON, logger)
	trackingservice.NewTrackingService(natsURL, offerURL, mongoURL).
		WithLogger(logger).
		Start()
}
