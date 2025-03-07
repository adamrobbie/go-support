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
  screenWidth?: number;
  screenHeight?: number;
  mouseX?: number;
  mouseY?: number;
}

interface ServerInfo {
  uptime: string;
  clientCount: number;
  messageCount: number;
}

// Message types
const MessageTypes = {
  WELCOME: 'welcome',
  NOTIFICATION: 'notification',
  CLIENT_LIST: 'clientList',
  SERVER_INFO: 'serverInfo',
  SCREENSHOT: 'screenshot',
  TAKE_SCREENSHOT: 'takeScreenshot',
  MOUSE_EVENT: 'mouseEvent',
  KEYBOARD_EVENT: 'keyboardEvent',
  SCREEN_SIZE: 'screenSize',
  MOUSE_POSITION: 'mousePosition',
};

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

  // Send mouse event to client
  const handleSendMouseEvent = useCallback((
    targetClientId: string, 
    action: string, 
    x: number, 
    y: number, 
    button: string = 'left', 
    double: boolean = false, 
    amount: number = 0
  ) => {
    if (isConnected) {
      sendMessage({
        type: MessageTypes.MOUSE_EVENT,
        targetClientId,
        action,
        x,
        y,
        button,
        double,
        amount
      });
      console.log(`Sent mouse event to client ${targetClientId}: ${action} at (${x},${y})`);
    }
  }, [isConnected, sendMessage]);

  // Send keyboard event to client
  const handleSendKeyboardEvent = useCallback((
    targetClientId: string, 
    action: string, 
    key: string, 
    keys?: string[], 
    text?: string
  ) => {
    if (isConnected) {
      sendMessage({
        type: MessageTypes.KEYBOARD_EVENT,
        targetClientId,
        action,
        key,
        keys,
        text
      });
      console.log(`Sent keyboard event to client ${targetClientId}: ${action} ${key}`);
    }
  }, [isConnected, sendMessage]);

  // Request screen size from client
  const handleRequestScreenSize = useCallback((targetClientId: string) => {
    if (isConnected) {
      sendMessage({
        type: MessageTypes.SCREEN_SIZE,
        targetClientId
      });
      console.log(`Requested screen size from client ${targetClientId}`);
    }
  }, [isConnected, sendMessage]);

  // Request mouse position from client
  const handleRequestMousePosition = useCallback((targetClientId: string) => {
    if (isConnected) {
      sendMessage({
        type: MessageTypes.MOUSE_POSITION,
        targetClientId
      });
      console.log(`Requested mouse position from client ${targetClientId}`);
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
          setClients(latestMessage.clients);
          
          // Update selected client if it exists in the new client list
          if (selectedClient) {
            const updatedSelectedClient = latestMessage.clients.find((c: Client) => c.id === selectedClient.id);
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
          onSendMouseEvent={handleSendMouseEvent}
          onSendKeyboardEvent={handleSendKeyboardEvent}
          onRequestScreenSize={handleRequestScreenSize}
          onRequestMousePosition={handleRequestMousePosition}
        />
      )}
    </div>
  );
};

export default App; 