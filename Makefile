# Stigmer SDK - Root Makefile

.PHONY: help
help: ## Display this help message
	@echo "Stigmer SDK - Available targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Language-specific commands:"
	@echo "  make go-<command>      - Run command in Go SDK"
	@echo "  make python-<command>  - Run command in Python SDK"

# --------------------------------------------------------------------------- #
#  Dependency Management
# --------------------------------------------------------------------------- #

.PHONY: update-deps
update-deps: ## Update all SDK dependencies (Go + Python)
	@echo "============================================"
	@echo "Updating All SDK Dependencies"
	@echo "============================================"
	@echo ""
	@echo "Step 1: Updating Go SDK dependencies..."
	@$(MAKE) -C go update-deps
	@echo ""
	@echo "Step 2: Updating Python SDK dependencies..."
	@echo "  (Python uses poetry - run 'cd python && poetry update' manually if needed)"
	@echo ""
	@echo "============================================"
	@echo "✓ SDK dependencies updated successfully!"
	@echo "============================================"

.PHONY: go-update-deps
go-update-deps: ## Update Go SDK dependencies only
	@$(MAKE) -C go update-deps

.PHONY: python-update-deps
python-update-deps: ## Update Python SDK dependencies (poetry)
	@echo "Updating Python SDK dependencies..."
	@cd python && poetry update
	@echo "✓ Python dependencies updated!"

# --------------------------------------------------------------------------- #
#  Release Management
# --------------------------------------------------------------------------- #

.PHONY: release
release: ## Create and push a release tag (usage: make release TAG=v0.1.3)
ifndef TAG
	@echo "ERROR: TAG is required. Usage: make release TAG=v0.1.3"
	@exit 1
endif
	@echo "Creating release tag: $(TAG)"
	@if git rev-parse "$(TAG)" >/dev/null 2>&1; then \
		echo "ERROR: Tag $(TAG) already exists"; \
		exit 1; \
	fi
	@git tag -a "$(TAG)" -m "Release $(TAG)"
	@git push origin "$(TAG)"
	@echo "✓ Release tag $(TAG) created and pushed successfully!"
	@echo ""
	@echo "View release: https://github.com/leftbin/stigmer-sdk/releases/tag/$(TAG)"

# --------------------------------------------------------------------------- #
#  Go SDK Commands
# --------------------------------------------------------------------------- #

.PHONY: go-build
go-build: ## Build Go SDK
	@$(MAKE) -C go build

.PHONY: go-test
go-test: ## Run Go SDK tests
	@$(MAKE) -C go test

.PHONY: go-test-coverage
go-test-coverage: ## Run Go SDK tests with coverage
	@$(MAKE) -C go test-coverage

.PHONY: go-lint
go-lint: ## Run Go SDK linters
	@$(MAKE) -C go lint

.PHONY: go-fmt
go-fmt: ## Format Go SDK code
	@$(MAKE) -C go fmt

.PHONY: go-tidy
go-tidy: ## Tidy Go SDK dependencies
	@$(MAKE) -C go tidy

.PHONY: go-examples
go-examples: ## Run Go SDK examples
	@$(MAKE) -C go examples

# --------------------------------------------------------------------------- #
#  Python SDK Commands
# --------------------------------------------------------------------------- #

.PHONY: python-install
python-install: ## Install Python SDK dependencies
	@echo "Installing Python SDK dependencies..."
	@cd python && poetry install

.PHONY: python-test
python-test: ## Run Python SDK tests
	@echo "Running Python SDK tests..."
	@cd python && poetry run pytest

.PHONY: python-lint
python-lint: ## Run Python SDK linters
	@echo "Running Python SDK linters..."
	@cd python && poetry run ruff check .

.PHONY: python-fmt
python-fmt: ## Format Python SDK code
	@echo "Formatting Python SDK code..."
	@cd python && poetry run ruff format .

# --------------------------------------------------------------------------- #
#  Combined Commands
# --------------------------------------------------------------------------- #

.PHONY: build
build: go-build ## Build all SDKs
	@echo "✓ All SDKs built!"

.PHONY: test
test: go-test python-test ## Run all tests
	@echo "✓ All tests completed!"

.PHONY: lint
lint: go-lint python-lint ## Run all linters
	@echo "✓ All linting completed!"

.PHONY: fmt
fmt: go-fmt python-fmt ## Format all code
	@echo "✓ All code formatted!"

.PHONY: clean
clean: ## Clean all build artifacts
	@echo "Cleaning Go SDK..."
	@$(MAKE) -C go clean
	@echo "Cleaning Python SDK..."
	@cd python && rm -rf dist/ build/ *.egg-info .pytest_cache/ .coverage htmlcov/
	@echo "✓ All artifacts cleaned!"

.DEFAULT_GOAL := help
