.PHONY: help dev-up dev-down dev-restart dev-logs dev-shell dev-status dev-clean
.PHONY: dev-db-reset dev-db-backup dev-install-xray
.PHONY: prod-build prod-up prod-down prod-logs
.PHONY: test test-verbose lint format
.PHONY: clean deps

# Default target
.DEFAULT_GOAL := help

# Colors
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

## help: Show this help message
help:
	@echo "$(BLUE)X-UI Development Makefile$(NC)"
	@echo ""
	@echo "$(GREEN)Development Commands:$(NC)"
	@echo "  make dev-up          - Start development environment with hot-reload"
	@echo "  make dev-down        - Stop development environment"
	@echo "  make dev-restart     - Restart development environment"
	@echo "  make dev-logs        - View development container logs"
	@echo "  make dev-shell       - Enter development container shell"
	@echo "  make dev-status      - Show development environment status"
	@echo "  make dev-clean       - Clean build cache and temporary files"
	@echo ""
	@echo "$(GREEN)Database Commands:$(NC)"
	@echo "  make dev-db-reset    - Reset database (deletes all data!)"
	@echo "  make dev-db-backup   - Backup database"
	@echo ""
	@echo "$(GREEN)Setup Commands:$(NC)"
	@echo "  make dev-install-xray - Install xray binary in container"
	@echo "  make deps            - Install Go dependencies"
	@echo ""
	@echo "$(GREEN)Production Commands:$(NC)"
	@echo "  make prod-build      - Build production Docker image"
	@echo "  make prod-up         - Start production environment"
	@echo "  make prod-down       - Stop production environment"
	@echo "  make prod-logs       - View production container logs"
	@echo ""
	@echo "$(GREEN)Code Quality:$(NC)"
	@echo "  make test            - Run tests"
	@echo "  make test-verbose    - Run tests with verbose output"
	@echo "  make lint            - Run linter"
	@echo "  make format          - Format code"
	@echo "  make clean           - Clean build artifacts"
	@echo ""
	@echo "For detailed documentation, see $(YELLOW)DEVELOPMENT.md$(NC)"

## dev-up: Start development environment
dev-up:
	@echo "$(BLUE)Starting development environment...$(NC)"
	@./dev.sh up

## dev-down: Stop development environment
dev-down:
	@echo "$(BLUE)Stopping development environment...$(NC)"
	@./dev.sh down

## dev-restart: Restart development environment
dev-restart:
	@echo "$(BLUE)Restarting development environment...$(NC)"
	@./dev.sh restart

## dev-logs: View development logs
dev-logs:
	@./dev.sh logs

## dev-shell: Enter development container shell
dev-shell:
	@./dev.sh shell

## dev-status: Show development environment status
dev-status:
	@./dev.sh status

## dev-clean: Clean build cache
dev-clean:
	@./dev.sh clean

## dev-db-reset: Reset database
dev-db-reset:
	@./dev.sh db-reset

## dev-db-backup: Backup database
dev-db-backup:
	@./dev.sh db-backup

## dev-install-xray: Install xray binary
dev-install-xray:
	@./dev.sh install-xray

## prod-build: Build production image
prod-build:
	@echo "$(BLUE)Building production Docker image...$(NC)"
	@docker compose build

## prod-up: Start production environment
prod-up:
	@echo "$(BLUE)Starting production environment...$(NC)"
	@docker compose up -d
	@echo "$(GREEN)✓$(NC) Production environment started"
	@echo "$(BLUE)Access at: http://localhost:54321$(NC)"

## prod-down: Stop production environment
prod-down:
	@echo "$(BLUE)Stopping production environment...$(NC)"
	@docker compose down

## prod-logs: View production logs
prod-logs:
	@docker compose logs -f

## test: Run tests
test:
	@echo "$(BLUE)Running tests...$(NC)"
	@go test -race ./...

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "$(BLUE)Running tests (verbose)...$(NC)"
	@go test -v -race ./...

## lint: Run linter
lint:
	@echo "$(BLUE)Running linter...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Install: https://golangci-lint.run/usage/install/$(NC)"; \
		exit 1; \
	fi

## format: Format code
format:
	@echo "$(BLUE)Formatting code...$(NC)"
	@gofmt -w -s .
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "$(YELLOW)goimports not installed. Run: go install golang.org/x/tools/cmd/goimports@latest$(NC)"; \
	fi
	@echo "$(GREEN)✓$(NC) Code formatted"

## deps: Install Go dependencies
deps:
	@echo "$(BLUE)Installing Go dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✓$(NC) Dependencies installed"

## clean: Clean build artifacts
clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf tmp/ build/ bin/x-ui dist/
	@go clean -cache -testcache -modcache
	@echo "$(GREEN)✓$(NC) Clean complete"

# Quick commands (aliases)
up: dev-up
down: dev-down
logs: dev-logs
restart: dev-restart
shell: dev-shell
status: dev-status
