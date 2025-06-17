package main

import (
	"flag"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/mediaservice"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/mediaservice/service"
	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/logstash"
	"github.com/rs/zerolog"
)

const (
	DEFAULT_PORT = 8080
)

var (
	port            int
	isVerbose       bool
	minioUrl        string
	accessKeyID     string
	secretAccessKey string
)

func main() {
	flag.IntVar(&port, "port", DEFAULT_PORT, "Port to listen on")
	flag.BoolVar(&isVerbose, "verbose", false, "Enable verbose logging")
	flag.StringVar(&secretAccessKey, "secret-access-key", "secret-access-key", "Secret access key for MinIO")
	flag.StringVar(&accessKeyID, "access-key-id", "access-key-id", "Access key ID for MinIO")
	flag.StringVar(&minioUrl, "minio-url", "minio:9000", "MinIO URL")
	flag.Parse()

	var loglevel = zerolog.InfoLevel
	if isVerbose {
		loglevel = zerolog.DebugLevel
	}
	logger := logstash.NewZerologLogger("media-service", loglevel)

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
