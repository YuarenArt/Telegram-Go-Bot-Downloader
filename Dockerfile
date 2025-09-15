ARG GO_VERSION=1.23
ARG ALPINE_VERSION=latest

FROM golang:${GO_VERSION}-alpine as builder

RUN apk add --no-cache git ffmpeg

WORKDIR /usr/src/telegram-bot

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /telegram-bot cmd/telegram-bot/main.go
RUN go build -o /user-database cmd/user-database/main.go

FROM alpine:${ALPINE_VERSION}

RUN apk --no-cache add ca-certificates ffmpeg openssl

WORKDIR /root/

COPY --from=builder /telegram-bot .
COPY --from=builder /user-database .
COPY cmd/locales cmd/locales

# Create self-signed certificates for development
RUN openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

RUN update-ca-certificates

# Copy env.example as env.example (environment variables will be overridden by docker-compose)
COPY env. .env
COPY cookies.txt .

EXPOSE 8080/tcp
EXPOSE 8082/tcp

# Default command runs telegram bot
CMD ["./telegram-bot"]
