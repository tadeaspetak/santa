# variables
APP_NAME := secret-reindeer

.PHONY: all run build test fmt lint clean tools

all: run

run:
	go run .

test:
	go test ./...

fmt:
	go fmt ./...

lint: fmt
	make tools
	go vet ./...
	staticcheck ./...

build-cli: lint
	go build -o ./bin/$(APP_NAME) .

build-local: lint
  go build -o ./$(APP_NAME) .

clean:
	rm -rf ./bin/$(APP_NAME)

tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go mod tidy
