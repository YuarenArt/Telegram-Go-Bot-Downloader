BINARY_NAME=dowload_bot.exe

all: build

build:
	go build -o $(BINARY_NAME) main.go

run:
	go run main.go

clean:
	go clean
	rm -f $(BINARY_NAME)

test:
	go test -v ./...

lint:
	golint ./...

.PHONY: all build run clean test lint