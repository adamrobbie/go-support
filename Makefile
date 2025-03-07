.PHONY: build clean test test-verbose test-coverage run run-verbose run-skip-permissions run-ts-ws release-dry-run release

# Application name
APP_NAME=go-support

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w"

# Main build target
build:
	$(GOBUILD) $(LDFLAGS) -o $(APP_NAME) .

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(APP_NAME)
	rm -rf dist/
	rm -f coverage.out

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with verbose output
test-verbose:
	$(GOTEST) -v -count=1 ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Run the application
run:
	$(GOBUILD) -o $(APP_NAME) .
	./$(APP_NAME)

# Run with verbose logging
run-verbose:
	$(GOBUILD) -o $(APP_NAME) .
	./$(APP_NAME) -verbose

# Skip permissions
run-skip-permissions:
	$(GOBUILD) -o $(APP_NAME) .
	./$(APP_NAME) -skip-permissions

# Run with TypeScript WebSocket server
run-ts-ws:
	$(GOBUILD) -o $(APP_NAME) .
	./$(APP_NAME) -use-ts-ws

# Run with TypeScript WebSocket server and verbose logging
run-ts-ws-verbose:
	$(GOBUILD) -o $(APP_NAME) .
	./$(APP_NAME) -use-ts-ws -verbose

# Update dependencies
deps:
	$(GOMOD) tidy

# GoReleaser dry run
release-dry-run:
	goreleaser release --snapshot --clean

# GoReleaser release
release:
	goreleaser release --clean

# Create a new tag for release
tag:
	@echo "Current tags:"
	@git tag
	@echo ""
	@read -p "Enter new version (e.g. v1.0.0): " version; \
	git tag -a $$version -m "Release $$version"; \
	echo "Tag $$version created. Push with: git push origin $$version" 