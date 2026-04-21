# Simple Makefile for a Go project

LIBOQS_GO_MOD := $(shell go list -m -f '{{.Path}}@{{.Version}}' github.com/open-quantum-safe/liboqs-go 2>/dev/null)
LIBOQS_GO_PKGCONFIG_DIR := $(shell go env GOPATH)/pkg/mod/$(LIBOQS_GO_MOD)/.config
LIBOQS_ENV = PKG_CONFIG_PATH="$(LIBOQS_GO_PKGCONFIG_DIR):$$PKG_CONFIG_PATH" LD_LIBRARY_PATH="/usr/local/lib:$$LD_LIBRARY_PATH"

# Build the application
all: build test

build:
	@echo "Building..."
	
	
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Run the application with liboqs enabled
run-liboqs:
	@$(LIBOQS_ENV) go run -tags liboqs cmd/api/main.go

# Test the application with liboqs enabled
test-liboqs:
	@echo "Testing (liboqs)..."
	@$(LIBOQS_ENV) go test -tags liboqs ./... -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run test run-liboqs test-liboqs clean watch
