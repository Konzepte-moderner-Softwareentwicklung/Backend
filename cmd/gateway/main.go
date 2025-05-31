package main

import (
	"flag"
	"os"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/gateway"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

const (
	DEFAULT_PORT = 8080
)

var (
	port    int
	natsURL string
	jwtKey  string
)

func main() {
	flag.IntVar(&port, "port", DEFAULT_PORT, "Port to listen on")
	flag.StringVar(&natsURL, "nats", nats.DefaultURL, "NATS URL")
	flag.StringVar(&jwtKey, "jwt", "some jwt key", "JWT key")
	flag.Parse()

	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)

	gateway.New(natsURL, []byte(jwtKey)).
		WithLogger(logger).
		WithLogRequest().
		WithPort(port).ListenAndServe()
}
