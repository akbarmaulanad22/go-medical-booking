.PHONY: run build test clean migrate-up migrate-down migrate-create lint help

# Variables
APP_NAME=go-template-clean-architecture
MAIN_PATH=./cmd/main.go
BUILD_DIR=./build
MIGRATIONS_DIR=./migrations

# Database
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= clean_architecture
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Colors
GREEN=\033[0;32m
NC=\033[0m

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  run              Run the application"
	@echo "  build            Build the application binary"
	@echo "  test             Run all tests"
	@echo "  clean            Remove build artifacts"
	@echo "  migrate-up       Run database migrations"
	@echo "  migrate-down     Rollback database migrations"
	@echo "  migrate-create   Create a new migration (usage: make migrate-create name=migration_name)"
	@echo "  lint             Run linter"
	@echo "  tidy             Run go mod tidy"
	@echo "  deps             Download dependencies"

## run: Run the application
run:
	@echo "$(GREEN)Starting application...$(NC)"
	go run $(MAIN_PATH)

## build: Build the application binary
build:
	@echo "$(GREEN)Building application...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(APP_NAME)$(NC)"

## test: Run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v -cover ./...

## clean: Remove build artifacts
clean:
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)Clean complete$(NC)"

## migrate-up: Run database migrations
migrate-up:
	@echo "$(GREEN)Running migrations...$(NC)"
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@echo "$(GREEN)Migrations complete$(NC)"

## migrate-down: Rollback database migrations
migrate-down:
	@echo "$(GREEN)Rolling back migrations...$(NC)"
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down
	@echo "$(GREEN)Rollback complete$(NC)"

## migrate-down-one: Rollback one migration
migrate-down-one:
	@echo "$(GREEN)Rolling back one migration...$(NC)"
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1
	@echo "$(GREEN)Rollback complete$(NC)"

## migrate-create: Create a new migration
migrate-create:
ifndef name
	$(error name is required. Usage: make migrate-create name=migration_name)
endif
	@echo "$(GREEN)Creating migration: $(name)$(NC)"
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
	@echo "$(GREEN)Migration created$(NC)"

## migrate-force: Force migration version
migrate-force:
ifndef version
	$(error version is required. Usage: make migrate-force version=1)
endif
	@echo "$(GREEN)Forcing migration version: $(version)$(NC)"
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(version)
	@echo "$(GREEN)Migration force complete$(NC)"

## lint: Run linter
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	golangci-lint run ./...

## tidy: Run go mod tidy
tidy:
	@echo "$(GREEN)Running go mod tidy...$(NC)"
	go mod tidy

## deps: Download dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	go mod download
