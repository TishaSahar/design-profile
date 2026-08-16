BINARY    := server
MAIN      := ./cmd/main/main.go
CONFIG    := ./config/config.yaml
SWAG_OUT  := ./docs

.PHONY: all run build test swagger lint tidy docker-build docker-run help

all: swagger build

## run: start the server (requires a running PostgreSQL)
run:
	go run $(MAIN) -config $(CONFIG)

## build: compile the server binary
build: swagger
	go build -o $(BINARY) $(MAIN)

## test: run all unit tests with coverage
test:
	go test ./... -v -cover -count=1

## swagger: regenerate OpenAPI docs from handler annotations
swagger:
	@command -v swag >/dev/null 2>&1 || { \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	}
	swag init -g $(MAIN) -o $(SWAG_OUT)

## tidy: clean up go.mod and go.sum
tidy:
	go mod tidy

## lint: run golangci-lint (install separately: https://golangci-lint.run/usage/install)
lint:
	golangci-lint run ./...

## docker-build: build the Docker image
docker-build:
	docker build -t design-profile-backend -f docker/Dockerfile .

## docker-run: run the container (configure env vars as needed)
docker-run:
	docker run --rm -p 8080:8080 \
		-v $(PWD)/config:/app/config \
		design-profile-backend

## help: display this help message
help:
	@grep -E '^##' Makefile | sed 's/## //'
