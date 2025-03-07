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

interface VideoFrame {
  frameData: string; // Base64 encoded image data
  timestamp: string;
}

interface Client {
  id: string;
  connectedAt: string;
  type: 'dashboard' | 'regular';
  screenshots?: Screenshot[];
  videoFrames?: VideoFrame[];
  platform?: string;
  version?: string;
  ipAddress?: string;
  screenWidth?: number;
  screenHeight?: number;
  mouseX?: number;
  mouseY?: number;
  isStreaming?: boolean;
  isRecording?: boolean;
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
  VIDEO_FRAME: 'videoFrame',
  START_VIDEO: 'startVideo',
  STOP_VIDEO: 'stopVideo',
  START_RECORDING: 'startRecording',
  STOP_RECORDING: 'stopRecording',
};

// Maximum number of video frames to keep per client
const MAX_VIDEO_FRAMES = 100;

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

  // Start video streaming from client
  const handleStartVideoStream = useCallback((targetClientId: string) => {
    if (isConnected) {
      sendMessage({
        type: MessageTypes.START_VIDEO,
        targetClientId
      });
      console.log(`Requested to start video streaming from client ${targetClientId}`);
      
      // Update client's streaming status
      setClients(prevClients => {
        return prevClients.map(client => {
          if (client.id === targetClientId) {
            return {
              ...client,
              isStreaming: true
            };
          }
          return client;
        });
      });
      
      // Update selected client if it's the one we're streaming from
      if (selectedClient && selectedClient.id === targetClientId) {
        setSelectedClient(prevSelected => {
          if (!prevSelected) return null;
          return {
            ...prevSelected,
            isStreaming: true
          };
        });
      }
    }
  }, [isConnected, sendMessage, selectedClient]);

  // Stop video streaming from client
  const handleStopVideoStream = useCallback((targetClientId: string) => {
    if (isConnected) {
      sendMessage({
        type: MessageTypes.STOP_VIDEO,
        targetClientId
      });
      console.log(`Requested to stop video streaming from client ${targetClientId}`);
      
      // Update client's streaming status
      setClients(prevClients => {
        return prevClients.map(client => {
          if (client.id === targetClientId) {
            return {
              ...client,
              isStreaming: false
            };
          }
          return client;
        });
      });
      
      // Update selected client if it's the one we're streaming from
      if (selectedClient && selectedClient.id === targetClientId) {
        setSelectedClient(prevSelected => {
          if (!prevSelected) return null;
          return {
            ...prevSelected,
            isStreaming: false
          };
        });
      }
    }
  }, [isConnected, sendMessage, selectedClient]);

  // Start recording from client
  const handleStartRecording = useCallback((targetClientId: string) => {
    if (isConnected) {
      sendMessage({
        type: MessageTypes.START_RECORDING,
        targetClientId
      });
      console.log(`Requested to start recording from client ${targetClientId}`);
      
      // Update client's recording status
      setClients(prevClients => {
        return prevClients.map(client => {
          if (client.id === targetClientId) {
            return {
              ...client,
              isRecording: true
            };
          }
          return client;
        });
      });
      
      // Update selected client if it's the one we're recording from
      if (selectedClient && selectedClient.id === targetClientId) {
        setSelectedClient(prevSelected => {
          if (!prevSelected) return null;
          return {
            ...prevSelected,
            isRecording: true
          };
        });
      }
    }
  }, [isConnected, sendMessage, selectedClient]);

  // Stop recording from client
  const handleStopRecording = useCallback((targetClientId: string) => {
    if (isConnected) {
      sendMessage({
        type: MessageTypes.STOP_RECORDING,
        targetClientId
      });
      console.log(`Requested to stop recording from client ${targetClientId}`);
      
      // Update client's recording status
      setClients(prevClients => {
        return prevClients.map(client => {
          if (client.id === targetClientId) {
            return {
              ...client,
              isRecording: false
            };
          }
          return client;
        });
      });
      
      // Update selected client if it's the one we're recording from
      if (selectedClient && selectedClient.id === targetClientId) {
        setSelectedClient(prevSelected => {
          if (!prevSelected) return null;
          return {
            ...prevSelected,
            isRecording: false
          };
        });
      }
    }
  }, [isConnected, sendMessage, selectedClient]);

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
      case MessageTypes.SCREEN_SIZE:
        // Handle screen size message
        if (latestMessage.clientId && latestMessage.width && latestMessage.height) {
          setClients(prevClients => {
            return prevClients.map(client => {
              if (client.id === latestMessage.clientId) {
                return {
                  ...client,
                  screenWidth: latestMessage.width,
                  screenHeight: latestMessage.height
                };
              }
              return client;
            });
          });
          
          // Update selected client if it's the one that sent the screen size
          if (selectedClient && selectedClient.id === latestMessage.clientId) {
            setSelectedClient(prevSelected => {
              if (!prevSelected) return null;
              return {
                ...prevSelected,
                screenWidth: latestMessage.width,
                screenHeight: latestMessage.height
              };
            });
          }
        }
        break;
      case MessageTypes.MOUSE_POSITION:
        // Handle mouse position message
        if (latestMessage.clientId && latestMessage.x !== undefined && latestMessage.y !== undefined) {
          setClients(prevClients => {
            return prevClients.map(client => {
              if (client.id === latestMessage.clientId) {
                return {
                  ...client,
                  mouseX: latestMessage.x,
                  mouseY: latestMessage.y
                };
              }
              return client;
            });
          });
          
          // Update selected client if it's the one that sent the mouse position
          if (selectedClient && selectedClient.id === latestMessage.clientId) {
            setSelectedClient(prevSelected => {
              if (!prevSelected) return null;
              return {
                ...prevSelected,
                mouseX: latestMessage.x,
                mouseY: latestMessage.y
              };
            });
          }
        }
        break;
      case MessageTypes.VIDEO_FRAME:
        // Handle video frame message
        if (latestMessage.clientId && latestMessage.frameData) {
          setClients(prevClients => {
            return prevClients.map(client => {
              if (client.id === latestMessage.clientId) {
                const newFrame: VideoFrame = {
                  frameData: latestMessage.frameData,
                  timestamp: latestMessage.timestamp || new Date().toISOString()
                };
                
                // Keep only the last MAX_VIDEO_FRAMES frames
                const existingFrames = client.videoFrames || [];
                const newFrames = [...existingFrames, newFrame];
                if (newFrames.length > MAX_VIDEO_FRAMES) {
                  newFrames.shift(); // Remove oldest frame
                }
                
                return {
                  ...client,
                  videoFrames: newFrames,
                  isStreaming: true
                };
              }
              return client;
            });
          });
          
          // Update selected client if it's the one that sent the video frame
          if (selectedClient && selectedClient.id === latestMessage.clientId) {
            setSelectedClient(prevSelected => {
              if (!prevSelected) return null;
              
              const newFrame: VideoFrame = {
                frameData: latestMessage.frameData,
                timestamp: latestMessage.timestamp || new Date().toISOString()
              };
              
              // Keep only the last MAX_VIDEO_FRAMES frames
              const existingFrames = prevSelected.videoFrames || [];
              const newFrames = [...existingFrames, newFrame];
              if (newFrames.length > MAX_VIDEO_FRAMES) {
                newFrames.shift(); // Remove oldest frame
              }
              
              return {
                ...prevSelected,
                videoFrames: newFrames,
                isStreaming: true
              };
            });
          }
        }
        break;
      default:
        break;
    }
  }, [messages, selectedClient]);

  // Register as dashboard when connected
  useEffect(() => {
    if (isConnected) {
      registerAsDashboard();
    }
  }, [isConnected, registerAsDashboard]);

  return (
    <div className="app">
      <header className="app-header">
        <h1>WebSocket Dashboard</h1>
        <ConnectionStatus isConnected={isConnected} serverUrl={wsUrl} />
      </header>
      <div className="app-content">
        <div className="sidebar">
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
        <div className="main-content">
          {selectedClient && (
            <ClientDetails 
              client={selectedClient} 
              onClose={handleCloseClientDetails} 
              onRequestScreenshot={handleRequestScreenshot}
              onSendMouseEvent={handleSendMouseEvent}
              onSendKeyboardEvent={handleSendKeyboardEvent}
              onRequestScreenSize={handleRequestScreenSize}
              onRequestMousePosition={handleRequestMousePosition}
              onStartVideoStream={handleStartVideoStream}
              onStopVideoStream={handleStopVideoStream}
              onStartRecording={handleStartRecording}
              onStopRecording={handleStopRecording}
            />
          )}
        </div>
      </div>
    </div>
  );
};

export default App; 