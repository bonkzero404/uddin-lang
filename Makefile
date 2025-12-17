.PHONY: build install version tag-release test clean help

# Variables
BINARY_NAME=uddinlang
MAIN_PATH=./cmd/uddin-lang
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_TAG=$(shell git describe --exact-match --tags HEAD 2>/dev/null || echo "")

# Build flags
LDFLAGS=-ldflags "-X github.com/bonkzero404/uddin-lang/internal/version.Version=$(VERSION) \
                  -X github.com/bonkzero404/uddin-lang/internal/version.BuildTime=$(BUILD_TIME) \
                  -X github.com/bonkzero404/uddin-lang/internal/version.GitCommit=$(GIT_COMMIT) \
                  -X github.com/bonkzero404/uddin-lang/internal/version.GitTag=$(GIT_TAG)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: ./$(BINARY_NAME)"

install: ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME) version $(VERSION)..."
	@go install $(LDFLAGS) $(MAIN_PATH)
	@echo "Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)"

test: ## Run tests
	@go test ./... -v

clean: ## Clean build artifacts
	@rm -f $(BINARY_NAME)
	@echo "Cleaned build artifacts"

version: ## Show current version information
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Tag: $(GIT_TAG)"

tag-release: ## Create and push a new git tag (usage: make tag-release VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make tag-release VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Creating git tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "Tag $(VERSION) created. Push with: git push origin $(VERSION)"

release: ## Build and prepare for release
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make release VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Building release $(VERSION)..."
	@$(MAKE) build VERSION=$(VERSION)
	@echo "Release $(VERSION) built successfully"

# Go module commands
mod-tidy: ## Run go mod tidy
	@go mod tidy

mod-vendor: ## Create vendor directory
	@go mod vendor

mod-download: ## Download dependencies
	@go mod download

# Development commands
dev: ## Build and run in development mode
	@$(MAKE) build VERSION=dev
	@./$(BINARY_NAME) --version

