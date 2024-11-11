APP_NAME := secret-reindeer
VERSION := v0.0.2
# note: get the variable name using go tool nm <your binary> | grep <your variable>
VERSION_LDFLAG = -ldflags "-X github.com/tadeaspetak/secret-reindeer/cmd/version.Version=$(VERSION)"

.PHONY: all run build test fmt lint clean tools

all: run

run:
	go run $(VERSION_LDFLAG) . $(PARAM)

test:
	go test ./...

fmt:
	go fmt ./...

lint: fmt
	make tools
	go vet ./...
	staticcheck ./...

build: lint
	env GOOS=darwin GOARCH=amd64 go build $(VERSION_LDFLAG) -o ./bin/ .
	env GOOS=windows GOARCH=amd64 go build $(VERSION_LDFLAG) -o ./bin/ .

clean:
	rm -rf ./bin/$(APP_NAME)

tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go mod tidy
