BINARY_TELEGRAM_BOT=telegram-bot.exe
BINARY_USER_DATABASE=user-database.exe

all: build

build:
	go build -o $(BINARY_TELEGRAM_BOT) cmd/telegram-bot/main.go
	go build -o $(BINARY_USER_DATABASE) cmd/user-database/main.go

build-telegram-bot:
	go build -o $(BINARY_TELEGRAM_BOT) cmd/telegram-bot/main.go

build-user-database:
	go build -o $(BINARY_USER_DATABASE) cmd/user-database/main.go

run-telegram-bot:
	go run cmd/telegram-bot/main.go

run-user-database:
	go run cmd/user-database/main.go

clean:
	go clean
	rm -f $(BINARY_TELEGRAM_BOT)
	rm -f $(BINARY_USER_DATABASE)

test:
	go test -v ./...

lint:
	golint ./...

docker-build:
	docker-compose build

docker-up:
	docker-compose up

docker-down:
	docker-compose down

swagger-db-api:
	swag init -g cmd/user-database/main.go -o docs/api

.PHONY: all build build-telegram-bot build-user-database run-telegram-bot run-user-database clean test lint docker-build docker-up docker-down