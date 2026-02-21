.PHONY: help build test install clean lint fmt coverage vendor setup-hooks

# Build variables
BASE_VERSION := $(shell cat VERSION 2>/dev/null || echo "0.0.0")
GIT_TAG := $(shell git describe --exact-match --tags 2>/dev/null)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version | awk '{print $$3}')

# If building from a git tag, use it. Otherwise append -dev+timestamp
ifeq ($(GIT_TAG),)
	VERSION := $(BASE_VERSION)-dev+$(shell date -u +"%Y%m%d%H%M%S")
else
	VERSION := $(GIT_TAG)
endif

LDFLAGS := -X 'github.com/fgrehm/cartage/cli.version=$(VERSION)' \
           -X 'github.com/fgrehm/cartage/cli.commit=$(COMMIT)' \
           -X 'github.com/fgrehm/cartage/cli.date=$(DATE)' \
           -X 'github.com/fgrehm/cartage/cli.goVersion=$(GO_VERSION)'

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

setup-hooks: ## Configure git hooks
	@git config core.hooksPath .githooks
	@chmod +x .githooks/*
	@echo "✓ Git hooks configured"

build: ## Build the cartage binary
	@echo "Building cartage..."
	@mkdir -p dist
	@go build -ldflags "$(LDFLAGS)" -o dist/cartage ./cmd/cartage
	@echo "✓ Built to dist/cartage"

test: ## Run tests
	@go test -race -shuffle=on ./...

install: build ## Install cartage to ~/.local/bin
	@echo "Installing to ~/.local/bin..."
	@mkdir -p ~/.local/bin
	@if [ -L ~/.local/bin/cartage ]; then \
		echo "✓ Already installed as symlink (rebuilt binary at dist/cartage)"; \
	elif [ -e ~/.local/bin/cartage ]; then \
		rm -f ~/.local/bin/cartage; \
		cp dist/cartage ~/.local/bin/cartage; \
		echo "✓ Replaced existing file and installed to ~/.local/bin/cartage"; \
	else \
		cp dist/cartage ~/.local/bin/cartage; \
		echo "✓ Installed to ~/.local/bin/cartage"; \
	fi

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -rf dist/
	@rm -f *.test *.out coverage.txt coverage.html
	@find . -name "*.test" -delete
	@echo "✓ Cleaned"

lint: ## Run golangci-lint
	@echo "Running linter..."
	@go tool golangci-lint run ./... && echo "✓ Lint passed"

fmt: ## Format code with gofumpt and goimports
	@echo "Formatting code..."
	@go tool golangci-lint fmt ./...
	@echo "✓ Formatted"

coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	@go test -race -shuffle=on -coverprofile=coverage.txt ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

vendor: ## Update vendored dependencies
	@echo "Vendoring dependencies..."
	@go mod tidy
	@go mod vendor
	@echo "✓ Dependencies vendored"
