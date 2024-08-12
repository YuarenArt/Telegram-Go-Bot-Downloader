
ARG GO_VERSION=1.21
ARG ALPINE_VERSION=latest


FROM golang:${GO_VERSION}-alpine as builder

RUN apk add --no-cache git ffmpeg

WORKDIR /usr/src/telegram-bot

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /main .


FROM alpine:${ALPINE_VERSION}

RUN apk --no-cache add ca-certificates ffmpeg

WORKDIR /root/

COPY --from=builder /main .
COPY config.yaml .
COPY .env .env

EXPOSE 8080/tcp

CMD ["./main"]
