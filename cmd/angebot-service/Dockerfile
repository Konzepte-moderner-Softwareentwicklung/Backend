# cmd/angebot-service/Dockerfile

FROM golang:1.24.3-alpine AS builder
RUN apk add --no-cache git

WORKDIR /app

COPY . .
RUN cd cmd/angebot-service && go run github.com/swaggo/swag/cmd/swag@latest init --pdl 2
RUN go generate ./...
RUN go build -o /bin/angebot-service cmd/angebot-service/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /bin/angebot-service .
COPY .env .env
EXPOSE 8080
CMD ["./angebot-service"]
