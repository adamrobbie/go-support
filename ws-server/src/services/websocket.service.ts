import WebSocket from 'ws';
import http from 'http';

export class WebSocketService {
  private wss: WebSocket.Server;
  private clients: Map<string, WebSocket> = new Map();
  private clientIdCounter: number = 0;

  constructor(server: http.Server) {
    this.wss = new WebSocket.Server({ server });
    this.setupWebSocketServer();
  }

  private setupWebSocketServer(): void {
    this.wss.on('connection', (ws: WebSocket) => {
      const clientId = this.generateClientId();
      this.clients.set(clientId, ws);
      
      console.log(`Client connected: ${clientId}`);
      
      // Send welcome message
      this.sendToClient(clientId, {
        type: 'welcome',
        message: `Welcome! Your client ID is ${clientId}`,
        clientId,
      });
      
      // Broadcast new connection to all clients
      this.broadcast({
        type: 'notification',
        message: `Client ${clientId} has joined`,
        clientCount: this.clients.size,
      }, clientId);
      
      // Handle messages from client
      ws.on('message', (message: WebSocket.Data) => {
        try {
          const parsedMessage = JSON.parse(message.toString());
          console.log(`Received message from ${clientId}:`, parsedMessage);
          
          // Handle different message types
          this.handleMessage(clientId, parsedMessage);
        } catch (error) {
          console.error('Error parsing message:', error);
          this.sendToClient(clientId, {
            type: 'error',
            message: 'Invalid message format. Please send JSON.',
          });
        }
      });
      
      // Handle client disconnection
      ws.on('close', () => {
        console.log(`Client disconnected: ${clientId}`);
        this.clients.delete(clientId);
        
        // Broadcast disconnection to all clients
        this.broadcast({
          type: 'notification',
          message: `Client ${clientId} has left`,
          clientCount: this.clients.size,
        });
      });
      
      // Handle errors
      ws.on('error', (error) => {
        console.error(`Error with client ${clientId}:`, error);
        this.clients.delete(clientId);
      });
    });
  }
  
  private generateClientId(): string {
    return `client-${++this.clientIdCounter}`;
  }
  
  private handleMessage(clientId: string, message: any): void {
    // Handle different message types
    switch (message.type) {
      case 'chat':
        // Broadcast chat message to all clients
        this.broadcast({
          type: 'chat',
          message: message.message,
          sender: clientId,
          timestamp: new Date().toISOString(),
        });
        break;
        
      case 'ping':
        // Respond with pong
        this.sendToClient(clientId, {
          type: 'pong',
          timestamp: new Date().toISOString(),
        });
        break;
        
      default:
        // Echo back the message
        this.sendToClient(clientId, {
          type: 'echo',
          originalMessage: message,
          timestamp: new Date().toISOString(),
        });
        break;
    }
  }
  
  public sendToClient(clientId: string, message: any): void {
    const client = this.clients.get(clientId);
    if (client && client.readyState === WebSocket.OPEN) {
      client.send(JSON.stringify(message));
    }
  }
  
  public broadcast(message: any, excludeClientId?: string): void {
    this.clients.forEach((client, clientId) => {
      if (excludeClientId !== clientId && client.readyState === WebSocket.OPEN) {
        client.send(JSON.stringify(message));
      }
    });
  }
  
  public getClientCount(): number {
    return this.clients.size;
  }
}

export default WebSocketService; 