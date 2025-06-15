package main

import (
	"flag"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/trackingservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

var (
	natsURL   string
	offerURL  string
	isVerbose = false
	mongoURL  string
)

func main() {
	flag.StringVar(&natsURL, "nats", nats.DefaultURL, "NATS URL")
	flag.StringVar(&offerURL, "offer-url", "http://angebot-service:8080", "Offer service URL")
	flag.BoolVar(&isVerbose, "verbose", false, "Enable verbose logging")
	flag.StringVar(&mongoURL, "mongo-url", "mongodb://mongo:27017", "MongoDB URL")
	flag.Parse()

	var loglevel zerolog.Level = zerolog.InfoLevel
	if isVerbose {
		loglevel = zerolog.DebugLevel
	}
	logger := logstash.NewZerologLogger("tracking-service", loglevel)

	trackingservice.NewTrackingService(natsURL, offerURL, mongoURL).
		WithLogger(logger).
		Start()
}
