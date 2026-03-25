.PHONY: build test lint install clean

build:
	go build -o bin/pipedrive-cli ./cmd/pipedrive-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/pipedrive-cli

clean:
	rm -rf bin/
