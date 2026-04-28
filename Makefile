.DEFAULT_GOAL := help

BINARY_NAME := avms-api
DOCKER_IMAGE := avms

LIBOQS_GO_MOD := $(shell go list -m -f '{{.Path}}@{{.Version}}' github.com/open-quantum-safe/liboqs-go 2>/dev/null)
LIBOQS_GO_PKGCONFIG_DIR := $(shell go env GOPATH)/pkg/mod/$(LIBOQS_GO_MOD)/.config
LIBOQS_ENV = PKG_CONFIG_PATH="$(LIBOQS_GO_PKGCONFIG_DIR):$$PKG_CONFIG_PATH" LD_LIBRARY_PATH="/usr/local/lib:$$LD_LIBRARY_PATH"

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: build test ## Build and test everything

.PHONY: build
build: ## Build the Go backend binary
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd/api

.PHONY: run
run: ## Run the Go backend in development mode
	@go run ./cmd/api

.PHONY: test
test: ## Run Go tests
	@echo "Testing..."
	@go test ./... -v

.PHONY: run-liboqs
run-liboqs: ## Run with liboqs (PQC) support enabled
	@$(LIBOQS_ENV) go run -tags liboqs ./cmd/api

.PHONY: test-liboqs
test-liboqs: ## Test with liboqs (PQC) support enabled
	@echo "Testing (liboqs)..."
	@$(LIBOQS_ENV) go test -tags liboqs ./... -v

.PHONY: clean
clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME) main
	@cd frontend && vp exec rimraf dist 2>/dev/null || rm -rf dist

.PHONY: watch
watch: ## Run with live reload via air
	@if command -v air > /dev/null; then \
		air; \
	else \
		read -p "Go's 'air' is not installed. Install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/air-verse/air@latest; \
			air; \
		else \
			echo "Aborting."; \
			exit 1; \
		fi; \
	fi

# ------------------------------------------------------------------------------
# Frontend
# ------------------------------------------------------------------------------

.PHONY: frontend-check
frontend-check: ## Run frontend lint, format, and type checks
	@cd frontend && vp check

.PHONY: frontend-test
frontend-test: ## Run frontend tests
	@cd frontend && vp test

.PHONY: frontend-build
frontend-build: ## Build frontend for production
	@cd frontend && vp run build

.PHONY: frontend-dev
frontend-dev: ## Run frontend development server
	@cd frontend && vp dev

# ------------------------------------------------------------------------------
# Docker
# ------------------------------------------------------------------------------

.PHONY: docker-build
docker-build: ## Build Docker image
	@docker build -t $(DOCKER_IMAGE):latest .

.PHONY: docker-run
docker-run: ## Run Docker container (builds image first)
	@docker run --rm -p 8080:8080 -v "$(PWD)/avms.db:/app/avms.db" $(DOCKER_IMAGE):latest

.PHONY: docker-stop
docker-stop: ## Stop running Docker container
	@docker stop $(DOCKER_IMAGE) 2>/dev/null || true
