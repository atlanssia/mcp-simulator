import { useState, useEffect } from 'react';
import { type LLMConfig, type Provider, type ModelInfo, api } from '../lib/api';
import { Settings, X, Save } from 'lucide-react';

interface LLMSettingsProps {
    onClose: () => void;
}

const DEFAULT_SYSTEM_PROMPT = `You are an expert API designer. Generate a JSON schema for the following tool description.

Tool Description: {description}

Requirements:
1. Return ONLY valid JSON schema (no markdown, no explanations)
2. Use "type": "object" as the root
3. Define "properties" with appropriate types (string, number, boolean, array, object)
4. Include "description" for each property
5. Specify "required" array for mandatory fields
6. Use clear, descriptive property names

Example format:
{
  "type": "object",
  "properties": {
    "param1": {
      "type": "string",
      "description": "Description of param1"
    }
  },
  "required": ["param1"]
}`;

export function LLMSettings({ onClose }: LLMSettingsProps) {
    const [config, setConfig] = useState<LLMConfig>({
        provider: 'openai',
        api_key: '',
        base_url: 'https://api.openai.com/v1',
        model: 'gpt-4o-mini',
        temperature: 0.7,
        system_prompt: DEFAULT_SYSTEM_PROMPT,
    });
    const [providers, setProviders] = useState<Provider[]>([]);
    const [models, setModels] = useState<ModelInfo[]>([]);
    const [freeOnly, setFreeOnly] = useState(false);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        loadConfig();
        loadProviders();
    }, []);

    useEffect(() => {
        if (config.provider) {
            loadModels(config.provider, freeOnly);
        }
    }, [config.provider, freeOnly]);

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

    const loadModels = async (provider: string, free: boolean) => {
        try {
            // Try dynamic first for supported providers
            if ((provider === 'siliconflow' && config.api_key) || provider === 'modelscope') {
                try {
                    const response = await api.listDynamicModels(provider, free);
                    setModels(response.data);

                    // Auto-select first model if current model not in list
                    if (response.data.length > 0 && !response.data.find(m => m.name === config.model)) {
                        setConfig(prev => ({ ...prev, model: response.data[0].name }));
                    }
                    return;
                } catch (error) {
                    console.warn('Dynamic model fetch failed, falling back to static:', error);
                }
            }

            // Fallback to static presets
            const response = await api.listModels(provider, free);
            setModels(response.data);

            // Auto-select first model if current model not in list
            if (response.data.length > 0 && !response.data.find(m => m.name === config.model)) {
                setConfig(prev => ({ ...prev, model: response.data[0].name }));
            }
        } catch (error) {
            console.error('Failed to load models:', error);
        }
    };

    const handleProviderChange = (providerId: string) => {
        const provider = providers.find(p => p.id === providerId);
        if (provider) {
            setConfig(prev => ({
                ...prev,
                provider: providerId,
                base_url: provider.base_url,
            }));
        }
    };

    const handleSave = async () => {
        setSaving(true);
        try {
            await api.updateLLMConfig(config);
            alert('配置已保存！');
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

                    {/* API Key */}
                    <div>
                        <label className="block text-sm font-medium text-gray-300 mb-2">
                            API Key
                        </label>
                        <input
                            type="password"
                            value={config.api_key}
                            onChange={(e) => setConfig({ ...config, api_key: e.target.value })}
                            className="w-full bg-gray-800 border border-gray-600 rounded-lg px-4 py-2 text-white"
                            placeholder="sk-..."
                        />
                    </div>

                    {/* Base URL */}
                    <div>
                        <label className="block text-sm font-medium text-gray-300 mb-2">
                            API 端点
                        </label>
                        <input
                            type="text"
                            value={config.base_url}
                            onChange={(e) => setConfig({ ...config, base_url: e.target.value })}
                            className="w-full bg-gray-800 border border-gray-600 rounded-lg px-4 py-2 text-white"
                            placeholder="https://api.example.com/v1"
                        />
                    </div>

                    {/* Model Selection with Free Filter */}
                    <div>
                        <div className="flex items-center justify-between mb-2">
                            <label className="block text-sm font-medium text-gray-300">
                                模型
                            </label>
                            <label className="flex items-center gap-2 text-sm text-gray-400">
                                <input
                                    type="checkbox"
                                    checked={freeOnly}
                                    onChange={(e) => setFreeOnly(e.target.checked)}
                                    className="rounded"
                                />
                                仅显示免费模型
                            </label>
                        </div>
                        <select
                            value={config.model}
                            onChange={(e) => setConfig({ ...config, model: e.target.value })}
                            className="w-full bg-gray-800 border border-gray-600 rounded-lg px-4 py-2 text-white"
                        >
                            {models.map((model) => (
                                <option key={model.name} value={model.name}>
                                    {model.display_name} {model.free && '🆓'}
                                </option>
                            ))}
                        </select>
                    </div>

                    {/* Temperature */}
                    <div>
                        <label className="block text-sm font-medium text-gray-300 mb-2">
                            Temperature: {config.temperature}
                        </label>
                        <input
                            type="range"
                            min="0"
                            max="2"
                            step="0.1"
                            value={config.temperature}
                            onChange={(e) => setConfig({ ...config, temperature: parseFloat(e.target.value) })}
                            className="w-full"
                        />
                        <div className="flex justify-between text-xs text-gray-500 mt-1">
                            <span>精确 (0.0)</span>
                            <span>平衡 (1.0)</span>
                            <span>创造 (2.0)</span>
                        </div>
                    </div>

                    {/* System Prompt */}
                    <div>
                        <label className="block text-sm font-medium text-gray-300 mb-2">
                            系统提示词模板
                        </label>
                        <textarea
                            value={config.system_prompt}
                            onChange={(e) => setConfig({ ...config, system_prompt: e.target.value })}
                            className="w-full bg-gray-800 border border-gray-600 rounded-lg px-4 py-2 text-white font-mono text-sm"
                            rows={10}
                            placeholder="使用 {description} 作为占位符"
                        />
                        <p className="text-xs text-gray-500 mt-1">
                            提示：使用 <code className="bg-gray-800 px-1 rounded">{'{description}'}</code> 作为工具描述的占位符
                        </p>
                    </div>

                    {/* Action Buttons */}
                    <div className="flex gap-3 pt-4 border-t border-gray-700">
                        <button
                            onClick={handleSave}
                            disabled={saving}
                            className="flex items-center gap-2 px-6 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg text-white transition-colors disabled:opacity-50"
                        >
                            <Save className="w-4 h-4" />
                            {saving ? '保存中...' : '保存配置'}
                        </button>
                        <button
                            onClick={() => setConfig(prev => ({ ...prev, system_prompt: DEFAULT_SYSTEM_PROMPT }))}
                            className="px-6 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-white transition-colors"
                        >
                            重置提示词
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
        </div>
    );
}
