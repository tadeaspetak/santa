APP_NAME := santa
VERSION := v1.0.1
# note: get the variable name using go tool nm <your binary> | grep <your variable>
VERSION_LDFLAG = -ldflags "-X github.com/tadeaspetak/santa/cmd/version.Version=$(VERSION)"

.PHONY: all run build binaries test fmt lint clean tools

all: run

run:
	go run $(VERSION_LDFLAG) . $(PARAM)

test:
	go test ./...
	rm -rf santa-batch-*

vet:
	go vet ./...

fmt: vet
	go fmt ./...

lint: fmt
	staticcheck ./...

build: lint binaries

# the release workflow runs this directly, CI covers the lint
binaries:
	env GOOS=darwin GOARCH=amd64 go build $(VERSION_LDFLAG) -o ./bin/santa-mac .
	env GOOS=windows GOARCH=amd64 go build $(VERSION_LDFLAG) -o ./bin/santa-win.exe .

clean:
	rm -rf ./bin
	rm -rf santa-batch-*

tools:
	go install honnef.co/go/tools/cmd/staticcheck@2024.1.1
	go mod tidy
