# Go WebSocket CLI

A simple cross-platform command-line application that connects to a WebSocket server using configuration from a `.env` file. The application also includes screen sharing permission management and platform-specific application identification.

## Features

- Connect to a WebSocket server specified in the `.env` file
- Receive and display messages from the WebSocket server
- Send periodic ping messages to keep the connection alive
- Graceful shutdown on interrupt signals (Ctrl+C)
- Verbose mode for additional logging
- Cross-platform screen sharing permission management
- Platform-specific application identification
- Support for connecting to a TypeScript WebSocket server

## Project Structure

The project is organized into the following directories:

- `app/`: Main Go application code
- `client/`: WebSocket client implementation
- `pkg/`: Shared packages
  - `appid/`: Application identification
  - `permissions/`: Permission management
- `ws-server/`: TypeScript WebSocket server

## Installation

### Prerequisites

- Go 1.21 or higher
- Node.js 14.x or higher (for TypeScript WebSocket server)
- npm 6.x or higher (for TypeScript WebSocket server)

### Setup

1. Clone the repository:
   ```
   git clone https://github.com/adamrobbie/go-support.git
   cd go-support
   ```

2. Install dependencies:
   ```
   make deps
   ```

3. Create a `.env` file in the app directory with your WebSocket URLs:
   ```
   cd app
   cp .env.example .env
   ```
   Then edit the `.env` file to set your WebSocket URLs:
   ```
   WEBSOCKET_URL=wss://your-websocket-server.com/ws
   TS_WEBSOCKET_URL=ws://localhost:8080
   ```

## Usage

### Go Application

Using the Makefile:

```bash
# Build the application
make build

# Run the application
make run

# Run with verbose logging
make run-verbose

# Run with permissions skipped
make run-skip-permissions

# Run with TypeScript WebSocket server
make run-ts-ws

# Run with TypeScript WebSocket server and verbose logging
make run-ts-ws-verbose
```

Or manually:

```bash
# Build
cd app && go build -o ../go-support .

# Run
./go-support

# With verbose logging
./go-support -verbose

# Skip permission checks
./go-support -skip-permissions

# Use TypeScript WebSocket server
./go-support -use-ts-ws

# Use TypeScript WebSocket server with verbose logging
./go-support -use-ts-ws -verbose
```

### TypeScript WebSocket Server

Using the Makefile:

```bash
# Install dependencies
make ts-install

# Build the server
make ts-build

# Start the server in development mode (with hot reloading)
make ts-dev

# Start the server in production mode
make ts-start

# Stop the server
make ts-stop

# Run both the Go application and TypeScript server together
make run-all
```

Or manually:

```bash
# Install dependencies
cd ws-server && npm install

# Start in development mode
cd ws-server && npm run dev

# Build
cd ws-server && npm run build

# Start in production mode
cd ws-server && npm start
```

## Command-line Options

- `-verbose`: Enable verbose logging
- `-skip-permissions`: Skip screen sharing permission checks
- `-use-ts-ws`: Use the TypeScript WebSocket server instead of the default WebSocket server

## Screen Sharing Permissions

The application includes a permission manager that handles screen sharing permissions across different platforms:

- **macOS**: Uses the `screencapture` command to check for and request screen recording permissions. Opens System Preferences to the Screen Recording privacy settings and provides an interactive retry mechanism.
- **Windows**: Opens Windows Settings to the Screen Recording privacy section
- **Linux**: Attempts to open privacy settings using xdg-open (varies by desktop environment)

### Permission Flow

When the application starts, it will check for screen sharing permissions. If permissions are not granted:

1. The application will display clear instructions on how to grant permissions
2. It will open the appropriate system settings panel
3. For macOS, you can:
   - Press 'r' to retry the permission check after granting permission
   - Press 'q' to quit the application and restart it later

If you don't want to deal with permissions, you can use the `-skip-permissions` flag to bypass the permission checks.

## Application Identification

The application sets up platform-specific identification to make it recognizable by the operating system:

- **macOS**: Creates an application bundle with a proper Info.plist file containing the bundle identifier
- **Windows**: Sets up an Application User Model ID (AUMID)
- **Linux**: Creates a desktop entry file in the user's local applications directory

This allows the application to be properly recognized by the operating system, which is particularly useful when running through Cursor or other IDEs.

## TypeScript WebSocket Server

The application includes a TypeScript WebSocket server in the `ws-server` directory. This server provides:

- WebSocket communication with JSON messages
- RESTful API endpoints
- Interactive test client

To use the TypeScript WebSocket server:

1. Start the server:
   ```
   make ts-dev
   ```

2. Run the Go client with the `-use-ts-ws` flag:
   ```
   make run-ts-ws
   ```

Or run both together:
```
make run-all
```

For more information about the TypeScript WebSocket server, see the [ws-server README](ws-server/README.md).

## Development

### Requirements

- Go 1.21+
- GoReleaser (for releases)
- Node.js 14+ (for TypeScript WebSocket server)
- npm 6+ (for TypeScript WebSocket server)

### Makefile Commands

The project includes a Makefile with common commands:

- `make build`: Build the application
- `make clean`: Clean build artifacts
- `make test`: Run tests
- `make test-verbose`: Run tests with verbose output
- `make test-coverage`: Run tests with coverage and open the coverage report
- `make run`: Build and run the application
- `make run-verbose`: Run with verbose logging
- `make run-skip-permissions`: Run with permissions skipped
- `make run-ts-ws`: Run with TypeScript WebSocket server
- `make run-ts-ws-verbose`: Run with TypeScript WebSocket server and verbose logging
- `make ts-install`: Install TypeScript WebSocket server dependencies
- `make ts-build`: Build the TypeScript WebSocket server
- `make ts-dev`: Start the TypeScript WebSocket server in development mode
- `make ts-start`: Start the TypeScript WebSocket server in production mode
- `make ts-stop`: Stop the TypeScript WebSocket server
- `make run-all`: Run both the Go application and TypeScript server together
- `make deps`: Update dependencies for both Go and TypeScript
- `make release-dry-run`: Test the release process with a snapshot build
- `make release`: Create a release (requires a tag)
- `make tag`: Create a new Git tag for release

### Testing

The application includes a comprehensive test suite:

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage and open the coverage report
make test-coverage
```

The test suite includes:

- **Unit tests**: Test individual functions and methods
- **Integration tests**: Test the interaction between components
- **Mock implementations**: Test components in isolation

The tests are organized by package:

- `pkg/permissions`: Tests for the permission manager
- `pkg/appid`: Tests for the application identifier
- `app`: Tests for the main application logic

### Release Process

This project uses [GoReleaser](https://goreleaser.com/) to automate the release process:

1. Create a new tag:
   ```
   make tag
   ```

2. Push the tag to GitHub:
   ```
   git push origin <tag>
   ```

3. GitHub Actions will automatically build and publish the release

Alternatively, you can manually create a release:

1. Create and push a tag:
   ```
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. Run GoReleaser:
   ```
   make release
   ```

## License

[MIT](LICENSE) 