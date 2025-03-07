import React from 'react';

interface ServerStatsProps {
  uptime: string;
  clientCount: number;
  messageCount: number;
}

const ServerStats: React.FC<ServerStatsProps> = ({ uptime, clientCount, messageCount }) => {
  return (
    <div className="card">
      <div className="card-header">
        <h2 className="card-title">Server Statistics</h2>
      </div>
      <div>
        <p><strong>Uptime:</strong> {uptime}</p>
        <p><strong>Connected Clients:</strong> {clientCount}</p>
        <p><strong>Messages Processed:</strong> {messageCount}</p>
      </div>
    </div>
  );
};

export default ServerStats; 