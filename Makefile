# variables
APP_NAME := secret-reindeer

.PHONY: all run build test fmt lint clean tools

all: run

run:
	go run ./...

test:
	go test ./...

fmt:
	go fmt ./...

lint: fmt
  staticcheck ./...
	go vet ./...

# TODO
build: lint
	go build -o ./bin/$(APP_NAME) ./cmd

# TODO
clean:
	rm -rf ./bin/$(APP_NAME)

tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go mod tidy
