import { Request, Response, Router } from 'express';
import WebSocketService from '../services/websocket.service';

export class ApiController {
  private router: Router;
  private wsService: WebSocketService;

  constructor(wsService: WebSocketService) {
    this.router = Router();
    this.wsService = wsService;
    this.setupRoutes();
  }

  private setupRoutes(): void {
    // Health check endpoint
    this.router.get('/health', this.healthCheck.bind(this));
    
    // Get client count
    this.router.get('/clients/count', this.getClientCount.bind(this));
    
    // Get client list
    this.router.get('/clients', this.getClients.bind(this));
    
    // Get server info
    this.router.get('/server-info', this.getServerInfo.bind(this));
    
    // Broadcast message to all clients
    this.router.post('/broadcast', this.broadcastMessage.bind(this));
    
    // Remote control endpoints
    this.router.post('/clients/:clientId/mouse', this.sendMouseEvent.bind(this));
    this.router.post('/clients/:clientId/keyboard', this.sendKeyboardEvent.bind(this));
    this.router.get('/clients/:clientId/screen-size', this.requestScreenSize.bind(this));
    this.router.get('/clients/:clientId/mouse-position', this.requestMousePosition.bind(this));
    this.router.post('/clients/:clientId/screenshot', this.requestScreenshot.bind(this));
  }

  private healthCheck(req: Request, res: Response): void {
    res.status(200).json({
      status: 'ok',
      timestamp: new Date().toISOString(),
    });
  }

  private getClientCount(req: Request, res: Response): void {
    res.status(200).json({
      clientCount: this.wsService.getRegularClientCount(),
      timestamp: new Date().toISOString(),
    });
  }
  
  private getClients(req: Request, res: Response): void {
    res.status(200).json({
      clients: this.wsService.getRegularClients(),
      timestamp: new Date().toISOString(),
    });
  }
  
  private getServerInfo(req: Request, res: Response): void {
    const startTime = new Date(Date.now() - process.uptime() * 1000);
    const uptime = this.formatUptime(process.uptime());
    
    res.status(200).json({
      uptime,
      startTime: startTime.toISOString(),
      clientCount: this.wsService.getRegularClientCount(),
      messageCount: this.wsService.getMessageCount(),
      timestamp: new Date().toISOString(),
    });
  }

  private broadcastMessage(req: Request, res: Response): void {
    const { message, type = 'broadcast' } = req.body;
    
    if (!message) {
      res.status(400).json({
        error: 'Message is required',
      });
      return;
    }
    
    this.wsService.broadcast({
      type,
      message,
      source: 'api',
      timestamp: new Date().toISOString(),
    });
    
    res.status(200).json({
      success: true,
      message: 'Message broadcasted successfully',
    });
  }
  
  private sendMouseEvent(req: Request, res: Response): void {
    const { clientId } = req.params;
    const { action, x, y, button = 'left', double = false, amount = 0 } = req.body;
    
    if (!action) {
      res.status(400).json({
        error: 'Action is required',
      });
      return;
    }
    
    this.wsService.sendMouseEvent(clientId, action, x, y, button, double, amount);
    
    res.status(200).json({
      success: true,
      message: `Mouse event sent to client ${clientId}`,
    });
  }
  
  private sendKeyboardEvent(req: Request, res: Response): void {
    const { clientId } = req.params;
    const { action, key, keys, text } = req.body;
    
    if (!action) {
      res.status(400).json({
        error: 'Action is required',
      });
      return;
    }
    
    this.wsService.sendKeyboardEvent(clientId, action, key, keys, text);
    
    res.status(200).json({
      success: true,
      message: `Keyboard event sent to client ${clientId}`,
    });
  }
  
  private requestScreenSize(req: Request, res: Response): void {
    const { clientId } = req.params;
    
    this.wsService.requestScreenSize(clientId);
    
    res.status(200).json({
      success: true,
      message: `Screen size requested from client ${clientId}`,
    });
  }
  
  private requestMousePosition(req: Request, res: Response): void {
    const { clientId } = req.params;
    
    this.wsService.requestMousePosition(clientId);
    
    res.status(200).json({
      success: true,
      message: `Mouse position requested from client ${clientId}`,
    });
  }
  
  private requestScreenshot(req: Request, res: Response): void {
    const { clientId } = req.params;
    
    // Send a takeScreenshot message to the client
    this.wsService.sendToClient(clientId, {
      type: 'takeScreenshot',
      requestedBy: 'api',
      timestamp: new Date().toISOString()
    });
    
    res.status(200).json({
      success: true,
      message: `Screenshot requested from client ${clientId}`,
    });
  }
  
  private formatUptime(seconds: number): string {
    const days = Math.floor(seconds / (3600 * 24));
    const hours = Math.floor((seconds % (3600 * 24)) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);
    
    let uptime = '';
    if (days > 0) uptime += `${days}d `;
    if (hours > 0 || days > 0) uptime += `${hours}h `;
    if (minutes > 0 || hours > 0 || days > 0) uptime += `${minutes}m `;
    uptime += `${secs}s`;
    
    return uptime;
  }

  public getRouter(): Router {
    return this.router;
  }
}

export default ApiController; 