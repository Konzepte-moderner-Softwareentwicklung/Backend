package main

import (
	"os"
	"strconv"

	_ "github.com/Konzepte-moderner-Softwareentwicklung/Backend/cmd/chat-service/docs"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/chatservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/chatservice/service/repo"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
)

var (
	jwtSecret string
	mongoUrl  string
	natsUrl   string
	isVerbose bool = false
	port      int
)

// @title Chat Service API
// @version 1.0
// @description This is the API for the Chat Service
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

	var isSwagger = os.Getenv("SWAGGER") == "true"
	if isSwagger {
		svc.Router.PathPrefix(server.SWAGGER_PATH).Handler(httpSwagger.WrapHandler)
	}

	svc.
		WithPort(port).
		WithLogger(logger).
		ListenAndServe()

}
