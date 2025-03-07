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
    this.router.get('/clients', this.getClientCount.bind(this));
    
    // Broadcast message to all clients
    this.router.post('/broadcast', this.broadcastMessage.bind(this));
  }

  private healthCheck(req: Request, res: Response): void {
    res.status(200).json({
      status: 'ok',
      timestamp: new Date().toISOString(),
    });
  }

  private getClientCount(req: Request, res: Response): void {
    res.status(200).json({
      clientCount: this.wsService.getClientCount(),
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

  public getRouter(): Router {
    return this.router;
  }
}

export default ApiController; 