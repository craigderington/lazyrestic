.PHONY: help build test test-verbose test-short test-coverage test-integration clean install run fmt vet lint

# Binary name
BINARY_NAME=lazyrestic
INSTALL_PATH=/usr/local/bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Test parameters
TEST_FLAGS=-v -race
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Build flags
LDFLAGS=-ldflags "-s -w"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v

build-all: ## Build for all platforms
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe

test: ## Run all tests
	$(GOTEST) ./... -cover

test-verbose: ## Run tests with verbose output
	$(GOTEST) $(TEST_FLAGS) ./...

test-short: ## Run only unit tests (skip integration tests)
	$(GOTEST) -short ./...

test-coverage: ## Run tests with coverage report
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep total

test-integration: ## Run only integration tests
	$(GOTEST) -v ./pkg/restic -run Integration

test-bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

clean: ## Remove build artifacts and cache
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

install: build ## Install binary to system
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)"
	sudo mv $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installation complete!"

uninstall: ## Remove installed binary
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)

run: build ## Build and run the application
	./$(BINARY_NAME)

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

fmt: ## Format code
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOVET) ./...

lint: ## Run linter (requires golangci-lint)
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin" && exit 1)
	golangci-lint run

dev: ## Development mode: format, vet, test, build
	@$(MAKE) fmt
	@$(MAKE) vet
	@$(MAKE) test-short
	@$(MAKE) build

ci: ## CI pipeline: deps, fmt, vet, lint, test with coverage
	@$(MAKE) deps
	@$(MAKE) fmt
	@$(MAKE) vet
	@$(MAKE) test-coverage

setup-test-repo: ## Set up a test restic repository for testing
	@echo "Creating test repository at /tmp/restic-test"
	@rm -rf /tmp/restic-test
	@RESTIC_PASSWORD=testpassword restic -r /tmp/restic-test init
	@RESTIC_PASSWORD=testpassword restic -r /tmp/restic-test backup $(PWD)/CLAUDE.md --tag "config"
	@RESTIC_PASSWORD=testpassword restic -r /tmp/restic-test backup $(PWD)/README.md --tag "docs"
	@echo "Test repository created!"
	@echo ""
	@echo "Add this to your config:"
	@echo "  repositories:"
	@echo "    - name: test-repo"
	@echo "      path: /tmp/restic-test"
	@echo "      password: testpassword"

.DEFAULT_GOAL := help
