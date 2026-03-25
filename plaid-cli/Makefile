.PHONY: build test lint install clean

build:
	go build -o bin/plaid-cli ./cmd/plaid-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/plaid-cli

clean:
	rm -rf bin/
