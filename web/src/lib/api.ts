import axios from 'axios';


export interface ServerConfig {
    id: string;
    name: string;
    port: number;
    status?: string; // "stopped", "starting", "running", "error"
}

export interface GenerateRequest {
    prompt: string;
    schema: Record<string, unknown>;
}

export interface GenerationParams {
    model?: string;
    temperature?: number;
    system_prompt?: string;
    max_tokens?: number;
}

export interface GenerateSchemaRequest {
    description: string;
    params: GenerationParams;
}

export interface Tool {
    name: string;
    description: string;
    inputSchema: Record<string, unknown>;
}

export interface LLMConfig {
    provider: string;
    api_key: string;
    base_url: string;
    model: string;
}

export interface Provider {
    id: string;
    name: string;
    requires_api_key: boolean;
    default_base_url: string;
    supports_models: boolean;
    description: string;
}

export interface ModelInfo {
    name: string;
    display_name: string;
    free: boolean;
}

export interface CallToolResult {
    content: Array<{
        type: string;
        text: string;
    }>;
    isError?: boolean;
}

const axiosInstance = axios.create({
    baseURL: '/api',
});

export const api = {
    // Server management
    listServers: () => axiosInstance.get<ServerConfig[]>('/servers'),
    createServer: (config: ServerConfig) => axiosInstance.post<ServerConfig>('/servers', config),
    startServer: (id: string) => axiosInstance.post(`/servers/${id}/start`),
    stopServer: (id: string) => axiosInstance.post(`/servers/${id}/stop`),

    // Tool management
    listTools: (serverId: string) => axiosInstance.get<Tool[]>(`/servers/${serverId}/tools`),
    createTool: (serverId: string, tool: Tool) => axiosInstance.post<Tool>(`/servers/${serverId}/tools`, tool),
    updateTool: (serverId: string, toolName: string, tool: Tool) => axiosInstance.put<Tool>(`/servers/${serverId}/tools/${toolName}`, tool),
    deleteTool: (serverId: string, toolName: string) => axiosInstance.delete(`/servers/${serverId}/tools/${toolName}`),

    // LLM configuration
    getLLMConfig: () => axiosInstance.get<LLMConfig>('/config/llm'),
    updateLLMConfig: (config: LLMConfig) => axiosInstance.post('/config/llm', config),
    listProviders: () => axiosInstance.get<Provider[]>('/config/llm/providers'),
    listModels: (provider: string, freeOnly?: boolean) =>
        axiosInstance.get<ModelInfo[]>('/config/llm/models', {
            params: { provider, free: freeOnly || undefined }
        }),
    listDynamicModels: (provider: string, freeOnly?: boolean) =>
        axiosInstance.get<ModelInfo[]>('/config/llm/models/dynamic', {
            params: { provider, free: freeOnly || undefined }
        }),

    // AI generation
    generateMock: (request: GenerateRequest) => axiosInstance.post<CallToolResult>('/ai/generate', request),
    generateSchema: (request: GenerateSchemaRequest) => axiosInstance.post<Record<string, unknown>>('/ai/generate-schema', request),
    generateToolMock: (serverId: string, toolName: string, params: any, generation: GenerationParams) =>
        axiosInstance.post<any>(`/servers/${serverId}/tools/${toolName}/generate-mock`, { params, generation }),
};
