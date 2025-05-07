# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
BINARY_NAME=server
MAIN_PATH=cmd/api/main.go
AIR_VERSION=v1.49.0
GOPATH=$(shell go env GOPATH)
AIR=$(GOPATH)/bin/air
ENV_FILE=.env
MIGRATION_DIR=./migrations

# Docker parameters
DOCKER_COMPOSE=docker-compose
DOCKER_IMAGE_NAME=go-server-boilerplate
DOCKER_IMAGE_TAG=latest

.PHONY: all build run test clean deps fmt lint help docker-up docker-down docker-logs docker-build docker-clean install-air watch migrate-create migrate-up migrate-down docs

all: clean build

build: ## Build the application
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

run: ## Run the application
	$(GORUN) $(MAIN_PATH)

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

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

docker-up: ## Start Docker containers (PostgreSQL and Redis)
	$(DOCKER_COMPOSE) up -d --build postgres redis

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

docs: ## Generate API documentation
	@echo "Generating API documentation..."
	@if ! command -v swag > /dev/null; then \
		echo "Installing swag..." && \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	swag init -g $(MAIN_PATH) -o ./docs/swagger

migrate-create: ## Create a new migration (usage: make migrate-create name=migration_name)
	@echo "Creating migration files..."
	@if ! command -v migrate > /dev/null; then \
		echo "Installing migrate..." && \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(name)

migrate-up: ## Run migrations up
	@echo "Running migrations up..."
	@if ! command -v migrate > /dev/null; then \
		echo "Installing migrate..." && \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@if [ -f "$(ENV_FILE)" ]; then \
		migrate -database $$(grep DATABASE_URL $(ENV_FILE) | cut -d '=' -f2) -path $(MIGRATION_DIR) up; \
	else \
		echo "Error: $(ENV_FILE) file not found"; \
		exit 1; \
	fi

migrate-down: ## Run migrations down
	@echo "Running migrations down..."
	@if ! command -v migrate > /dev/null; then \
		echo "Installing migrate..." && \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@if [ -f "$(ENV_FILE)" ]; then \
		migrate -database $$(grep DATABASE_URL $(ENV_FILE) | cut -d '=' -f2) -path $(MIGRATION_DIR) down; \
	else \
		echo "Error: $(ENV_FILE) file not found"; \
		exit 1; \
	fi

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help 