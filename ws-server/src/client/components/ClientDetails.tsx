import React, { useState } from 'react';
import ScreenshotPlayer from './ScreenshotPlayer';
import RemoteControl from './RemoteControl';

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

interface ClientDetailsProps {
  client: Client | null;
  onClose: () => void;
  onRequestScreenshot: (clientId: string) => void;
  onSendMouseEvent: (clientId: string, action: string, x: number, y: number, button?: string, double?: boolean, amount?: number) => void;
  onSendKeyboardEvent: (clientId: string, action: string, key: string, keys?: string[], text?: string) => void;
  onRequestScreenSize: (clientId: string) => void;
  onRequestMousePosition: (clientId: string) => void;
}

const ClientDetails: React.FC<ClientDetailsProps> = ({ 
  client, 
  onClose, 
  onRequestScreenshot,
  onSendMouseEvent,
  onSendKeyboardEvent,
  onRequestScreenSize,
  onRequestMousePosition
}) => {
  const [viewMode, setViewMode] = useState<'grid' | 'video'>('grid');
  const [activeTab, setActiveTab] = useState<'info' | 'control'>('info');
  
  if (!client) {
    return null;
  }

  const screenshots = client.screenshots || [];
  const hasScreenshots = screenshots.length > 0;

  return (
    <div className="modal">
      <div className="modal-content">
        <div className="modal-header">
          <h2>Client Details: {client.id}</h2>
          <button className="btn btn-sm" onClick={onClose}>×</button>
        </div>
        
        <div className="modal-tabs">
          <button 
            className={`btn btn-sm ${activeTab === 'info' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('info')}
          >
            Information
          </button>
          <button 
            className={`btn btn-sm ${activeTab === 'control' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('control')}
          >
            Remote Control
          </button>
        </div>
        
        <div className="modal-body">
          {activeTab === 'info' ? (
            <>
              <div className="client-info">
                <h3>Client Information</h3>
                <table className="info-table">
                  <tbody>
                    <tr>
                      <td>ID:</td>
                      <td>{client.id}</td>
                    </tr>
                    <tr>
                      <td>Connected At:</td>
                      <td>{new Date(client.connectedAt).toLocaleString()}</td>
                    </tr>
                    <tr>
                      <td>Platform:</td>
                      <td>{client.platform || 'Unknown'}</td>
                    </tr>
                    <tr>
                      <td>Version:</td>
                      <td>{client.version || 'Unknown'}</td>
                    </tr>
                    <tr>
                      <td>IP Address:</td>
                      <td>{client.ipAddress || 'Unknown'}</td>
                    </tr>
                    <tr>
                      <td>Screen Size:</td>
                      <td>{client.screenWidth && client.screenHeight ? `${client.screenWidth}×${client.screenHeight}` : 'Unknown'}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              
              <div className="screenshots">
                <div className="screenshots-header">
                  <h3>Screenshots</h3>
                  {hasScreenshots && (
                    <div className="view-mode-toggle">
                      <button 
                        className={`btn btn-sm ${viewMode === 'grid' ? 'btn-primary' : 'btn-secondary'}`}
                        onClick={() => setViewMode('grid')}
                      >
                        Grid View
                      </button>
                      <button 
                        className={`btn btn-sm ${viewMode === 'video' ? 'btn-primary' : 'btn-secondary'}`}
                        onClick={() => setViewMode('video')}
                      >
                        Video View
                      </button>
                    </div>
                  )}
                </div>
                
                {!hasScreenshots ? (
                  <p>No screenshots available for this client.</p>
                ) : viewMode === 'video' ? (
                  <ScreenshotPlayer 
                    screenshots={screenshots}
                    fps={2}
                    autoPlay={true}
                    width={800}
                    height={450}
                  />
                ) : (
                  <div className="screenshot-grid">
                    {screenshots.map((screenshot) => (
                      <div key={screenshot.id} className="screenshot-item">
                        <img 
                          src={screenshot.imageUrl} 
                          alt={`Screenshot from ${new Date(screenshot.timestamp).toLocaleString()}`} 
                          onClick={() => window.open(screenshot.imageUrl, '_blank')}
                        />
                        <div className="screenshot-info">
                          <span>{new Date(screenshot.timestamp).toLocaleString()}</span>
                          <span>{screenshot.width}×{screenshot.height}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
              
              <div className="actions">
                <button 
                  className="btn btn-primary" 
                  onClick={() => onRequestScreenshot(client.id)}
                >
                  Request Screenshot
                </button>
              </div>
            </>
          ) : (
            <RemoteControl 
              clientId={client.id}
              screenWidth={client.screenWidth}
              screenHeight={client.screenHeight}
              onSendMouseEvent={onSendMouseEvent}
              onSendKeyboardEvent={onSendKeyboardEvent}
              onRequestScreenSize={onRequestScreenSize}
              onRequestMousePosition={onRequestMousePosition}
            />
          )}
        </div>
      </div>
    </div>
  );
};

export default ClientDetails; 