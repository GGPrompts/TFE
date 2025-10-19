.PHONY: test test-coverage test-verbose clean build

# Build the application
build:
	go build -o tfe .

# Run all tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total

# Run tests with verbose output
test-verbose:
	go test -v -cover ./...

# Clean build artifacts and test files
clean:
	rm -f tfe coverage.out coverage.html

# Run specific test
test-run:
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-run TEST=TestName"; \
		exit 1; \
	fi
	go test -v -run $(TEST)

# Help
help:
	@echo "Available targets:"
	@echo "  make build          - Build the application"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make test-run       - Run specific test (use TEST=TestName)"
	@echo "  make clean          - Clean build artifacts"
