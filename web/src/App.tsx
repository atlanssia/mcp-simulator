import { useState } from 'react';
import { ServerList } from './components/ServerList';
import { CreateServerForm } from './components/CreateServerForm';
import { LLMSettings } from './components/LLMSettings';
import { ToolManager } from './components/ToolManager';
import { type ServerConfig } from './lib/api';
import { Server, Settings } from 'lucide-react';

function App() {
  const [refreshKey, setRefreshKey] = useState(0);
  const [showSettings, setShowSettings] = useState(false);
  const [selectedServer, setSelectedServer] = useState<ServerConfig | null>(null);

  const handleServerCreated = () => {
    setRefreshKey(prev => prev + 1);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
      <div className="container mx-auto px-4 py-8">
        <header className="mb-8">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3 mb-2">
              <div className="p-2 bg-blue-600 rounded-lg">
                <Server className="w-8 h-8 text-white" />
              </div>
              <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                MCP Simulator
              </h1>
            </div>
            <button
              onClick={() => setShowSettings(true)}
              className="flex items-center gap-2 px-4 py-2 bg-gray-800 hover:bg-gray-700 rounded-lg text-white transition-colors"
              title="LLM 配置"
            >
              <Settings className="w-5 h-5" />
              <span className="hidden sm:inline">设置</span>
            </button>
          </div>
          <p className="text-gray-400 ml-14">
            Manage virtual MCP servers and generate AI-powered mock data
          </p>
        </header>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <ServerList
              key={refreshKey}
              onSelectServer={setSelectedServer}
            />
          </div>
          <div>
            <CreateServerForm onServerCreated={handleServerCreated} />
          </div>
        </div>
      </div>

      {showSettings && <LLMSettings onClose={() => setShowSettings(false)} />}

      {selectedServer && (
        <ToolManager
          serverId={selectedServer.id}
          serverName={selectedServer.name}
          onClose={() => setSelectedServer(null)}
        />
      )}
    </div>
  );
}

export default App;
