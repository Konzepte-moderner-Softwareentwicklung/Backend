package main

import (
	"log"
	"os"
	"strconv"

	_ "embed"

	_ "github.com/Konzepte-moderner-Softwareentwicklung/Backend/cmd/media-service/docs"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/mediaservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/mediaservice/service"
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
	port            int
	isVerbose       bool = false
	minioUrl        string
	accessKeyID     string
	secretAccessKey string
)

//go:embed version.json
var versionJSON string

// @title Media Service API
// @version 1.0
// @description This is the API for the Media Service
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
	minioUrl = os.Getenv("MINIO_URL")
	secretAccessKey = os.Getenv("MINIO_ACCESS_KEY")
	accessKeyID = os.Getenv("MINIO_ACCESS_KEY_ID")

	var loglevel = zerolog.InfoLevel
	if isVerbose {
		loglevel = zerolog.DebugLevel
	}
	logger := logstash.NewZerologLogger("media-service", loglevel)
	logger = version.LoggerWithVersion(versionJSON, logger)
	minio, err := service.New(minioUrl, accessKeyID, secretAccessKey)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create MinIO client")
	}
	ms := mediaservice.New(minio)

	var isSwagger = os.Getenv("SWAGGER") == "true"
	if isSwagger {
		ms.Router.PathPrefix(server.SWAGGER_PATH).Handler(httpSwagger.WrapHandler)
	}
	ms.
		WithPort(port).
		WithLogger(logger).
		WithLogRequest().
		WithVersion("1.0.0").
		ListenAndServe()
}
