.PHONY: build test lint install clean

build:
	go build -o bin/discord-cli ./cmd/discord-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/discord-cli

clean:
	rm -rf bin/
