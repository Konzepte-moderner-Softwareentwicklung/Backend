# cmd/gateway/Dockerfile

FROM golang:1.24.3-alpine AS builder
WORKDIR /app

COPY . .
RUN go build -o /bin/media-service cmd/media-service/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /bin/media-service .
EXPOSE 8080
CMD ["./media-service", "-port", "8080", "-verbose"]
