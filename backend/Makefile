.PHONY: build run migrate seed clean

build:
	go build -o bin/api ./cmd/api
	go build -o bin/migrate ./cmd/migrate
	go build -o bin/seed ./cmd/seed

run:
	go run ./cmd/api/main.go

migrate-up:
	go run ./cmd/migrate/main.go up

migrate-down:
	go run ./cmd/migrate/main.go down

seed:
	go run ./cmd/seed/main.go

clean:
	rm -rf bin

deps:
	go mod tidy
