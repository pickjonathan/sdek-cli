.PHONY: build test run clean install lint help

# Version injection
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE)"

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

install: ## Install sdek to $GOPATH/bin
	@echo "Installing sdek..."
	go install $(LDFLAGS) .

clean: ## Remove build artifacts
	@echo "Cleaning..."
	rm -f sdek
	rm -f coverage.out coverage.html
	rm -rf dist/

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
build-all: build-linux build-darwin build-windows ## Build for all platforms

build-linux: ## Build for Linux
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/sdek-linux-amd64 main.go

build-darwin: ## Build for macOS
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/sdek-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/sdek-darwin-arm64 main.go

build-windows: ## Build for Windows
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/sdek-windows-amd64.exe main.go

.DEFAULT_GOAL := help
