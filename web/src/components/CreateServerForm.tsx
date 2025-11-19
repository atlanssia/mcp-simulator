import { useState } from 'react';
import { Plus } from 'lucide-react';
import { api } from '../lib/api';
import { cn } from '../lib/utils';

interface CreateServerFormProps {
    onServerCreated: () => void;
}

export function CreateServerForm({ onServerCreated }: CreateServerFormProps) {
    const [id, setId] = useState('');
    const [name, setName] = useState('');
    const [port, setPort] = useState('');
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        try {
            await api.createServer({
                id,
                name,
                port: parseInt(port, 10),
            });
            setId('');
            setName('');
            setPort('');
            onServerCreated();
        } catch (error) {
            console.error('Failed to create server:', error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <form onSubmit={handleSubmit} className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-lg p-6">
            <h2 className="text-xl font-bold text-white mb-4">Create New Server</h2>
            <div className="space-y-4">
                <div>
                    <label className="block text-sm font-medium text-gray-300 mb-2">
                        Server ID
                    </label>
                    <input
                        type="text"
                        value={id}
                        onChange={(e) => setId(e.target.value)}
                        required
                        className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        placeholder="srv1"
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium text-gray-300 mb-2">
                        Server Name
                    </label>
                    <input
                        type="text"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        required
                        className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        placeholder="GitHub Mock"
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium text-gray-300 mb-2">
                        Port
                    </label>
                    <input
                        type="number"
                        value={port}
                        onChange={(e) => setPort(e.target.value)}
                        required
                        className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                        placeholder="8081"
                    />
                </div>
                <button
                    type="submit"
                    disabled={loading}
                    className={cn(
                        "w-full flex items-center justify-center gap-2 px-4 py-2 rounded-lg font-medium transition-colors",
                        "bg-blue-600 hover:bg-blue-700 text-white",
                        "disabled:opacity-50 disabled:cursor-not-allowed"
                    )}
                >
                    <Plus className="w-5 h-5" />
                    Create Server
                </button>
            </div>
        </form>
    );
}
