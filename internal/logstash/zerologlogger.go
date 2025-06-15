package logstash

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func NewZerologLogger(serviceName string, logLevel zerolog.Level) zerolog.Logger {
	godotenv.Load()
	var (
		conn net.Conn
		err  error
	)

	// attempt to connect to logstash
	for range 50 {
		conn, err = net.Dial("tcp", os.Getenv("LOGSTASH_URL"))
		log.Printf("LOGSTASH_URL: [%s]", os.Getenv("LOGSTASH_URL"))
		if err == nil {
			log.Println("Connected to Logstash")
			return zerolog.New(conn).Level(logLevel).With().Str("service", serviceName).Timestamp().Logger()
		}
		time.Sleep(time.Second)
	}
	log.Println("Failed to connect to Logstash", err)
	return zerolog.New(os.Stdout).Level(logLevel).With().Str("service", serviceName).Timestamp().Logger()
}
