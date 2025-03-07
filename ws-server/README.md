# WebSocket Server

A TypeScript WebSocket server designed to work with the Go WebSocket client application.

## Features

- WebSocket server using the `ws` library
- RESTful API endpoints using Express
- Cross-platform compatibility
- Interactive test client
- TypeScript for type safety

## Prerequisites

- Node.js 14.x or higher
- npm 6.x or higher

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/adamrobbie/go-support.git
   cd go-support/ws-server
   ```

2. Install dependencies:
   ```
   npm install
   ```

3. Create a `.env` file based on `.env.example`:
   ```
   cp .env.example .env
   ```

## Usage

### Development

Run the server in development mode with hot reloading:

```
npm run dev
```

### Production

Build and run the server in production mode:

```
npm run build
npm start
```

## API Endpoints

- `GET /api/health` - Health check endpoint
- `GET /api/clients` - Get the number of connected clients
- `POST /api/broadcast` - Broadcast a message to all connected clients

## WebSocket Protocol

The WebSocket server expects messages in JSON format with the following structure:

```json
{
  "type": "message_type",
  "message": "message_content",
  "additionalField1": "value1",
  "additionalField2": "value2"
}
```

### Message Types

- `chat` - Chat message
- `ping` - Ping message (server will respond with a pong)
- `custom` - Custom message type

### Example Messages

#### Chat Message

```json
{
  "type": "chat",
  "message": "Hello, world!"
}
```

#### Ping Message

```json
{
  "type": "ping"
}
```

## Testing

An interactive test client is available at the root URL (`http://localhost:8080`).

## License

[ISC](LICENSE) 