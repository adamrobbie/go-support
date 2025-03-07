import React, { useState } from 'react';
import ScreenshotPlayer from './ScreenshotPlayer';
import RemoteControl from './RemoteControl';
import VideoPlayer from './VideoPlayer';

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

interface ClientDetailsProps {
  client: Client | null;
  onClose: () => void;
  onRequestScreenshot: (clientId: string) => void;
  onSendMouseEvent: (clientId: string, action: string, x: number, y: number, button?: string, double?: boolean, amount?: number) => void;
  onSendKeyboardEvent: (clientId: string, action: string, key: string, keys?: string[], text?: string) => void;
  onRequestScreenSize: (clientId: string) => void;
  onRequestMousePosition: (clientId: string) => void;
  onStartVideoStream: (clientId: string) => void;
  onStopVideoStream: (clientId: string) => void;
  onStartRecording: (clientId: string) => void;
  onStopRecording: (clientId: string) => void;
}

const ClientDetails: React.FC<ClientDetailsProps> = ({ 
  client, 
  onClose, 
  onRequestScreenshot,
  onSendMouseEvent,
  onSendKeyboardEvent,
  onRequestScreenSize,
  onRequestMousePosition,
  onStartVideoStream,
  onStopVideoStream,
  onStartRecording,
  onStopRecording
}) => {
  const [viewMode, setViewMode] = useState<'grid' | 'video'>('grid');
  const [activeTab, setActiveTab] = useState<'info' | 'control' | 'video'>('info');
  
  if (!client) {
    return null;
  }

  const screenshots = client.screenshots || [];
  const videoFrames = client.videoFrames || [];
  const hasScreenshots = screenshots.length > 0;
  const hasVideoFrames = videoFrames.length > 0;
  const isStreaming = client.isStreaming || false;
  const isRecording = client.isRecording || false;

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
          <button 
            className={`btn btn-sm ${activeTab === 'video' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab('video')}
          >
            Video Stream
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
                    <tr>
                      <td>Streaming:</td>
                      <td>{isStreaming ? 'Yes' : 'No'}</td>
                    </tr>
                    <tr>
                      <td>Recording:</td>
                      <td>{isRecording ? 'Yes' : 'No'}</td>
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
          ) : activeTab === 'control' ? (
            <RemoteControl 
              clientId={client.id}
              screenWidth={client.screenWidth}
              screenHeight={client.screenHeight}
              onSendMouseEvent={onSendMouseEvent}
              onSendKeyboardEvent={onSendKeyboardEvent}
              onRequestScreenSize={onRequestScreenSize}
              onRequestMousePosition={onRequestMousePosition}
            />
          ) : (
            <div className="video-stream-container">
              <h3>Live Video Stream</h3>
              
              {hasVideoFrames ? (
                <VideoPlayer 
                  frames={videoFrames}
                  width={800}
                  height={450}
                  autoPlay={true}
                  fps={10}
                  showControls={true}
                />
              ) : (
                <p>No video frames available. Start streaming to see the live video.</p>
              )}
              
              <div className="video-controls">
                {!isStreaming ? (
                  <button 
                    className="btn btn-success" 
                    onClick={() => onStartVideoStream(client.id)}
                  >
                    Start Streaming
                  </button>
                ) : (
                  <button 
                    className="btn btn-danger" 
                    onClick={() => onStopVideoStream(client.id)}
                  >
                    Stop Streaming
                  </button>
                )}
                
                {!isRecording ? (
                  <button 
                    className="btn btn-success" 
                    onClick={() => onStartRecording(client.id)}
                    disabled={!isStreaming}
                  >
                    Start Recording
                  </button>
                ) : (
                  <button 
                    className="btn btn-danger" 
                    onClick={() => onStopRecording(client.id)}
                  >
                    Stop Recording
                  </button>
                )}
              </div>
              
              <div className="video-info">
                <p>
                  <strong>Status:</strong> {isStreaming ? 'Streaming' : 'Not Streaming'} 
                  {isRecording ? ' (Recording)' : ''}
                </p>
                <p><strong>Frames Received:</strong> {videoFrames.length}</p>
                {videoFrames.length > 0 && (
                  <p>
                    <strong>Last Frame:</strong> {new Date(videoFrames[videoFrames.length - 1].timestamp).toLocaleString()}
                  </p>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ClientDetails; 