import express from 'express';
import http from 'http';
import cors from 'cors';
import path from 'path';
import { config } from './config/env';
import WebSocketService from './services/websocket.service';
import ApiController from './controllers/api.controller';

// Create Express application
const app = express();

// Create HTTP server
const server = http.createServer(app);

// Initialize WebSocket service
const wsService = new WebSocketService(server);

// Middleware
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Configure CORS
const corsOptions = {
  origin: (origin: string | undefined, callback: (err: Error | null, allow?: boolean) => void) => {
    // Allow requests with no origin (like mobile apps, curl, Postman)
    if (!origin) {
      return callback(null, true);
    }
    
    // Check if the origin is allowed
    if (config.allowedOrigins.indexOf(origin) !== -1 || config.nodeEnv === 'development') {
      return callback(null, true);
    }
    
    callback(new Error('Not allowed by CORS'));
  },
  credentials: true,
};

app.use(cors(corsOptions));

// Serve static files from the public directory
app.use(express.static(path.join(__dirname, '../public')));

// API routes
const apiController = new ApiController(wsService);
app.use('/api', apiController.getRouter());

// Root route
app.get('/', (req, res) => {
  res.sendFile(path.join(__dirname, '../public/index.html'));
});

// Start the server
const PORT = parseInt(config.port, 10);
server.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
  console.log(`WebSocket server available at ws://localhost:${PORT}`);
  console.log(`HTTP API available at http://localhost:${PORT}`);
  console.log(`Test client available at http://localhost:${PORT}`);
  console.log(`Environment: ${config.nodeEnv}`);
}); 