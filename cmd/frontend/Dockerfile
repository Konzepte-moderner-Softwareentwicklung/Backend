# cmd/angebot-service/Dockerfile

FROM golang:1.24.3-alpine 
WORKDIR /app

COPY . .
RUN go build cmd/frontend/main.go

EXPOSE 8080
COPY .env .env
CMD ["./main"]
