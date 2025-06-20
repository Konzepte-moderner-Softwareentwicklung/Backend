# cmd/chat-service/Dockerfile

FROM golang:1.24.3-alpine AS builder
WORKDIR /app

COPY . .
RUN cd cmd/chat-service && go run github.com/swaggo/swag/cmd/swag@latest init
RUN go build -o /bin/chat-service cmd/chat-service/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /bin/chat-service .
COPY .env .env
EXPOSE 8080
CMD ["./chat-service"]
