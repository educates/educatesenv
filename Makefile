APP_NAME=educatesenv
MAIN_SRC=./cmd/educatesenv/main.go
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X github.com/educates/educatesenv/pkg/version.Version=$(VERSION)"
CONFIG_DIR=$(HOME)/.educatesenv

.PHONY: all
all: build

.PHONY: build
build:
	@echo "Building..."
	@mkdir -p bin/
	@go build $(LDFLAGS) -o bin/$(APP_NAME) $(MAIN_SRC)

.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf bin/

.PHONY: install
install: build
	@echo "Installing..."
	@mkdir -p $(CONFIG_DIR)/bin
	cp bin/$(APP_NAME) $(CONFIG_DIR)/bin/

.PHONY: lint
lint:
	@echo "Running linter..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Error: golangci-lint is not installed. Install it from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi
	@golangci-lint version
	@golangci-lint run ./...

.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all     - Build everything (default)"
	@echo "  build   - Build the binary"
	@echo "  clean   - Remove build artifacts"
	@echo "  install - Install binary to $(CONFIG_DIR)/bin"
	@echo "  lint    - Run golangci-lint"
	@echo "  test    - Run tests"
	@echo "  help    - Show this help message" 