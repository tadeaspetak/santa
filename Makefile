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

build: lint
	env GOOS=darwin GOARCH=amd64 go build -o ./bin/ .
	env GOOS=windows GOARCH=amd64 go build -o ./bin/ .


clean:
	rm -rf ./bin/$(APP_NAME)

tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go mod tidy
