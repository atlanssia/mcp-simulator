import { useState, useEffect } from 'react';
import { type LLMConfig, type Provider, type ModelInfo, api } from '../lib/api';
import { Settings, X, Save, RefreshCw } from 'lucide-react';

interface LLMSettingsProps {
    onClose: () => void;
}

export function LLMSettings({ onClose }: LLMSettingsProps) {
    const [config, setConfig] = useState<LLMConfig>({
        provider: 'openai',
        api_key: '',
        base_url: 'https://api.openai.com/v1',
        model: '',
    });
    const [providers, setProviders] = useState<Provider[]>([]);
    const [providerInfo, setProviderInfo] = useState<Provider | null>(null);
    const [models, setModels] = useState<ModelInfo[]>([]);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [fetchingModels, setFetchingModels] = useState(false);

    useEffect(() => {
        loadConfig();
        loadProviders();
    }, []);

    useEffect(() => {
        if (config.provider && providers.length > 0) {
            const info = providers.find(p => p.id === config.provider);
            setProviderInfo(info || null);

            // Clear model list when provider changes
            setModels([]);

            // If provider supports models, fetch them
            if (info?.supports_models) {
                fetchModels(info.id);
            }
        }
    }, [config.provider, providers]);

    const loadConfig = async () => {
        try {
            const response = await api.getLLMConfig();
            setConfig(response.data);
        } catch (error) {
            console.error('Failed to load config:', error);
        } finally {
            setLoading(false);
        }
    };

    const loadProviders = async () => {
        try {
            const response = await api.listProviders();
            setProviders(response.data);
        } catch (error) {
            console.error('Failed to load providers:', error);
        }
    };

    const fetchModels = async (providerId: string) => {
        setFetchingModels(true);
        try {
            // Try dynamic first
            try {
                const response = await api.listDynamicModels(providerId);
                if (response.data.length > 0) {
                    setModels(response.data);
                    return;
                }
            } catch (e) {
                console.warn('Failed to fetch dynamic models:', e);
            }

            // Fallback to static (though we removed most static presets, some might remain or be re-added)
            const response = await api.listModels(providerId);
            setModels(response.data);
        } catch (error) {
            console.error('Failed to fetch models:', error);
        } finally {
            setFetchingModels(false);
        }
    };

    const handleProviderChange = (providerId: string) => {
        const provider = providers.find(p => p.id === providerId);
        if (provider) {
            setConfig(prev => ({
                ...prev,
                provider: providerId,
                base_url: provider.default_base_url,
                model: '', // Reset model when provider changes
            }));
            setProviderInfo(provider);
        }
    };

    const handleSave = async () => {
        setSaving(true);
        try {
            await api.updateLLMConfig(config);
            alert('配置已保存！');

            // Refresh models if supported
            if (providerInfo?.supports_models) {
                await fetchModels(config.provider);
            }

            onClose();
        } catch (error) {
            console.error('Failed to save config:', error);
            alert('保存失败，请检查配置');
        } finally {
            setSaving(false);
        }
    };

    if (loading) {
        return (
            <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50">
                <div className="bg-gray-900 rounded-lg p-8">
                    <div className="text-white">加载配置中...</div>
                </div>
            </div>
        );
    }

    return (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
            <div className="bg-gray-900 rounded-lg p-6 max-w-3xl w-full max-h-[90vh] overflow-y-auto">
                <div className="flex items-center justify-between mb-6">
                    <div className="flex items-center gap-3">
                        <Settings className="w-6 h-6 text-blue-400" />
                        <h2 className="text-2xl font-bold text-white">LLM 配置</h2>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-2 hover:bg-gray-800 rounded-lg transition-colors"
                    >
                        <X className="w-5 h-5 text-gray-400" />
                    </button>
                </div>

                <div className="space-y-6">
                    {/* Provider Selection */}
                    <div>
                        <label className="block text-sm font-medium text-gray-300 mb-2">
                            AI 提供商
                        </label>
                        <select
                            value={config.provider}
                            onChange={(e) => handleProviderChange(e.target.value)}
                            className="w-full bg-gray-800 border border-gray-600 rounded-lg px-4 py-2 text-white"
                        >
                            {providers.map((provider) => (
                                <option key={provider.id} value={provider.id}>
                                    {provider.name}
                                </option>
                            ))}
                        </select>
                    </div>

                    {/* API Key - Optional based on provider */}
                    <div>
                        <label className="block text-sm font-medium text-gray-300 mb-2">
                            API Key {providerInfo?.requires_api_key ? '' : '(可选)'}
                        </label>
                        <input
                            type="password"
                            value={config.api_key}
                            onChange={(e) => setConfig({ ...config, api_key: e.target.value })}
                            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                            placeholder={providerInfo?.requires_api_key ? "必填" : "可选"}
                        />
                        {providerInfo?.requires_api_key && (
                            <p className="mt-1 text-xs text-gray-400">
                                此提供商需要 API Key
                            </p>
                        )}
                    </div>

                    {/* Base URL - Optional */}
                    <div>
                        <label className="block text-sm font-medium text-gray-300 mb-2">
                            Base URL (可选)
                        </label>
                        <input
                            type="text"
                            value={config.base_url}
                            onChange={(e) => setConfig({ ...config, base_url: e.target.value })}
                            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                            placeholder={providerInfo?.default_base_url || "自定义 API 端点"}
                        />
                        <p className="mt-1 text-xs text-gray-400">
                            留空使用默认: {providerInfo?.default_base_url}
                        </p>
                    </div>

                    {/* Default Model Selection/Input */}
                    <div>
                        <label className="block text-sm font-medium text-gray-300 mb-2">
                            默认模型 (Default Model)
                        </label>

                        {providerInfo?.supports_models ? (
                            <div className="space-y-2">
                                <div className="flex gap-2">
                                    <select
                                        value={config.model}
                                        onChange={(e) => setConfig({ ...config, model: e.target.value })}
                                        className="flex-1 bg-gray-800 border border-gray-600 rounded-lg px-4 py-2 text-white"
                                        disabled={fetchingModels}
                                    >
                                        <option value="">选择模型...</option>
                                        {models.map((model) => (
                                            <option key={model.name} value={model.name}>
                                                {model.display_name}
                                            </option>
                                        ))}
                                    </select>
                                    <button
                                        onClick={() => config.provider && fetchModels(config.provider)}
                                        className="p-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-white transition-colors"
                                        title="刷新模型列表"
                                        disabled={fetchingModels}
                                    >
                                        <RefreshCw className={`w-5 h-5 ${fetchingModels ? 'animate-spin' : ''}`} />
                                    </button>
                                </div>
                                <p className="text-xs text-gray-400">
                                    从提供商获取的可用模型列表
                                </p>
                            </div>
                        ) : (
                            <div className="space-y-2">
                                <input
                                    type="text"
                                    value={config.model}
                                    onChange={(e) => setConfig({ ...config, model: e.target.value })}
                                    className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white"
                                    placeholder="输入模型名称 (例如: llama3:8b, qwen2.5-7b)"
                                />
                                <p className="text-xs text-gray-400">
                                    此提供商不支持自动获取模型列表，请手动输入模型名称
                                </p>
                            </div>
                        )}
                    </div>

                    {/* Info Box */}
                    {providerInfo && (
                        <div className="bg-blue-900/30 border border-blue-700 rounded-lg p-3">
                            <p className="text-sm text-blue-300">
                                <strong>说明:</strong> {providerInfo.description}
                            </p>
                            {providerInfo.supports_models && (
                                <p className="text-xs text-blue-400 mt-1">
                                    ✓ 支持动态模型列表获取
                                </p>
                            )}
                        </div>
                    )}
                </div>

                {/* Actions */}
                <div className="flex gap-3 pt-4 border-t border-gray-700 mt-6">
                    <button
                        onClick={handleSave}
                        disabled={saving}
                        className="flex items-center gap-2 px-6 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg text-white transition-colors disabled:opacity-50"
                    >
                        <Save className="w-4 h-4" />
                        {saving ? '保存中...' : '保存配置'}
                    </button>
                    <button
                        onClick={onClose}
                        className="px-6 py-2 bg-gray-600 hover:bg-gray-500 rounded-lg text-white transition-colors ml-auto"
                    >
                        取消
                    </button>
                </div>
            </div>
        </div>
    );
}
