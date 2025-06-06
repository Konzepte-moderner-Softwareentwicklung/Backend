# cmd/gateway/Dockerfile

FROM golang:1.24.3-alpine AS builder
WORKDIR /app

COPY . .
RUN go build -o /bin/tracking-service cmd/tracking-service/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /bin/tracking-service .
COPY .env .env
EXPOSE 8080
CMD ["./tracking-service", "-nats", "nats://nats:4222", "-verbose", "offer-url", "http://angebot-service:8080/"]
