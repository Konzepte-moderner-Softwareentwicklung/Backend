# cmd/gateway/Dockerfile

FROM golang:1.24.3-alpine AS builder
WORKDIR /app

COPY . .
RUN go build -o /bin/gateway cmd/gateway/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /bin/gateway .
COPY .env .env
EXPOSE 8080
CMD ["./gateway", "-port", "8080", "-nats", "nats://nats:4222", "-verbose"]
