.PHONY: build, clean, deps, format, lint, race, release, run, test

build:
	go mod tidy
	go build -o ./bin/ .

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
	go run .

deps:
