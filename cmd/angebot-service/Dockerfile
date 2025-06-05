# cmd/angebot-service/Dockerfile

FROM golang:1.24.3-alpine AS builder
WORKDIR /app

COPY . .
RUN go build -o /bin/angebot-service cmd/angebot-service/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /bin/angebot-service .
COPY .env .env
EXPOSE 8080
CMD ["./angebot-service", "-port", "8080", "-verbose"]
