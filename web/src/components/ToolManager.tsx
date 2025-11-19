import { useState, useEffect } from 'react';
import { type Tool, api } from '../lib/api';
import { Plus, Trash2, Edit2, Save, X } from 'lucide-react';

interface ToolManagerProps {
    serverId: string;
    serverName: string;
    onClose: () => void;
}

export function ToolManager({ serverId, serverName, onClose }: ToolManagerProps) {
    const [tools, setTools] = useState<Tool[]>([]);
    const [loading, setLoading] = useState(true);
    const [editingTool, setEditingTool] = useState<Tool | null>(null);
    const [isCreating, setIsCreating] = useState(false);

    const fetchTools = async () => {
        try {
            const response = await api.listTools(serverId);
            setTools(response.data);
        } catch (error) {
            console.error('Failed to fetch tools:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchTools();
    }, [serverId]);

    const handleCreate = () => {
        setIsCreating(true);
        setEditingTool({
            name: '',
            description: '',
            inputSchema: { type: 'object', properties: {} },
        });
    };

    const handleSave = async () => {
        if (!editingTool) return;

        try {
            if (isCreating) {
                await api.createTool(serverId, editingTool);
            } else {
                await api.updateTool(serverId, editingTool.name, editingTool);
            }
            await fetchTools();
            setEditingTool(null);
            setIsCreating(false);
        } catch (error) {
            console.error('Failed to save tool:', error);
        }
    };

    const handleDelete = async (toolName: string) => {
        if (!confirm(`Delete tool "${toolName}"?`)) return;

        try {
            await api.deleteTool(serverId, toolName);
            await fetchTools();
        } catch (error) {
            console.error('Failed to delete tool:', error);
        }
    };

    const handleCancel = () => {
        setEditingTool(null);
        setIsCreating(false);
    };

    if (loading) {
        return (
            <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
                <div className="bg-gray-900 rounded-lg p-8 max-w-4xl w-full mx-4">
                    <div className="text-white text-center">Loading tools...</div>
                </div>
            </div>
        );
    }

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
            <div className="bg-gray-900 rounded-lg p-6 max-w-4xl w-full mx-4 max-h-[90vh] overflow-y-auto">
                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-2xl font-bold text-white">
                        Tools for {serverName}
                    </h2>
                    <button
                        onClick={onClose}
                        className="p-2 hover:bg-gray-800 rounded-lg transition-colors"
                    >
                        <X className="w-5 h-5 text-gray-400" />
                    </button>
                </div>

                {editingTool ? (
                    <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-4 mb-4">
                        <h3 className="text-lg font-semibold text-white mb-4">
                            {isCreating ? 'Create New Tool' : 'Edit Tool'}
                        </h3>
                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-300 mb-1">
                                    Tool Name
                                </label>
                                <input
                                    type="text"
                                    value={editingTool.name}
                                    onChange={(e) => setEditingTool({ ...editingTool, name: e.target.value })}
                                    disabled={!isCreating}
                                    className="w-full bg-gray-800 border border-gray-600 rounded-lg px-3 py-2 text-white disabled:opacity-50"
                                    placeholder="e.g., get_weather"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-300 mb-1">
                                    Description
                                </label>
                                <div className="flex gap-2">
                                    <textarea
                                        value={editingTool.description}
                                        onChange={(e) => setEditingTool({ ...editingTool, description: e.target.value })}
                                        className="flex-1 bg-gray-800 border border-gray-600 rounded-lg px-3 py-2 text-white"
                                        rows={2}
                                        placeholder="What does this tool do?"
                                    />
                                    <button
                                        onClick={async () => {
                                            if (!editingTool.description) {
                                                alert('Please enter a description first');
                                                return;
                                            }
                                            try {
                                                const response = await api.generateMock({
                                                    prompt: `Generate a JSON schema for a tool that: ${editingTool.description}`,
                                                    schema: { type: 'object' }
                                                });
                                                if (response.data && typeof response.data === 'object') {
                                                    setEditingTool({ ...editingTool, inputSchema: response.data as unknown as Record<string, unknown> });
                                                }
                                            } catch (error) {
                                                console.error('Failed to generate schema:', error);
                                                alert('Failed to generate schema. Please try again.');
                                            }
                                        }}
                                        className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg text-white text-sm whitespace-nowrap transition-colors"
                                        title="Use AI to generate input schema based on description"
                                    >
                                        ✨ Generate Schema
                                    </button>
                                </div>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-300 mb-1">
                                    Input Schema (JSON)
                                </label>
                                <textarea
                                    value={JSON.stringify(editingTool.inputSchema, null, 2)}
                                    onChange={(e) => {
                                        try {
                                            const schema = JSON.parse(e.target.value);
                                            setEditingTool({ ...editingTool, inputSchema: schema });
                                        } catch { }
                                    }}
                                    className="w-full bg-gray-800 border border-gray-600 rounded-lg px-3 py-2 text-white font-mono text-sm"
                                    rows={8}
                                />
                            </div>
                            <div className="flex gap-2">
                                <button
                                    onClick={handleSave}
                                    className="flex items-center gap-2 px-4 py-2 bg-green-600 hover:bg-green-700 rounded-lg text-white transition-colors"
                                >
                                    <Save className="w-4 h-4" />
                                    Save
                                </button>
                                <button
                                    onClick={handleCancel}
                                    className="flex items-center gap-2 px-4 py-2 bg-gray-600 hover:bg-gray-700 rounded-lg text-white transition-colors"
                                >
                                    <X className="w-4 h-4" />
                                    Cancel
                                </button>
                            </div>
                        </div>
                    </div>
                ) : (
                    <>
                        <button
                            onClick={handleCreate}
                            className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg text-white transition-colors mb-4"
                        >
                            <Plus className="w-4 h-4" />
                            Add Tool
                        </button>

                        {tools.length === 0 ? (
                            <div className="text-gray-400 text-center py-8">
                                No tools defined. Click "Add Tool" to create one.
                            </div>
                        ) : (
                            <div className="grid gap-3">
                                {tools.map((tool) => (
                                    <div
                                        key={tool.name}
                                        className="bg-gray-800/50 border border-gray-700 rounded-lg p-4 hover:border-gray-600 transition-colors"
                                    >
                                        <div className="flex items-start justify-between">
                                            <div className="flex-1">
                                                <h4 className="text-lg font-semibold text-white">{tool.name}</h4>
                                                <p className="text-sm text-gray-400 mt-1">{tool.description}</p>
                                                <details className="mt-2">
                                                    <summary className="text-sm text-blue-400 cursor-pointer hover:text-blue-300">
                                                        View Schema
                                                    </summary>
                                                    <pre className="mt-2 bg-gray-900 border border-gray-700 rounded p-2 text-xs text-gray-300 overflow-x-auto">
                                                        {JSON.stringify(tool.inputSchema, null, 2)}
                                                    </pre>
                                                </details>
                                            </div>
                                            <div className="flex gap-2 ml-4">
                                                <button
                                                    onClick={() => {
                                                        setEditingTool(tool);
                                                        setIsCreating(false);
                                                    }}
                                                    className="p-2 hover:bg-gray-700 rounded-lg transition-colors"
                                                >
                                                    <Edit2 className="w-4 h-4 text-blue-400" />
                                                </button>
                                                <button
                                                    onClick={() => handleDelete(tool.name)}
                                                    className="p-2 hover:bg-gray-700 rounded-lg transition-colors"
                                                >
                                                    <Trash2 className="w-4 h-4 text-red-400" />
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </>
                )}
            </div>
        </div>
    );
}
