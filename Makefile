APP_NAME := santa
VERSION := v0.0.4
# note: get the variable name using go tool nm <your binary> | grep <your variable>
VERSION_LDFLAG = -ldflags "-X github.com/tadeaspetak/santa/cmd/version.Version=$(VERSION)"

.PHONY: all run build test fmt lint clean tools

all: run

run:
	go run $(VERSION_LDFLAG) . $(PARAM)

test:
	go test ./...
	rm -f santa-batch-*
	rm -f internal/app/santa-batch-*

vet:
	go vet ./...

fmt: vet
	go fmt ./...

lint: fmt
	staticcheck ./...

build: lint
	env GOOS=darwin GOARCH=amd64 go build $(VERSION_LDFLAG) -o ./bin/santa-mac .
	env GOOS=windows GOARCH=amd64 go build $(VERSION_LDFLAG) -o ./bin/santa-win.exe .

clean:
	rm -rf ./bin
	rm -f santa-batch-*
	rm -f internal/app/santa-batch-*

tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go mod tidy
