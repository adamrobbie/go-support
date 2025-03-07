import WebSocket from 'ws';
import http from 'http';

interface Screenshot {
  id: string;
  timestamp: string;
  imageUrl: string;
  width: number;
  height: number;
}

interface Client {
  id: string;
  connectedAt: string;
  type: 'dashboard' | 'regular';
  platform?: string;
  version?: string;
  ipAddress?: string;
  screenshots?: Screenshot[];
}

export class WebSocketService {
  private wss: WebSocket.Server;
  private clients: Map<string, { 
    ws: WebSocket, 
    connectedAt: string, 
    type: 'dashboard' | 'regular',
    platform?: string,
    version?: string,
    ipAddress?: string,
    screenshots: Screenshot[]
  }> = new Map();
  private clientIdCounter: number = 0;
  private startTime: Date = new Date();
  private messageCount: number = 0;

  constructor(server: http.Server) {
    this.wss = new WebSocket.Server({ server });
    this.setupWebSocketServer();
    
    // Send server info to all clients every 10 seconds
    setInterval(() => {
      this.broadcastServerInfo();
    }, 10000);
  }

  private setupWebSocketServer(): void {
    this.wss.on('connection', (ws: WebSocket, req: http.IncomingMessage) => {
      const clientId = this.generateClientId();
      const connectedAt = new Date().toISOString();
      const ipAddress = req.socket.remoteAddress || 'Unknown';
      
      // Default to regular client type
      this.clients.set(clientId, { 
        ws, 
        connectedAt, 
        type: 'regular',
        ipAddress,
        screenshots: []
      });
      
      console.log(`Client connected: ${clientId} from ${ipAddress}`);
      
      // Send welcome message
      this.sendToClient(clientId, {
        type: 'welcome',
        message: `Welcome! Your client ID is ${clientId}`,
        clientId,
      });
      
      // Send current client list to the new client
      this.sendClientList(clientId);
      
      // Send server info to the new client
      this.sendServerInfo(clientId);
      
      // Broadcast new connection to all clients
      this.broadcast({
        type: 'notification',
        message: `Client ${clientId} has joined`,
        clientCount: this.getRegularClientCount(),
      }, clientId);
      
      // Handle messages from client
      ws.on('message', (message: WebSocket.Data) => {
        try {
          const parsedMessage = JSON.parse(message.toString());
          console.log(`Received message from ${clientId}:`, parsedMessage);
          
          // Increment message count
          this.messageCount++;
          
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
          clientCount: this.getRegularClientCount(),
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
        
      case 'getClients':
        // Send client list
        this.sendClientList(clientId);
        break;
        
      case 'getServerInfo':
        // Send server info
        this.sendServerInfo(clientId);
        break;
        
      case 'registerAsDashboard':
        // Register this client as a dashboard
        const client = this.clients.get(clientId);
        if (client) {
          client.type = 'dashboard';
          this.clients.set(clientId, client);
          console.log(`Client ${clientId} registered as dashboard`);
          
          // Send updated client list and server info
          this.sendClientList(clientId);
          this.sendServerInfo(clientId);
        }
        break;
        
      case 'clientInfo':
        // Update client information
        const clientToUpdate = this.clients.get(clientId);
        if (clientToUpdate && message.platform) {
          clientToUpdate.platform = message.platform;
          clientToUpdate.version = message.version || clientToUpdate.version;
          this.clients.set(clientId, clientToUpdate);
          
          // Broadcast updated client list
          this.broadcastClientList();
        }
        break;
        
      case 'screenshot':
        // Handle screenshot from client
        const clientWithScreenshot = this.clients.get(clientId);
        if (clientWithScreenshot && message.imageUrl) {
          const screenshot: Screenshot = {
            id: `screenshot-${Date.now()}`,
            timestamp: new Date().toISOString(),
            imageUrl: message.imageUrl,
            width: message.width || 800,
            height: message.height || 600
          };
          
          clientWithScreenshot.screenshots.push(screenshot);
          this.clients.set(clientId, clientWithScreenshot);
          
          // Broadcast screenshot to all dashboard clients
          this.broadcastToDashboards({
            type: 'screenshot',
            clientId,
            timestamp: screenshot.timestamp,
            imageUrl: screenshot.imageUrl,
            width: screenshot.width,
            height: screenshot.height
          });
        }
        break;
        
      case 'requestScreenshot':
        // Request screenshot from a specific client
        if (message.targetClientId) {
          const targetClient = this.clients.get(message.targetClientId);
          if (targetClient && targetClient.type === 'regular') {
            this.sendToClient(message.targetClientId, {
              type: 'takeScreenshot',
              requestedBy: clientId,
              timestamp: new Date().toISOString()
            });
          }
        }
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
  
  private sendClientList(clientId: string): void {
    // Get only regular clients (not dashboards)
    const clientList: Client[] = Array.from(this.clients.entries())
      .filter(([id, { type }]) => type === 'regular')
      .map(([id, { connectedAt, type, platform, version, ipAddress, screenshots }]) => ({
        id,
        connectedAt,
        type,
        platform,
        version,
        ipAddress,
        screenshots
      }));
    
    this.sendToClient(clientId, {
      type: 'clientList',
      clients: clientList,
    });
  }
  
  private broadcastClientList(): void {
    // Get only regular clients (not dashboards)
    const clientList: Client[] = Array.from(this.clients.entries())
      .filter(([id, { type }]) => type === 'regular')
      .map(([id, { connectedAt, type, platform, version, ipAddress, screenshots }]) => ({
        id,
        connectedAt,
        type,
        platform,
        version,
        ipAddress,
        screenshots
      }));
    
    // Send to all dashboard clients
    Array.from(this.clients.entries())
      .filter(([id, { type }]) => type === 'dashboard')
      .forEach(([id]) => {
        this.sendToClient(id, {
          type: 'clientList',
          clients: clientList,
        });
      });
  }
  
  private sendServerInfo(clientId: string): void {
    this.sendToClient(clientId, {
      type: 'serverInfo',
      uptime: this.getUptime(),
      clientCount: this.getRegularClientCount(),
      messageCount: this.messageCount,
    });
  }
  
  private broadcastServerInfo(): void {
    this.broadcast({
      type: 'serverInfo',
      uptime: this.getUptime(),
      clientCount: this.getRegularClientCount(),
      messageCount: this.messageCount,
    });
  }
  
  private broadcastToDashboards(message: any): void {
    Array.from(this.clients.entries())
      .filter(([id, { type }]) => type === 'dashboard')
      .forEach(([id]) => {
        this.sendToClient(id, message);
      });
  }
  
  private getUptime(): string {
    const now = new Date();
    const uptimeMs = now.getTime() - this.startTime.getTime();
    
    const seconds = Math.floor(uptimeMs / 1000) % 60;
    const minutes = Math.floor(uptimeMs / (1000 * 60)) % 60;
    const hours = Math.floor(uptimeMs / (1000 * 60 * 60)) % 24;
    const days = Math.floor(uptimeMs / (1000 * 60 * 60 * 24));
    
    let uptime = '';
    if (days > 0) uptime += `${days}d `;
    if (hours > 0 || days > 0) uptime += `${hours}h `;
    if (minutes > 0 || hours > 0 || days > 0) uptime += `${minutes}m `;
    uptime += `${seconds}s`;
    
    return uptime;
  }
  
  public sendToClient(clientId: string, message: any): void {
    const client = this.clients.get(clientId);
    if (client && client.ws.readyState === WebSocket.OPEN) {
      client.ws.send(JSON.stringify(message));
    }
  }
  
  public broadcast(message: any, excludeClientId?: string): void {
    this.clients.forEach((client, clientId) => {
      if (excludeClientId !== clientId && client.ws.readyState === WebSocket.OPEN) {
        client.ws.send(JSON.stringify(message));
      }
    });
  }
  
  public getClientCount(): number {
    return this.clients.size;
  }
  
  public getRegularClientCount(): number {
    return Array.from(this.clients.values()).filter(client => client.type === 'regular').length;
  }
  
  public getClients(): Client[] {
    return Array.from(this.clients.entries()).map(([id, { connectedAt, type, platform, version, ipAddress, screenshots }]) => ({
      id,
      connectedAt,
      type,
      platform,
      version,
      ipAddress,
      screenshots
    }));
  }
  
  public getRegularClients(): Client[] {
    return Array.from(this.clients.entries())
      .filter(([_, { type }]) => type === 'regular')
      .map(([id, { connectedAt, type, platform, version, ipAddress, screenshots }]) => ({
        id,
        connectedAt,
        type,
        platform,
        version,
        ipAddress,
        screenshots
      }));
  }
  
  public getMessageCount(): number {
    return this.messageCount;
  }
}

export default WebSocketService; 