FROM golang:1.24.3-alpine AS builder
WORKDIR /app

COPY . .
RUN go build -o /bin/rating-service cmd/rating-service/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /bin/rating-service .
EXPOSE 8080
CMD ["./rating-service", "-port", "8080", "-nats", "nats://nats:4222", "-mongo", "mongodb://mongo:27017", "-verbose"]

