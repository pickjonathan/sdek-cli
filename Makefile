.PHONY: build test run clean install uninstall lint help release build-all

# Version and build information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.gitCommit=$(GIT_COMMIT)"

# Build directories
DIST_DIR := dist
BIN_NAME := sdek

# Install location
INSTALL_PATH := $(shell go env GOPATH)/bin

# Targets
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the sdek binary
	@echo "Building sdek..."
	go build $(LDFLAGS) -o sdek main.go

test: ## Run all tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

run: build ## Build and run sdek
	./sdek

install: build ## Install sdek to $GOPATH/bin
	@echo "Installing sdek to $(INSTALL_PATH)..."
	@mkdir -p $(INSTALL_PATH)
	@cp $(BIN_NAME) $(INSTALL_PATH)/
	@echo "✓ sdek installed to $(INSTALL_PATH)/$(BIN_NAME)"

uninstall: ## Uninstall sdek from $GOPATH/bin
	@echo "Uninstalling sdek from $(INSTALL_PATH)..."
	@rm -f $(INSTALL_PATH)/$(BIN_NAME)
	@echo "✓ sdek uninstalled"

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -f $(BIN_NAME)
	@rm -f coverage.out coverage.html
	@rm -rf $(DIST_DIR)/
	@echo "✓ Clean complete"

lint: ## Run golangci-lint
	@echo "Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

# Cross-compilation targets
build-all: clean ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@$(MAKE) build-linux
	@$(MAKE) build-darwin
	@$(MAKE) build-windows
	@echo "✓ All builds complete in $(DIST_DIR)/"

build-linux: ## Build for Linux (amd64 and arm64)
	@echo "Building for Linux amd64..."
	@mkdir -p $(DIST_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BIN_NAME)-linux-amd64 main.go
	@echo "Building for Linux arm64..."
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BIN_NAME)-linux-arm64 main.go
	@echo "✓ Linux builds complete"

build-darwin: ## Build for macOS (amd64 and arm64)
	@echo "Building for macOS amd64..."
	@mkdir -p $(DIST_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BIN_NAME)-darwin-amd64 main.go
	@echo "Building for macOS arm64 (Apple Silicon)..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BIN_NAME)-darwin-arm64 main.go
	@echo "✓ macOS builds complete"

build-windows: ## Build for Windows (amd64)
	@echo "Building for Windows amd64..."
	@mkdir -p $(DIST_DIR)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BIN_NAME)-windows-amd64.exe main.go
	@echo "✓ Windows build complete"

# Release build with optimizations
release: clean test lint ## Build release binaries for all platforms
	@echo "Creating release build $(VERSION)..."
	@$(MAKE) build-all
	@cd $(DIST_DIR) && shasum -a 256 * > checksums.txt
	@echo "✓ Release build complete with checksums"
	@echo "Version: $(VERSION)"
	@echo "Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

# Development targets
dev: ## Build and run in development mode
	@echo "Running in development mode..."
	@go run main.go

watch: ## Watch for changes and rebuild (requires entr)
	@echo "Watching for changes..."
	@find . -name '*.go' | entr -c make build

# Show build info
info: ## Display build information
	@echo "Build Information:"
	@echo "  Version:    $(VERSION)"
	@echo "  Commit:     $(GIT_COMMIT)"
	@echo "  Build Date: $(BUILD_DATE)"
	@echo "  Go Version: $$(go version)"
	@echo "  Install:    $(INSTALL_PATH)"

.DEFAULT_GOAL := help
