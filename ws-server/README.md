# WebSocket Server for Go Support

This is a WebSocket server for the Go Support application, providing real-time communication between clients and the dashboard.

## Features

- Real-time WebSocket communication
- Screenshot sharing and viewing
- Video streaming from clients
- Remote control functionality
- Client management dashboard

## Local Development

### Prerequisites

- Node.js (v18.x or later)
- npm (v9.x or later)

### Installation

1. Clone the repository
2. Navigate to the ws-server directory
3. Install dependencies

```bash
cd ws-server
npm install
```

### Configuration

Create a `.env` file in the root directory based on the `.env.example` file:

```bash
cp .env.example .env
```

Edit the `.env` file to match your environment.

### Running the Server

For development:

```bash
npm run dev
```

This will start the server with hot reloading.

For production:

```bash
npm run build
npm start
```

## Heroku Deployment

### Prerequisites

- [Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli)
- Git

### Deployment Steps

1. Login to Heroku

```bash
heroku login
```

2. Create a new Heroku app

```bash
heroku create your-app-name
```

3. Set environment variables

```bash
heroku config:set NODE_ENV=production
heroku config:set ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

4. Deploy to Heroku

```bash
git subtree push --prefix ws-server heroku main
```

Or if you're deploying from a subdirectory:

```bash
git push heroku `git subtree split --prefix ws-server main`:main --force
```

5. Open the app

```bash
heroku open
```

### Scaling on Heroku

To ensure your WebSocket server can handle multiple connections, you may need to scale your dynos:

```bash
heroku ps:scale web=1:standard-1x
```

### Monitoring

Monitor your application logs:

```bash
heroku logs --tail
```

## API Documentation

The server exposes the following API endpoints:

- `GET /api/health` - Health check endpoint
- `GET /api/clients` - Get list of connected clients
- `GET /api/clients/count` - Get count of connected clients
- `GET /api/server-info` - Get server information
- `POST /api/broadcast` - Broadcast a message to all clients
- `POST /api/clients/:clientId/screenshot` - Request a screenshot from a client
- `POST /api/clients/:clientId/video/start` - Start video streaming from a client
- `POST /api/clients/:clientId/video/stop` - Stop video streaming from a client
- `POST /api/clients/:clientId/video/record/start` - Start recording from a client
- `POST /api/clients/:clientId/video/record/stop` - Stop recording from a client

## WebSocket Protocol

The WebSocket server uses the following message types:

- `welcome` - Sent to clients when they connect
- `notification` - General notifications
- `clientList` - List of connected clients
- `serverInfo` - Server information
- `screenshot` - Screenshot data
- `takeScreenshot` - Request to take a screenshot
- `mouseEvent` - Mouse control events
- `keyboardEvent` - Keyboard control events
- `screenSize` - Screen size information
- `mousePosition` - Mouse position information
- `videoFrame` - Video frame data
- `startVideo` - Start video streaming
- `stopVideo` - Stop video streaming
- `startRecording` - Start recording
- `stopRecording` - Stop recording

## License

ISC 