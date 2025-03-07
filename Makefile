.PHONY: build clean test test-verbose test-coverage run run-verbose run-skip-permissions run-ts-ws run-ts-ws-verbose ts-install ts-build ts-start ts-dev ts-stop run-all run-ts-ws-verbose release-dry-run release

# Application name
APP_NAME=go-support

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Node parameters
NPM=npm
NPMI=$(NPM) install
NPMS=$(NPM) start
NPMB=$(NPM) run build
NPMD=$(NPM) run dev

# Build flags
LDFLAGS=-ldflags "-s -w"

# Main build target
build:
	cd app && $(GOBUILD) $(LDFLAGS) -o ../$(APP_NAME) .

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(APP_NAME)
	rm -rf dist/
	rm -f coverage.out
	rm -rf ws-server/dist/

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
run: build
	./$(APP_NAME)

# Run with verbose logging
run-verbose: build
	./$(APP_NAME) -verbose

# Skip permissions
run-skip-permissions: build
	./$(APP_NAME) -skip-permissions

# Run with TypeScript WebSocket server
run-ts-ws: build
	./$(APP_NAME) -use-ts-ws

# Run with TypeScript WebSocket server and verbose logging
run-ts-ws-verbose: build
	./$(APP_NAME) -use-ts-ws -verbose

# Install TypeScript WebSocket server dependencies
ts-install:
	cd ws-server && $(NPMI)

# Build TypeScript WebSocket server
ts-build: ts-install
	cd ws-server && $(NPMB)

# Start TypeScript WebSocket server in production mode
ts-start: ts-build
	cd ws-server && $(NPMS)

# Start TypeScript WebSocket server in development mode
ts-dev: ts-install
	cd ws-server && $(NPMD)

# Stop TypeScript WebSocket server
ts-stop:
	-pkill -f "node.*ws-server" || true

# Run both Go app and TypeScript server
run-all: ts-dev-bg run-ts-ws

# Run TypeScript server in background
ts-dev-bg: ts-install
	cd ws-server && $(NPMD) &

# Update dependencies
deps:
	$(GOMOD) tidy
	cd ws-server && $(NPMI)

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