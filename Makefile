# ============================================================================
# Synapse: Doc-First Event Processing
# ============================================================================
# A masterful Makefile for spec-driven development
#
# Usage:
#   make help          Show all available targets
#   make setup         One-time setup for new clones
#   make generate      Regenerate code from specs
#   make test          Run all tests
#   make run           Start the server
# ============================================================================

.PHONY: help setup generate build test test-short test-conformance test-pipeline \
        run clean lint fmt vet validate-specs diagrams docker-up docker-down \
        deps tidy coverage benchmark

# Colors for pretty output
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m
BOLD := \033[1m

# ============================================================================
# HELP
# ============================================================================

help: ## Show this help message
	@echo ""
	@echo "$(BOLD)Synapse: Doc-First Event Processing$(RESET)"
	@echo "======================================"
	@echo ""
	@echo "$(CYAN)Quick Start:$(RESET)"
	@echo "  make setup      → First-time setup (deps + generate)"
	@echo "  make test       → Run all tests"
	@echo "  make run        → Start the server"
	@echo ""
	@echo "$(CYAN)Available targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-18s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ============================================================================
# SETUP & DEPENDENCIES
# ============================================================================

setup: deps generate ## One-time setup for new clones
	@echo "$(GREEN)✓ Setup complete!$(RESET)"
	@echo ""
	@echo "Next steps:"
	@echo "  make test    → Run tests"
	@echo "  make run     → Start server"

deps: ## Download Go dependencies
	@echo "$(CYAN)→ Downloading dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies ready$(RESET)"

tidy: ## Tidy go.mod and go.sum
	@go mod tidy

# ============================================================================
# CODE GENERATION
# ============================================================================

generate: ## Regenerate code from OpenAPI & AsyncAPI specs
	@echo "$(CYAN)→ Generating code from specs...$(RESET)"
	@go run ./cmd/synctl
	@$(MAKE) fmt-generated
	@echo "$(GREEN)✓ Code generated$(RESET)"

fmt-generated: ## Format generated code
	@gofmt -w ./internal/generated/

# ============================================================================
# BUILD
# ============================================================================

build: ## Build the synapse binary
	@echo "$(CYAN)→ Building synapse...$(RESET)"
	@go build -o bin/synapse ./cmd/synapse
	@echo "$(GREEN)✓ Built: bin/synapse$(RESET)"

build-synctl: ## Build the code generator
	@echo "$(CYAN)→ Building synctl...$(RESET)"
	@go build -o bin/synctl ./cmd/synctl
	@echo "$(GREEN)✓ Built: bin/synctl$(RESET)"

build-all: build build-synctl ## Build all binaries

# ============================================================================
# TESTING
# ============================================================================

test: ## Run all tests (requires Docker)
	@echo "$(CYAN)→ Running all tests...$(RESET)"
	@go test ./... -v -count=1
	@echo "$(GREEN)✓ All tests passed$(RESET)"

test-short: ## Run fast tests only (no Docker)
	@echo "$(CYAN)→ Running short tests...$(RESET)"
	@go test ./... -short -v
	@echo "$(GREEN)✓ Short tests passed$(RESET)"

test-conformance: ## Run conformance tests only
	@echo "$(CYAN)→ Running conformance tests...$(RESET)"
	@go test ./internal/conformance/... -v -count=1
	@echo "$(GREEN)✓ Conformance tests passed$(RESET)"

test-pipeline: ## Run pipeline integration tests
	@echo "$(CYAN)→ Running pipeline tests...$(RESET)"
	@go test ./internal/pipeline/... -v -count=1
	@echo "$(GREEN)✓ Pipeline tests passed$(RESET)"

coverage: ## Run tests with coverage report
	@echo "$(CYAN)→ Running tests with coverage...$(RESET)"
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(RESET)"

benchmark: ## Run benchmarks
	@echo "$(CYAN)→ Running benchmarks...$(RESET)"
	@go test ./... -bench=. -benchmem

# ============================================================================
# CODE QUALITY
# ============================================================================

lint: fmt vet ## Run all linters (fmt + vet)
	@echo "$(GREEN)✓ Lint complete$(RESET)"

fmt: ## Format all Go code
	@echo "$(CYAN)→ Formatting code...$(RESET)"
	@gofmt -w .

vet: ## Run go vet
	@echo "$(CYAN)→ Running go vet...$(RESET)"
	@go vet ./...

check: lint test-short ## Quick check (lint + short tests)
	@echo "$(GREEN)✓ Quick check passed$(RESET)"

# ============================================================================
# SPECIFICATION VALIDATION
# ============================================================================

validate-specs: validate-openapi validate-asyncapi ## Validate all specs

validate-openapi: ## Validate OpenAPI specification
	@echo "$(CYAN)→ Validating OpenAPI spec...$(RESET)"
	@if command -v vacuum > /dev/null; then \
		vacuum lint openapi/openapi.yaml; \
	else \
		echo "$(YELLOW)⚠ vacuum not installed, skipping OpenAPI validation$(RESET)"; \
		echo "  Install: go install github.com/daveshanley/vacuum@latest"; \
	fi

validate-asyncapi: ## Validate AsyncAPI specification
	@echo "$(CYAN)→ Validating AsyncAPI spec...$(RESET)"
	@if command -v asyncapi > /dev/null; then \
		asyncapi validate asyncapi/asyncapi.yaml; \
	else \
		echo "$(YELLOW)⚠ asyncapi CLI not installed, skipping AsyncAPI validation$(RESET)"; \
		echo "  Install: npm install -g @asyncapi/cli"; \
	fi

# ============================================================================
# RUN
# ============================================================================

run: ## Run the synapse server
	@echo "$(CYAN)→ Starting Synapse server...$(RESET)"
	@go run ./cmd/synapse

run-dev: ## Run with hot reload (requires air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(YELLOW)⚠ air not installed$(RESET)"; \
		echo "  Install: go install github.com/cosmtrek/air@latest"; \
		$(MAKE) run; \
	fi

# ============================================================================
# DIAGRAMS
# ============================================================================

diagrams: ## Generate architecture diagrams
	@echo "$(CYAN)→ Generating diagrams...$(RESET)"
	@cd scripts && \
		if [ ! -d "venv" ]; then \
			python3 -m venv venv && \
			./venv/bin/pip install -q diagrams; \
		fi && \
		./venv/bin/python generate_all.py
	@echo "$(GREEN)✓ Diagrams generated in scripts/output/$(RESET)"

# ============================================================================
# DOCKER (for local infrastructure)
# ============================================================================

docker-up: ## Start local infrastructure (NATS, Postgres, Redis)
	@echo "$(CYAN)→ Starting infrastructure...$(RESET)"
	@docker compose up -d
	@echo "$(GREEN)✓ Infrastructure running$(RESET)"
	@echo "  NATS:     nats://localhost:4222"
	@echo "  Postgres: postgres://localhost:5432"
	@echo "  Redis:    redis://localhost:6379"

docker-down: ## Stop local infrastructure
	@echo "$(CYAN)→ Stopping infrastructure...$(RESET)"
	@docker compose down
	@echo "$(GREEN)✓ Infrastructure stopped$(RESET)"

docker-logs: ## Show infrastructure logs
	@docker compose logs -f

# ============================================================================
# CLEANUP
# ============================================================================

clean: ## Clean build artifacts
	@echo "$(CYAN)→ Cleaning...$(RESET)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -rf scripts/venv/
	@echo "$(GREEN)✓ Clean$(RESET)"

clean-generated: ## Remove generated code (use with caution)
	@echo "$(YELLOW)→ Removing generated code...$(RESET)"
	@rm -f internal/generated/*.gen.go
	@echo "$(GREEN)✓ Generated code removed$(RESET)"

# ============================================================================
# WORKFLOW SHORTCUTS
# ============================================================================

all: setup test build ## Full build pipeline
	@echo "$(GREEN)✓ All done!$(RESET)"

dev: generate test-short run ## Development cycle: generate → test → run

ci: deps generate lint test ## CI pipeline
	@echo "$(GREEN)✓ CI passed$(RESET)"

# ============================================================================
# INFO
# ============================================================================

info: ## Show project information
	@echo ""
	@echo "$(BOLD)Synapse: Doc-First Event Processing$(RESET)"
	@echo "======================================"
	@echo ""
	@echo "$(CYAN)Specs:$(RESET)"
	@echo "  OpenAPI:  openapi/openapi.yaml"
	@echo "  AsyncAPI: asyncapi/asyncapi.yaml"
	@echo ""
	@echo "$(CYAN)Generated Code:$(RESET)"
	@ls -la internal/generated/*.gen.go 2>/dev/null || echo "  (not yet generated - run 'make generate')"
	@echo ""
	@echo "$(CYAN)Workflow:$(RESET)"
	@echo "  1. Edit specs (openapi/ or asyncapi/)"
	@echo "  2. make generate"
	@echo "  3. Implement handlers"
	@echo "  4. make test-conformance"
	@echo ""

.DEFAULT_GOAL := help
