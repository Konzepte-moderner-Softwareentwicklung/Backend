FROM golang:1.24.3-alpine3.21 AS builder
WORKDIR /app
COPY . .

RUN go build -o bin/gateway main.go


FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/bin/gateway .
EXPOSE 8080
CMD ["./gateway", "-port", "8080", "-nats", "nats://nats:4222"]