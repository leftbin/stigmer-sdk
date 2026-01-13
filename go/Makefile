# Stigmer Agent SDK Go - Makefile

.PHONY: help
help: ## Display this help message
	@echo "Stigmer Agent SDK Go - Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the SDK
	@echo "Building SDK..."
	@go build ./...

.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	@go test -v -race -cover ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-short
test-short: ## Run short tests only
	@echo "Running short tests..."
	@go test -v -short ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running linters..."
	@golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code with gofmt
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

.PHONY: tidy
tidy: ## Run go mod tidy
	@echo "Tidying dependencies..."
	@go mod tidy

.PHONY: verify
verify: fmt vet lint test ## Run all verification steps (fmt, vet, lint, test)
	@echo "All verification steps passed!"

.PHONY: clean
clean: ## Clean build artifacts and test caches
	@echo "Cleaning..."
	@go clean -testcache
	@rm -f coverage.out coverage.html

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed successfully!"

.PHONY: proto-check
proto-check: ## Verify proto stubs are available
	@echo "Checking proto stubs..."
	@test -d ../../apis/stubs/go || (echo "ERROR: Proto stubs not found at ../../apis/stubs/go" && exit 1)
	@echo "Proto stubs found!"

.PHONY: examples
examples: ## Run all examples
	@echo "Running examples..."
	@for example in examples/*.go; do \
		echo "Running $$example..."; \
		go run $$example || exit 1; \
	done
	@echo "All examples completed!"

.DEFAULT_GOAL := help
