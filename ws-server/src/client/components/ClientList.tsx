import React from 'react';

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

interface ClientListProps {
  clients: Client[];
  onSelectClient: (client: Client) => void;
}

const ClientList: React.FC<ClientListProps> = ({ clients, onSelectClient }) => {
  // Filter out dashboard clients (should already be filtered on the server, but just in case)
  const regularClients = clients.filter(client => client.type === 'regular');
  
  return (
    <div className="card">
      <div className="card-header">
        <h2 className="card-title">Connected Clients</h2>
        <span className="badge badge-primary">{regularClients.length}</span>
      </div>
      {regularClients.length === 0 ? (
        <p>No clients connected</p>
      ) : (
        <div className="table-container">
          <table className="client-table">
            <thead>
              <tr>
                <th>Client ID</th>
                <th>Connected At</th>
                <th>Platform</th>
                <th>Screenshots</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {regularClients.map((client) => (
                <tr 
                  key={client.id} 
                  className="client-row"
                  onClick={() => onSelectClient(client)}
                >
                  <td>{client.id}</td>
                  <td>{new Date(client.connectedAt).toLocaleString()}</td>
                  <td>{client.platform || 'Unknown'}</td>
                  <td>{client.screenshots?.length || 0}</td>
                  <td>
                    <button 
                      className="btn btn-sm btn-primary"
                      onClick={(e) => {
                        e.stopPropagation();
                        onSelectClient(client);
                      }}
                    >
                      View Details
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default ClientList; 