# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
BINARY_NAME=server
MAIN_PATH=cmd/main.go
AIR_VERSION=v1.49.0
GOPATH=$(shell go env GOPATH)
AIR=$(GOPATH)/bin/air

# Docker parameters
DOCKER_COMPOSE=docker-compose
DOCKER_IMAGE_NAME=go-server
DOCKER_IMAGE_TAG=latest

.PHONY: all build run test clean deps fmt lint help docker-up docker-down docker-logs docker-build docker-clean install-air

all: clean build

build: ## Build the application
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

run: ## Run the application
	$(GORUN) $(MAIN_PATH)

test: ## Run tests
	$(GOTEST) -v ./...

clean: ## Clean build files
	rm -f $(BINARY_NAME)
	find . -type f -name '*.test' -delete
	find . -type f -name '*.out' -delete

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

fmt: ## Format code
	$(GOFMT) ./...

lint: ## Run linter
	golangci-lint run

docker-up: ## Start Docker containers
	$(DOCKER_COMPOSE) up -d --build

docker-down: ## Stop Docker containers
	$(DOCKER_COMPOSE) down

docker-logs: ## View Docker container logs
	$(DOCKER_COMPOSE) logs -f

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

docker-clean: ## Clean Docker resources
	$(DOCKER_COMPOSE) down -v --remove-orphans

install-air: ## Install air for hot reloading
	@if ! command -v air > /dev/null; then \
		echo "Installing air..." && \
		go install github.com/cosmtrek/air@$(AIR_VERSION); \
	fi

watch: install-air ## Run the application with hot reload
	$(AIR)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help 