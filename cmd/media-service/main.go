package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/mediaservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/mediaservice/service"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
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

	fmt.Printf("[%s]", accessKeyID)
	fmt.Printf("[%s]", secretAccessKey)

	minio, err := service.New(minioUrl, accessKeyID, secretAccessKey)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create MinIO client")
	}

	mediaservice.New(minio).
		WithPort(port).
		WithLogger(logger).
		WithLogRequest().
		WithVersion("1.0.0").
		ListenAndServe("media-service")
}
