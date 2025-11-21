import { useState, useEffect } from 'react';
import { type Tool, type ModelInfo, type GenerationParams, api } from '../lib/api';
import { Plus, Trash2, Edit2, Save, X, Settings2, ChevronDown, ChevronUp, FlaskConical } from 'lucide-react';

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
    const [testingTool, setTestingTool] = useState<Tool | null>(null);
    const [testParams, setTestParams] = useState('{}');
    const [mockResult, setMockResult] = useState<any>(null);
    const [testGenerating, setTestGenerating] = useState(false);

    // Generation settings
    const [showGenSettings, setShowGenSettings] = useState(false);
    const [models, setModels] = useState<ModelInfo[]>([]);
    const [genParams, setGenParams] = useState<GenerationParams>({
        temperature: 0.7,
        system_prompt: '',
        model: ''
    });
    const [generating, setGenerating] = useState(false);

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

    const fetchModels = async () => {
        try {
            const configRes = await api.getLLMConfig();
            const provider = configRes.data.provider;

            // Try dynamic first
            try {
                const response = await api.listDynamicModels(provider);
                if (response.data.length > 0) {
                    setModels(response.data);
                    setGenParams(prev => ({ ...prev, model: response.data[0].name }));
                    return;
                }
            } catch (e) {
                console.warn('Failed to fetch dynamic models:', e);
            }

            // Fallback to static
            const response = await api.listModels(provider);
            setModels(response.data);
            if (response.data.length > 0) {
                setGenParams(prev => ({ ...prev, model: response.data[0].name }));
            }
        } catch (error) {
            console.error('Failed to fetch models:', error);
        }
    };

    useEffect(() => {
        fetchTools();
        fetchModels();
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

    const handleGenerateSchema = async () => {
        if (!editingTool?.description) {
            alert('Please enter a description first');
            return;
        }

        setGenerating(true);
        try {
            const response = await api.generateSchema({
                description: editingTool.description,
                params: genParams
            });

            if (response.data && typeof response.data === 'object') {
                setEditingTool({ ...editingTool, inputSchema: response.data as Record<string, unknown> });
            }
        } catch (error: any) {
            console.error('Failed to generate schema:', error);
            const errorMessage = error.response?.data?.error || error.message || 'Unknown error';
            alert(`Failed to generate schema: ${errorMessage}`);
        } finally {
            setGenerating(false);
        }
    };

    const handleTestTool = async () => {
        if (!testingTool) return;

        setTestGenerating(true);
        setMockResult(null);
        try {
            let params = {};
            try {
                params = JSON.parse(testParams);
            } catch (e) {
                alert('Invalid JSON in parameters');
                return;
            }

            const response = await api.generateToolMock(serverId, testingTool.name, params, genParams);
            setMockResult(response.data);
        } catch (error: any) {
            console.error('Failed to generate mock data:', error);
            const errorMessage = error.response?.data?.error || error.message || 'Unknown error';
            alert(`Failed to generate mock data: ${errorMessage}`);
        } finally {
            setTestGenerating(false);
        }
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
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-start justify-center z-50 pt-10 pb-10 overflow-y-auto">
            <div className="bg-gray-900 rounded-lg p-6 max-w-4xl w-full mx-4 my-auto relative">
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
                                </div>
                            </div>

                            {/* Generation Settings */}
                            <div className="border border-gray-700 rounded-lg p-3 bg-gray-900/30">
                                <button
                                    onClick={() => setShowGenSettings(!showGenSettings)}
                                    className="flex items-center justify-between w-full text-sm font-medium text-gray-300 hover:text-white"
                                >
                                    <div className="flex items-center gap-2">
                                        <Settings2 className="w-4 h-4" />
                                        Generation Settings
                                    </div>
                                    {showGenSettings ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                                </button>

                                {showGenSettings && (
                                    <div className="mt-3 space-y-3 pt-3 border-t border-gray-700">
                                        <div>
                                            <label className="block text-xs text-gray-400 mb-1">Model</label>
                                            <select
                                                value={genParams.model}
                                                onChange={(e) => setGenParams({ ...genParams, model: e.target.value })}
                                                className="w-full bg-gray-800 border border-gray-600 rounded px-2 py-1 text-sm text-white"
                                            >
                                                <option value="">Default</option>
                                                {models.map(m => (
                                                    <option key={m.name} value={m.name}>{m.display_name}</option>
                                                ))}
                                            </select>
                                        </div>
                                        <div>
                                            <label className="block text-xs text-gray-400 mb-1">
                                                Temperature: {genParams.temperature}
                                            </label>
                                            <input
                                                type="range"
                                                min="0"
                                                max="2"
                                                step="0.1"
                                                value={genParams.temperature}
                                                onChange={(e) => setGenParams({ ...genParams, temperature: parseFloat(e.target.value) })}
                                                className="w-full"
                                            />
                                        </div>
                                        <div>
                                            <label className="block text-xs text-gray-400 mb-1">System Prompt (Optional)</label>
                                            <textarea
                                                value={genParams.system_prompt}
                                                onChange={(e) => setGenParams({ ...genParams, system_prompt: e.target.value })}
                                                className="w-full bg-gray-800 border border-gray-600 rounded px-2 py-1 text-sm text-white"
                                                rows={2}
                                                placeholder="Custom system prompt..."
                                            />
                                        </div>
                                    </div>
                                )}

                                <div className="mt-3 flex justify-end">
                                    <button
                                        onClick={handleGenerateSchema}
                                        disabled={generating || !editingTool.description}
                                        className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg text-white text-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                    >
                                        {generating ? 'Generating...' : '✨ Generate Schema'}
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
                                                        setTestingTool(tool);
                                                        setTestParams('{}');
                                                        setMockResult(null);
                                                    }}
                                                    className="p-2 hover:bg-gray-700 rounded-lg transition-colors"
                                                    title="Test Tool"
                                                >
                                                    <FlaskConical className="w-4 h-4 text-green-400" />
                                                </button>
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

                {/* Test Tool Modal */}
                {testingTool && (
                    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
                        <div className="bg-gray-900 rounded-lg p-6 max-w-3xl w-full mx-4 max-h-[80vh] overflow-y-auto">
                            <div className="flex items-center justify-between mb-4">
                                <h3 className="text-xl font-bold text-white flex items-center gap-2">
                                    <FlaskConical className="w-5 h-5 text-green-400" />
                                    Test Tool: {testingTool.name}
                                </h3>
                                <button
                                    onClick={() => {
                                        setTestingTool(null);
                                        setMockResult(null);
                                    }}
                                    className="p-2 hover:bg-gray-800 rounded-lg transition-colors"
                                >
                                    <X className="w-5 h-5 text-gray-400" />
                                </button>
                            </div>

                            <div className="space-y-4">
                                <div>
                                    <p className="text-sm text-gray-400 mb-3">{testingTool.description}</p>
                                </div>

                                <div>
                                    <label className="block text-sm font-medium text-gray-300 mb-1">Sample Parameters (JSON)</label>
                                    <textarea
                                        value={testParams}
                                        onChange={(e) => setTestParams(e.target.value)}
                                        className="w-full bg-gray-800 border border-gray-600 rounded-lg px-3 py-2 text-white font-mono text-sm"
                                        rows={4}
                                        placeholder='{"city": "深圳"}'
                                    />
                                </div>

                                {/* Generation Settings */}
                                <div className="border border-gray-700 rounded-lg p-3 bg-gray-900/30">
                                    <div className="space-y-3">
                                        <div>
                                            <label className="block text-xs text-gray-400 mb-1">Custom Instructions (Optional)</label>
                                            <textarea
                                                value={genParams.system_prompt}
                                                onChange={(e) => setGenParams({ ...genParams, system_prompt: e.target.value })}
                                                className="w-full bg-gray-800 border border-gray-600 rounded px-2 py-1 text-sm text-white"
                                                rows={2}
                                                placeholder="e.g., 生成5条患者体征数据，包括体温、心率、血压..."
                                            />
                                        </div>
                                        <div className="grid grid-cols-2 gap-3">
                                            <div>
                                                <label className="block text-xs text-gray-400 mb-1">Model</label>
                                                <select
                                                    value={genParams.model}
                                                    onChange={(e) => setGenParams({ ...genParams, model: e.target.value })}
                                                    className="w-full bg-gray-800 border border-gray-600 rounded px-2 py-1 text-sm text-white"
                                                >
                                                    {models.map(m => (
                                                        <option key={m.name} value={m.name}>{m.display_name}</option>
                                                    ))}
                                                </select>
                                            </div>
                                            <div>
                                                <label className="block text-xs text-gray-400 mb-1">Temperature: {genParams.temperature}</label>
                                                <input
                                                    type="range"
                                                    min="0"
                                                    max="2"
                                                    step="0.1"
                                                    value={genParams.temperature}
                                                    onChange={(e) => setGenParams({ ...genParams, temperature: parseFloat(e.target.value) })}
                                                    className="w-full"
                                                />
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                <button
                                    onClick={handleTestTool}
                                    disabled={testGenerating}
                                    className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-green-600 hover:bg-green-700 rounded-lg text-white font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    {testGenerating ? 'Generating...' : '🧪 Generate Sample Data'}
                                </button>

                                {mockResult && (
                                    <div>
                                        <label className="block text-sm font-medium text-gray-300 mb-2">Generated Mock Data:</label>
                                        <pre className="bg-gray-950 border border-gray-700 rounded-lg p-4 text-sm text-green-400 overflow-x-auto">
                                            {JSON.stringify(mockResult, null, 2)}
                                        </pre>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
