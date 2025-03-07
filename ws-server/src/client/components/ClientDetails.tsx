import React, { useState } from 'react';
import ScreenshotPlayer from './ScreenshotPlayer';

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

interface ClientDetailsProps {
  client: Client | null;
  onClose: () => void;
  onRequestScreenshot: (clientId: string) => void;
}

const ClientDetails: React.FC<ClientDetailsProps> = ({ client, onClose, onRequestScreenshot }) => {
  const [viewMode, setViewMode] = useState<'grid' | 'video'>('grid');
  
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
        <div className="modal-body">
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
        </div>
      </div>
    </div>
  );
};

export default ClientDetails; 