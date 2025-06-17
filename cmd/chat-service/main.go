package main

import (
	"os"
	"strconv"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/chatservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/chatservice/service/repo"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

var (
	jwtSecret string
	mongoUrl  string
	natsUrl   string
	isVerbose bool = false
	port      int
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic(err)
	}
	if os.Getenv("VERBOSE") == "true" {
		isVerbose = true
	}
	mongoUrl = os.Getenv("MONGO_URL")
	jwtSecret = os.Getenv("JWT_SECRET")
	natsUrl = os.Getenv("NATS_URL")

	repo := repo.NewMongoRepo(mongoUrl)

	var level = zerolog.InfoLevel
	if isVerbose {
		level = zerolog.DebugLevel
	}

	logger := logstash.NewZerologLogger("chat-service", level)
	svc := chatservice.New([]byte(jwtSecret), repo, natsUrl)
	svc.
		WithPort(port).
		WithLogger(logger).
		ListenAndServe()

}
