.PHONY: help build run test test-race test-ci clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the server binary
	@echo "Building lineradb-server..."
	@mkdir -p bin
	go build -o bin/lineradb-server ./cmd

run: build ## Build and run the server
	@echo "Starting lineradb-server..."
	./bin/lineradb-server

test: ## Run all tests (fast, no race detector)
	@echo "Running tests..."
	go test -v ./test/...

test-race: ## Run tests with race detector (requires CGO)
	@echo "Running tests with race detector..."
	CGO_ENABLED=1 go test -v -race ./test/...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out
	go clean

.DEFAULT_GOAL := help