ARG GO_VERSION=1.21
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

RUN apk --no-cache add ca-certificates ffmpeg

WORKDIR /root/

COPY --from=builder /telegram-bot .
COPY --from=builder /user-database .
COPY cmd/locales cmd/locales

# Create dummy certificates for development
RUN echo "-----BEGIN CERTIFICATE-----" > /usr/local/share/ca-certificates/tg-database.crt && \
    echo "-----END CERTIFICATE-----" >> /usr/local/share/ca-certificates/tg-database.crt && \
    echo "-----BEGIN PRIVATE KEY-----" > key.pem && \
    echo "-----END PRIVATE KEY-----" >> key.pem

RUN update-ca-certificates

# Copy env.example as .env (environment variables will be overridden by docker-compose)
COPY env.example .env

EXPOSE 8080/tcp
EXPOSE 8082/tcp

# Default command runs telegram bot
CMD ["./telegram-bot"]
