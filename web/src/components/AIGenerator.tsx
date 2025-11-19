import { useState } from 'react';
import { Sparkles } from 'lucide-react';
import { api } from '../lib/api';
import { cn } from '../lib/utils';

export function AIGenerator() {
    const [prompt, setPrompt] = useState('');
    const [result, setResult] = useState('');
    const [loading, setLoading] = useState(false);

    const handleGenerate = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setResult('');
        try {
            const { data } = await api.generateMock({
                prompt,
                schema: {
                    type: 'object',
                    properties: {
                        data: { type: 'array' }
                    }
                }
            });

            if (data.content && data.content.length > 0) {
                setResult(data.content[0].text);
            }
        } catch (error: any) {
            setResult(`Error: ${error.response?.data?.error || error.message}`);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-lg p-6">
            <h2 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
                <Sparkles className="w-6 h-6 text-yellow-400" />
                AI Mock Generator
            </h2>
            <form onSubmit={handleGenerate} className="space-y-4">
                <div>
                    <label className="block text-sm font-medium text-gray-300 mb-2">
                        Describe what you want to generate
                    </label>
                    <textarea
                        value={prompt}
                        onChange={(e) => setPrompt(e.target.value)}
                        required
                        rows={3}
                        className="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-yellow-500 resize-none"
                        placeholder="Generate a list of 5 recent orders with id, customer name, and total amount"
                    />
                </div>
                <button
                    type="submit"
                    disabled={loading}
                    className={cn(
                        "w-full flex items-center justify-center gap-2 px-4 py-2 rounded-lg font-medium transition-colors",
                        "bg-gradient-to-r from-yellow-600 to-orange-600 hover:from-yellow-700 hover:to-orange-700 text-white",
                        "disabled:opacity-50 disabled:cursor-not-allowed"
                    )}
                >
                    <Sparkles className="w-5 h-5" />
                    {loading ? 'Generating...' : 'Generate Mock Data'}
                </button>
            </form>

            {result && (
                <div className="mt-4">
                    <label className="block text-sm font-medium text-gray-300 mb-2">
                        Result
                    </label>
                    <pre className="bg-gray-900 border border-gray-700 rounded-lg p-4 text-sm text-green-400 overflow-x-auto">
                        {result}
                    </pre>
                </div>
            )}
        </div>
    );
}
