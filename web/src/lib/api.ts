import axios from 'axios';

const API_BASE = '/api';

export interface ServerConfig {
    id: string;
    name: string;
    port: number;
}

export interface GenerateRequest {
    prompt: string;
    schema: Record<string, any>;
}

export interface CallToolResult {
    content: Array<{
        type: string;
        text: string;
    }>;
    isError?: boolean;
}

export const api = {
    // Server management
    listServers: () => axios.get<ServerConfig[]>(`${API_BASE}/servers`),
    createServer: (config: ServerConfig) => axios.post<ServerConfig>(`${API_BASE}/servers`, config),
    startServer: (id: string) => axios.post(`${API_BASE}/servers/${id}/start`),
    stopServer: (id: string) => axios.post(`${API_BASE}/servers/${id}/stop`),

    // AI generation
    generateMock: (req: GenerateRequest) => axios.post<CallToolResult>(`${API_BASE}/ai/generate`, req),
};
