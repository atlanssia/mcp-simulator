import { useState, useEffect } from 'react';
import { type ServerConfig, api } from '../lib/api';
import { Play, Square, Wrench } from 'lucide-react';
import { cn } from '../lib/utils';
import { ToolManager } from './ToolManager';

export function ServerList() {
    const [servers, setServers] = useState<ServerConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedServer, setSelectedServer] = useState<ServerConfig | null>(null);

    const fetchServers = async () => {
        try {
            const response = await api.listServers();
            setServers(response.data);
        } catch (error) {
            console.error('Failed to fetch servers:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchServers();
        // Poll for status updates every 2 seconds
        const interval = setInterval(fetchServers, 2000);
        return () => clearInterval(interval);
    }, []);

    const handleStart = async (id: string) => {
        try {
            await api.startServer(id);
            fetchServers();
        } catch (error) {
            console.error('Failed to start server:', error);
        }
    };

    const handleStop = async (id: string) => {
        try {
            await api.stopServer(id);
            fetchServers();
        } catch (error) {
            console.error('Failed to stop server:', error);
        }
    };

    const getStatusBadge = (status?: string) => {
        const statusColors = {
            running: 'bg-green-500/20 text-green-400 border-green-500/30',
            stopped: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
            starting: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
            error: 'bg-red-500/20 text-red-400 border-red-500/30',
        };

        const color = statusColors[status as keyof typeof statusColors] || statusColors.stopped;
        const displayStatus = status || 'stopped';

        return (
            <span className={cn(
                "px-2 py-1 rounded-full text-xs font-medium border",
                color
            )}>
                {displayStatus}
            </span>
        );
    };

    if (loading) {
        return (
            <div className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-lg p-6">
                <h2 className="text-2xl font-bold text-white mb-4">Virtual Servers</h2>
                <div className="text-gray-400 text-center py-8">Loading...</div>
            </div>
        );
    }

    // Sort servers: running first, then by status priority, then alphabetically by name
    const sortedServers = [...servers].sort((a, b) => {
        // Define status priority: running > starting > stopped > error
        const statusPriority: { [key: string]: number } = {
            running: 0,
            starting: 1,
            stopped: 2,
            error: 3,
        };

        const aPriority = statusPriority[a.status || 'stopped'] ?? 2;
        const bPriority = statusPriority[b.status || 'stopped'] ?? 2;

        // First sort by status priority
        if (aPriority !== bPriority) {
            return aPriority - bPriority;
        }

        // Then sort alphabetically by name
        return a.name.localeCompare(b.name);
    });

    return (
        <div className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-lg p-6">
            <h2 className="text-2xl font-bold text-white mb-4">Virtual Servers</h2>
            {servers.length === 0 ? (
                <div className="text-gray-400 text-center py-8">
                    No servers yet. Create one to get started!
                </div>
            ) : (
                <div className="grid gap-4">
                    {sortedServers.map((server) => (
                        <div
                            key={server.id}
                            className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-lg p-4 hover:border-gray-600 transition-colors"
                        >
                            <div className="flex items-center justify-between">
                                <div className="flex-1">
                                    <div className="flex items-center gap-3 mb-2">
                                        <h3 className="text-lg font-semibold text-white">{server.name}</h3>
                                        {getStatusBadge(server.status)}
                                    </div>
                                    <p className="text-sm text-gray-400">
                                        ID: {server.id} | Port: {server.port}
                                    </p>
                                </div>
                                <div className="flex gap-2">
                                    <button
                                        onClick={() => handleStart(server.id)}
                                        disabled={server.status === 'running' || server.status === 'starting'}
                                        className={cn(
                                            "p-2 rounded-lg transition-colors",
                                            server.status === 'running' || server.status === 'starting'
                                                ? "bg-gray-700 text-gray-500 cursor-not-allowed"
                                                : "bg-green-600 hover:bg-green-700 text-white"
                                        )}
                                    >
                                        <Play className="w-5 h-5" />
                                    </button>
                                    <button
                                        onClick={() => handleStop(server.id)}
                                        disabled={server.status === 'stopped' || server.status === 'starting'}
                                        className={cn(
                                            "p-2 rounded-lg transition-colors",
                                            server.status === 'stopped' || server.status === 'starting'
                                                ? "bg-gray-700 text-gray-500 cursor-not-allowed"
                                                : "bg-red-600 hover:bg-red-700 text-white"
                                        )}
                                    >
                                        <Square className="w-5 h-5" />
                                    </button>
                                    <button
                                        onClick={() => setSelectedServer(server)}
                                        className="p-2 rounded-lg transition-colors bg-purple-600 hover:bg-purple-700 text-white"
                                        title="Manage Tools"
                                    >
                                        <Wrench className="w-5 h-5" />
                                    </button>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            )}

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
