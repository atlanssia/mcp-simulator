import { useState } from 'react';
import { ServerList } from './components/ServerList';
import { CreateServerForm } from './components/CreateServerForm';
import { AIGenerator } from './components/AIGenerator';
import { Server } from 'lucide-react';

function App() {
  const [refreshKey, setRefreshKey] = useState(0);

  const handleServerCreated = () => {
    setRefreshKey(prev => prev + 1);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <div className="container mx-auto px-4 py-8">
        <header className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="p-2 bg-blue-600 rounded-lg">
              <Server className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
              MCP Simulator
            </h1>
          </div>
          <p className="text-gray-400 ml-14">
            Manage virtual MCP servers and generate AI-powered mock data
          </p>
        </header>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <ServerList key={refreshKey} />
            <AIGenerator />
          </div>
          <div>
            <CreateServerForm onServerCreated={handleServerCreated} />
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
