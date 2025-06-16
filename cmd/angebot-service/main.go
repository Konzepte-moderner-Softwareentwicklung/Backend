package main

import (
	"flag"
	"os"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service"
	repoangebot "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service/repo_angebot"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/rs/zerolog"
)

const (
	DEFAULT_PORT = 8080
)

var (
	port      int
	isVerbose bool
	mongoUrl  string
	jwtSecret string
)

func main() {
	flag.IntVar(&port, "port", DEFAULT_PORT, "Port to listen on")
	flag.BoolVar(&isVerbose, "verbose", false, "Enable verbose logging")
	flag.StringVar(&mongoUrl, "mongo-url", "mongodb://mongo:27017", "MongoDB URL")
	flag.StringVar(&jwtSecret, "jwt", "some jwt key", "JWT Secret")
	flag.Parse()

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
