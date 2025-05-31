package main

import (
	"flag"
	"os"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/gateway"
	"github.com/rs/zerolog"
)

const (
	DEFAULT_PORT = 8080
)

var (
	port int
)

func main() {
	flag.IntVar(&port, "port", DEFAULT_PORT, "Port to listen on")
	flag.Parse()

	logger := zerolog.New(os.Stdout)

	gateway.New().
		WithLogger(logger).
		WithLogRequest().
		WithPort(port).ListenAndServe()
}
