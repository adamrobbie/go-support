import React, { useState, useEffect, useCallback } from 'react';
import useWebSocket from '../hooks/useWebSocket';
import ClientList from './ClientList';
import ClientDetails from './ClientDetails';
import ServerStats from './ServerStats';
import ConnectionStatus from './ConnectionStatus';

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
  screenshots?: Screenshot[];
  platform?: string;
  version?: string;
  ipAddress?: string;
}

interface ServerInfo {
  uptime: string;
  clientCount: number;
  messageCount: number;
}

const App: React.FC = () => {
  const [clients, setClients] = useState<Client[]>([]);
  const [serverInfo, setServerInfo] = useState<ServerInfo>({
    uptime: '0s',
    clientCount: 0,
    messageCount: 0,
  });
  const [clientId, setClientId] = useState<string | null>(null);
  const [selectedClient, setSelectedClient] = useState<Client | null>(null);

  // Get the WebSocket URL from the current window location
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}`;

  const { isConnected, messages, sendMessage } = useWebSocket(wsUrl, {
    reconnectInterval: 3000,
    reconnectAttempts: 10,
  });

  // Register as dashboard when connected
  const registerAsDashboard = useCallback(() => {
    if (isConnected && clientId) {
      sendMessage({
        type: 'registerAsDashboard',
      });
      console.log('Registered as dashboard client');
    }
  }, [isConnected, clientId, sendMessage]);

  // Handle client selection
  const handleSelectClient = useCallback((client: Client) => {
    setSelectedClient(client);
  }, []);

  // Close client details modal
  const handleCloseClientDetails = useCallback(() => {
    setSelectedClient(null);
  }, []);

  // Request screenshot from client
  const handleRequestScreenshot = useCallback((targetClientId: string) => {
    if (isConnected) {
      sendMessage({
        type: 'requestScreenshot',
        targetClientId
      });
      console.log(`Requested screenshot from client ${targetClientId}`);
    }
  }, [isConnected, sendMessage]);

  // Process WebSocket messages
  useEffect(() => {
    if (messages.length === 0) return;

    const latestMessage = messages[messages.length - 1];

    switch (latestMessage.type) {
      case 'welcome':
        // When we connect, we get our client ID
        if (latestMessage.clientId) {
          setClientId(latestMessage.clientId);
        }
        break;
      case 'notification':
        // Update client count from notifications
        if (latestMessage.clientCount !== undefined) {
          setServerInfo((prev) => ({
            ...prev,
            clientCount: latestMessage.clientCount,
          }));
        }
        break;
      case 'clientList':
        // Update client list
        if (Array.isArray(latestMessage.clients)) {
          // Add mock data for demonstration purposes
          const enhancedClients = latestMessage.clients.map((client: Client) => ({
            ...client,
            platform: client.platform || ['Windows', 'macOS', 'Linux'][Math.floor(Math.random() * 3)],
            version: client.version || `1.${Math.floor(Math.random() * 10)}.${Math.floor(Math.random() * 10)}`,
            ipAddress: client.ipAddress || `192.168.1.${Math.floor(Math.random() * 255)}`,
            screenshots: client.screenshots || []
          }));
          
          setClients(enhancedClients);
          
          // Update selected client if it exists in the new client list
          if (selectedClient) {
            const updatedSelectedClient = enhancedClients.find(c => c.id === selectedClient.id);
            if (updatedSelectedClient) {
              setSelectedClient(updatedSelectedClient);
            }
          }
        }
        break;
      case 'serverInfo':
        // Update server info
        if (latestMessage.uptime) {
          setServerInfo({
            uptime: latestMessage.uptime,
            clientCount: latestMessage.clientCount || 0,
            messageCount: latestMessage.messageCount || 0,
          });
        }
        break;
      case 'screenshot':
        // Handle screenshot message
        if (latestMessage.clientId && latestMessage.imageUrl) {
          setClients(prevClients => {
            return prevClients.map(client => {
              if (client.id === latestMessage.clientId) {
                const newScreenshot: Screenshot = {
                  id: `screenshot-${Date.now()}`,
                  timestamp: latestMessage.timestamp || new Date().toISOString(),
                  imageUrl: latestMessage.imageUrl,
                  width: latestMessage.width || 800,
                  height: latestMessage.height || 600
                };
                
                return {
                  ...client,
                  screenshots: [...(client.screenshots || []), newScreenshot]
                };
              }
              return client;
            });
          });
          
          // Update selected client if it's the one that sent the screenshot
          if (selectedClient && selectedClient.id === latestMessage.clientId) {
            setSelectedClient(prevSelected => {
              if (!prevSelected) return null;
              
              const newScreenshot: Screenshot = {
                id: `screenshot-${Date.now()}`,
                timestamp: latestMessage.timestamp || new Date().toISOString(),
                imageUrl: latestMessage.imageUrl,
                width: latestMessage.width || 800,
                height: latestMessage.height || 600
              };
              
              return {
                ...prevSelected,
                screenshots: [...(prevSelected.screenshots || []), newScreenshot]
              };
            });
          }
        }
        break;
      default:
        break;
    }
  }, [messages, selectedClient]);

  // Register as dashboard when connected and we have a client ID
  useEffect(() => {
    if (isConnected && clientId) {
      registerAsDashboard();
    }
  }, [isConnected, clientId, registerAsDashboard]);

  return (
    <div className="container">
      <header className="header">
        <h1>Go Support Dashboard</h1>
      </header>
      
      <ConnectionStatus isConnected={isConnected} serverUrl={wsUrl} />
      
      <div className="grid">
        <ServerStats 
          uptime={serverInfo.uptime}
          clientCount={serverInfo.clientCount}
          messageCount={serverInfo.messageCount}
        />
        
        <ClientList 
          clients={clients} 
          onSelectClient={handleSelectClient} 
        />
      </div>
      
      {selectedClient && (
        <ClientDetails 
          client={selectedClient} 
          onClose={handleCloseClientDetails} 
          onRequestScreenshot={handleRequestScreenshot}
        />
      )}
    </div>
  );
};

export default App; 