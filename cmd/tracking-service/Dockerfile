# cmd/gateway/Dockerfile

FROM golang:1.24.3-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY . .

RUN go generate ./...
RUN go build -o /bin/tracking-service cmd/tracking-service/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /bin/tracking-service .
COPY .env .env
EXPOSE 8080
CMD ["./tracking-service"]
