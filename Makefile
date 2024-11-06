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

# TODO
build: lint
	go build -o ./bin/$(APP_NAME) .

# TODO
clean:
	rm -rf ./bin/$(APP_NAME)

tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go mod tidy
