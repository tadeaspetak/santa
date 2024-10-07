# variables
APP_NAME := secret-reindeer

.PHONY: all run test fmt lint vet build clean

all: run

run:
	go run ./...

build: vet
	go build -o ./bin/$(APP_NAME) ./cmd/app

test:
	go test ./...

# format
fmt:
	go fmt ./...

# lint
lint: fmt
	golint ./...

# find suspicious constructs
vet: lint
	go vet ./...

clean:
	rm -rf ./bin/$(APP_NAME)
