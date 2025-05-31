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
)

func main() {
	flag.IntVar(&port, "port", DEFAULT_PORT, "Port to listen on")
	flag.StringVar(&natsURL, "nats", nats.DefaultURL, "NATS URL")
	flag.Parse()

	logger := zerolog.New(os.Stdout)

	gateway.New(natsURL).
		WithLogger(logger).
		WithLogRequest().
		WithPort(port).ListenAndServe()
}
