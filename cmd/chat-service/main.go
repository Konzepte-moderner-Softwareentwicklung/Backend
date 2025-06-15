package main

import (
	"flag"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/chatservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/chatservice/service/repo"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/rs/zerolog"
)

var (
	secretString string
	mongoUrl     string
	natsUrl      string
	verbose      bool
	port         int
)

func main() {
	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.StringVar(&mongoUrl, "mongo-url", "mongodb://mongo:27017", "MongoDB URL")
	flag.StringVar(&natsUrl, "nats-url", "nats://nats:4222", "NATS URL")
	flag.StringVar(&secretString, "jwt", "some jwt key", "JWT key")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	repo := repo.NewMongoRepo(mongoUrl)

	var level zerolog.Level = zerolog.InfoLevel
	if verbose {
		level = zerolog.DebugLevel
	}

	logger := logstash.NewZerologLogger("chat-service", level)
	svc := chatservice.New([]byte(secretString), repo, natsUrl)
	svc.
		WithPort(port).
		WithLogger(logger).
		ListenAndServe()

}
