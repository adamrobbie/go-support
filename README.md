# Go WebSocket CLI

A simple cross-platform command-line application that connects to a WebSocket server using configuration from a `.env` file. The application also includes screen sharing permission management, platform-specific application identification, and high-quality screenshot capabilities.

## Features

- Connect to a WebSocket server specified in the `.env` file
- Receive and display messages from the WebSocket server
- Send periodic ping messages to keep the connection alive
- Graceful shutdown on interrupt signals (Ctrl+C)
- Verbose mode for additional logging
- Cross-platform screen sharing permission management
- Platform-specific application identification
- Support for connecting to a TypeScript WebSocket server
- High-quality screenshot capture across platforms (Windows, macOS, Linux)
- Screenshot region selection and quality settings
- Image format conversion and compression

## Project Structure

The project is organized into the following directories:

- `app/`: Main Go application code
- `client/`: WebSocket client implementation
- `pkg/`: Shared packages
  - `appid/`: Application identification
  - `permissions/`: Permission management
  - `screenshot/`: Cross-platform screenshot functionality
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

3. Create a `.env` file in the app directory with your WebSocket URLs and screenshot configuration:
   ```
   cd app
   cp .env.example .env
   ```
   Then edit the `.env` file to set your WebSocket URLs:
   ```
   WEBSOCKET_URL=wss://your-websocket-server.com/ws
   TS_WEBSOCKET_URL=ws://localhost:8080
   ```

## Screenshot Functionality

The application includes a cross-platform screenshot module that works on Windows, macOS, and Linux. The module provides the following features:

- Capture full-screen screenshots with different quality settings
- Capture screenshots of specific regions
- Save screenshots to a configurable directory
- Convert between image formats (PNG, JPEG)
- Compress images to reduce file size
- Resize images with high-quality interpolation
- Send screenshots through WebSocket to a server

### Screenshot Configuration

You can configure the screenshot functionality using the following environment variables or command-line flags:

- `SCREENSHOT_DIR` or `--screenshot-dir`: Directory to save screenshots (default: `~/Screenshots`)

### Screenshot Quality Settings

The screenshot module supports three quality settings:

- `Low`: Faster capture with smaller file size
- `Medium`: Balanced quality and performance
- `High`: Highest quality with larger file size

### WebSocket Screenshot Commands

The application supports sending screenshots through WebSocket connections. You can use the following commands in the console:

- `screenshot`: Take and send a medium-quality screenshot
- `screenshot low`: Take and send a low-quality screenshot
- `screenshot high`: Take and send a high-quality screenshot
- `region x y w h`: Take and send a screenshot of a specific region
- `message <text>`: Send a chat message
- `help`: Show available commands
- `exit`: Exit the application

### Server-Initiated Screenshots

The server can request screenshots by sending a message with the type `request_screenshot`. The message can include the following fields:

- `quality`: Quality setting (`low`, `medium`, or `high`)
- `message`: Description of the screenshot request
- `region`: Region to capture (object with `x`, `y`, `width`, and `height` properties)

Example server request for a full screenshot:
```json
{
  "type": "request_screenshot",
  "message": "Please send a screenshot of your desktop",
  "metadata": {
    "quality": "high"
  }
}
```

Example server request for a region screenshot:
```json
{
  "type": "request_screenshot",
  "message": "Please send a screenshot of the specified region",
  "metadata": {
    "quality": "medium",
    "region": {
      "x": 100,
      "y": 100,
      "width": 800,
      "height": 600
    }
  }
}
```

### Screenshot Message Format

When a screenshot is sent through the WebSocket connection, it uses the following format:

```json
{
  "type": "screenshot",
  "message": "Screenshot description",
  "timestamp": "2023-03-07T12:34:56Z",
  "screenshotData": "base64-encoded-image-data",
  "imageFormat": "png",
  "width": 1920,
  "height": 1080,
  "metadata": {
    "platform": "darwin",
    "arch": "amd64"
  }
}
```

### Example Usage

```go
// Capture a full-screen screenshot with high quality
ss, err := screenshot.Capture(screenshot.High)
if err != nil {
    log.Fatalf("Failed to capture screenshot: %v", err)
}

// Save the screenshot to a file
if err := ss.SaveToFile("screenshot.png"); err != nil {
    log.Fatalf("Failed to save screenshot: %v", err)
}

// Capture a specific region of the screen
region := screenshot.Region{
    X:      100,
    Y:      100,
    Width:  800,
    Height: 600,
}
ss, err = screenshot.CaptureRegion(region, screenshot.High)
if err != nil {
    log.Fatalf("Failed to capture region: %v", err)
}

// Resize the screenshot
if err := ss.Resize(400, 300); err != nil {
    log.Fatalf("Failed to resize screenshot: %v", err)
}

// Compress the screenshot
if err := ss.Compress(80); err != nil {
    log.Fatalf("Failed to compress screenshot: %v", err)
}

// Convert to JPEG format
if err := ss.ConvertToFormat("jpeg"); err != nil {
    log.Fatalf("Failed to convert format: %v", err)
}

// Send the screenshot through WebSocket
base64Data := ss.ToBase64()
err = wsClient.SendScreenshot(base64Data, ss.Format, ss.Width, ss.Height, "Screenshot description")
if err != nil {
    log.Fatalf("Failed to send screenshot: %v", err)
}
```

## Usage

### Command-Line Flags

The application supports the following command-line flags:

- `--verbose`: Enable verbose logging
- `--skip-permissions`: Skip permission checks
- `--use-ts-ws`: Use TypeScript WebSocket server
- `--ws`: WebSocket server URL (overrides environment variable)
- `--ts-ws`: TypeScript WebSocket server URL (overrides environment variable)
- `--screenshot-dir`: Directory to save screenshots (overrides environment variable)

### Makefile Commands

The project includes a Makefile with the following commands:

- `make deps`: Install dependencies
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
- `pkg/screenshot`: Tests for the screenshot functionality
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