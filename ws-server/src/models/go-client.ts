export interface GoClient {
  id: string;
  connectionTime: Date;
  lastActivity: Date;
  metadata?: {
    version?: string;
    platform?: string;
    [key: string]: any;
  };
}

export class GoClientManager {
  private clients: Map<string, GoClient> = new Map();

  public addClient(clientId: string, metadata?: any): GoClient {
    const client: GoClient = {
      id: clientId,
      connectionTime: new Date(),
      lastActivity: new Date(),
      metadata,
    };

    this.clients.set(clientId, client);
    return client;
  }

  public getClient(clientId: string): GoClient | undefined {
    return this.clients.get(clientId);
  }

  public removeClient(clientId: string): boolean {
    return this.clients.delete(clientId);
  }

  public updateClientActivity(clientId: string): void {
    const client = this.clients.get(clientId);
    if (client) {
      client.lastActivity = new Date();
    }
  }

  public updateClientMetadata(clientId: string, metadata: any): void {
    const client = this.clients.get(clientId);
    if (client) {
      client.metadata = { ...client.metadata, ...metadata };
    }
  }

  public getAllClients(): GoClient[] {
    return Array.from(this.clients.values());
  }

  public getClientCount(): number {
    return this.clients.size;
  }
}

export default GoClientManager; 