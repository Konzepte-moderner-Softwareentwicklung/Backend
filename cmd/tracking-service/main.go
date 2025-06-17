package main

import (
	"os"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/trackingservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

var (
	natsURL   string
	offerURL  string
	isVerbose = false
	mongoURL  string
)

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

	trackingservice.NewTrackingService(natsURL, offerURL, mongoURL).
		WithLogger(logger).
		Start()
}
