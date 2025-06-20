# cmd/gateway/Dockerfile

FROM golang:1.24.3-alpine AS builder
WORKDIR /app

COPY . .
RUN cd cmd/gateway && go run github.com/swaggo/swag/cmd/swag@latest init --pdl 2
RUN go build -o /bin/gateway cmd/gateway/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /bin/gateway .
COPY .env .env
EXPOSE 8080
CMD ["./gateway"]
