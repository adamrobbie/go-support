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

## Installation

### Prerequisites

- Go 1.21 or higher

### Setup

1. Clone the repository:
   ```
   git clone https://github.com/adamrobbie/go-support.git
   cd go-support
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Create a `.env` file in the project root with your WebSocket URL:
   ```
   WEBSOCKET_URL=wss://your-websocket-server.com/ws
   ```

## Usage

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
```

Or manually:

```bash
# Build
go build -o go-support .

# Run
./go-support

# With verbose logging
./go-support -verbose

# Skip permission checks
./go-support -skip-permissions
```

## Command-line Options

- `-verbose`: Enable verbose logging
- `-skip-permissions`: Skip screen sharing permission checks

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

## Development

### Requirements

- Go 1.21+
- GoReleaser (for releases)

### Makefile Commands

The project includes a Makefile with common commands:

- `make build`: Build the application
- `make clean`: Clean build artifacts
- `make test`: Run tests
- `make run`: Build and run the application
- `make run-verbose`: Run with verbose logging
- `make run-skip-permissions`: Run with permissions skipped
- `make deps`: Update dependencies
- `make release-dry-run`: Test the release process with a snapshot build
- `make release`: Create a release (requires a tag)
- `make tag`: Create a new Git tag for release

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