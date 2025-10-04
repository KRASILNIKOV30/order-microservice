all: build test check

modules:
	go mod tidy

build: modules
	go build main.go

test:
	go test ./...

check:
	golangci-lint run
