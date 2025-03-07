import React from 'react';

interface ConnectionStatusProps {
  isConnected: boolean;
  serverUrl: string;
}

const ConnectionStatus: React.FC<ConnectionStatusProps> = ({ isConnected, serverUrl }) => {
  return (
    <div className="card">
      <div className="card-header">
        <h2 className="card-title">Connection Status</h2>
        <span className={`badge ${isConnected ? 'badge-success' : 'badge-danger'}`}>
          {isConnected ? 'Connected' : 'Disconnected'}
        </span>
      </div>
      <div>
        <p><strong>Server URL:</strong> {serverUrl}</p>
      </div>
    </div>
  );
};

export default ConnectionStatus; 