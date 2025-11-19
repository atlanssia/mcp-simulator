import { useState, useEffect } from 'react';
import { Play, Square } from 'lucide-react';
import { api, type ServerConfig } from '../lib/api';
import { cn } from '../lib/utils';

export function ServerList() {
    const [servers, setServers] = useState<ServerConfig[]>([]);
    const [loading, setLoading] = useState(false);

    const loadServers = async () => {
        try {
            const { data } = await api.listServers();
            setServers(data || []);
        } catch (error) {
            console.error('Failed to load servers:', error);
        }
    };

    useEffect(() => {
        loadServers();
    }, []);

    const handleStart = async (id: string) => {
        setLoading(true);
        try {
            await api.startServer(id);
            await loadServers();
        } catch (error) {
            console.error('Failed to start server:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleStop = async (id: string) => {
        setLoading(true);
        try {
            await api.stopServer(id);
            await loadServers();
        } catch (error) {
            console.error('Failed to stop server:', error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="space-y-4">
            <h2 className="text-2xl font-bold text-white">Virtual Servers</h2>
            {servers.length === 0 ? (
                <div className="text-gray-400 text-center py-8">
                    No servers configured. Create one to get started.
                </div>
            ) : (
                <div className="grid gap-4">
                    {servers.map((server) => (
                        <div
                            key={server.id}
                            className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-lg p-4 hover:border-gray-600 transition-colors"
                        >
                            <div className="flex items-center justify-between">
                                <div>
                                    <h3 className="text-lg font-semibold text-white">{server.name}</h3>
                                    <p className="text-sm text-gray-400">
                                        ID: {server.id} • Port: {server.port}
                                    </p>
                                </div>
                                <div className="flex gap-2">
                                    <button
                                        onClick={() => handleStart(server.id)}
                                        disabled={loading}
                                        className={cn(
                                            "p-2 rounded-lg transition-colors",
                                            "bg-green-600 hover:bg-green-700 text-white",
                                            "disabled:opacity-50 disabled:cursor-not-allowed"
                                        )}
                                    >
                                        <Play className="w-5 h-5" />
                                    </button>
                                    <button
                                        onClick={() => handleStop(server.id)}
                                        disabled={loading}
                                        className={cn(
                                            "p-2 rounded-lg transition-colors",
                                            "bg-red-600 hover:bg-red-700 text-white",
                                            "disabled:opacity-50 disabled:cursor-not-allowed"
                                        )}
                                    >
                                        <Square className="w-5 h-5" />
                                    </button>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
