.PHONY: build, clean, deps, format, lint, race, release, run, test

build:
	go mod tidy
	CGO_ENABLED=0 go build -o ./bin/deordie-bot ./app

format:
	go fmt ./...

lint:
	golangci-lint run -v

test:
	go test -v ./...

race:
	go test -race ./...

clean:
	go clean -testcache

release: format lint test race build

run:
	go run ./app

deps:
	go get -u github.com/joho/godotenv
	go get -u gopkg.in/telebot.v3
	go get -u github.com/google/go-github/v57
	go get -u github.com/stretchr/testify
