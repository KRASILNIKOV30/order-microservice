all: build test check

modules:
	go mod tidy

build: modules
	go build -o bin/hello ./cmd/hello/main.go

test:
	go test ./...

check:
	golangci-lint run
