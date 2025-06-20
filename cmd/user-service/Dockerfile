# cmd/user-service/Dockerfile

FROM golang:1.24.3-alpine AS builder
WORKDIR /app

COPY . .
RUN cd cmd/user-service && go run github.com/swaggo/swag/cmd/swag@latest init --pdl 2
RUN go build -o /bin/user-service cmd/user-service/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /bin/user-service .
EXPOSE 8080
COPY .env .env
CMD ["./user-service"]
