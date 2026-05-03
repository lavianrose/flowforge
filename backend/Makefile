.PHONY: build run clean deps test

build:
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api/main.go

clean:
	rm -rf bin

deps:
	go mod tidy

test:
	go test -v -race ./...

test-integration:
	go test -v -count=1 -timeout 120s ./tests/integration/
