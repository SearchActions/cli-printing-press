.PHONY: build test lint install clean

build:
	go build -o bin/linear-cli ./cmd/linear-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/linear-cli

clean:
	rm -rf bin/
