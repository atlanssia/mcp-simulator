# MCP Simulator 开发者文档

## 目录

1. [项目概述](#项目概述)
2. [架构设计](#架构设计)
3. [开发环境设置](#开发环境设置)
4. [代码结构](#代码结构)
5. [核心组件](#核心组件)
6. [API 参考](#api-参考)
7. [扩展开发](#扩展开发)
8. [测试指南](#测试指南)
9. [部署说明](#部署说明)

---

## 项目概述

MCP Simulator 是一个用于模拟 MCP (Model Context Protocol) 服务器的开发和测试工具。

### 技术栈

**后端**:
- **语言**: Go 1.21+
- **Web 框架**: Gin
- **并发**: Goroutines + sync.RWMutex
- **HTTP**: net/http

**前端**:
- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **样式**: Tailwind CSS
- **图标**: Lucide React
- **HTTP 客户端**: Axios

**AI 集成**:
- OpenAI 兼容接口
- 支持多个 LLM 提供商

### 项目特点

- 🏗️ 清晰的分层架构
- 🔌 可扩展的插件系统
- 🎨 现代化的 UI 设计
- 📦 单一二进制文件部署
- 🔄 实时状态更新

---

## 架构设计

### 分层架构

```
┌─────────────────────────────────────────────┐
│           Frontend (React + TS)             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │ServerList│  │ToolMgr   │  │LLMSettings│  │
│  └──────────┘  └──────────┘  └──────────┘  │
└─────────────────────────────────────────────┘
                     ↕ HTTP/JSON
┌─────────────────────────────────────────────┐
│        Application Layer (Layer 3)          │
│  ┌──────────────┐  ┌──────────────────────┐ │
│  │ Admin Handler│  │  AI Generator        │ │
│  └──────────────┘  └──────────────────────┘ │
└─────────────────────────────────────────────┘
                     ↕
┌─────────────────────────────────────────────┐
│          Core Layer (Layer 2)               │
│  ┌────────────┐  ┌──────────┐  ┌─────────┐ │
│  │VirtualServer│  │Registry  │  │Strategy │ │
│  └────────────┘  └──────────┘  └─────────┘ │
└─────────────────────────────────────────────┘
                     ↕
┌─────────────────────────────────────────────┐
│     Infrastructure Layer (Layer 1)          │
│  ┌──────────────┐  ┌──────────────────────┐ │
│  │ServerManager │  │  SSEAdapter          │ │
│  └──────────────┘  └──────────────────────┘ │
└─────────────────────────────────────────────┘
```

### 核心概念

**VirtualServer**: 虚拟 MCP 服务器实例
- 管理生命周期 (start/stop)
- 维护工具注册表
- 处理 HTTP 请求

**Registry**: 工具/资源/提示词注册表
- 存储工具定义
- 提供 CRUD 操作
- 线程安全

**MockStrategy**: 模拟数据生成策略
- Static: 静态数据
- Template: 模板渲染
- AI: AI 生成

---

## 开发环境设置

### 前置要求

- Go 1.21 或更高版本
- Node.js 18 或更高版本
- npm 或 yarn
- Git

### 克隆项目

```bash
git clone https://github.com/atlanssia/mcp-simulator.git
cd mcp-simulator
```

### 安装依赖

**后端**:
```bash
go mod download
```

**前端**:
```bash
cd web
npm install
```

### 开发模式运行

**后端**:
```bash
# 运行后端服务器
go run cmd/mcp-simulator/main.go
```

**前端**:
```bash
# 在另一个终端
cd web
npm run dev
```

前端开发服务器会在 `http://localhost:5173` 启动，并代理 API 请求到后端。

### 构建生产版本

```bash
make build
```

这会：
1. 安装前端依赖
2. 构建前端静态文件
3. 将前端嵌入 Go 二进制
4. 生成 `./mcp-simulator` 可执行文件

---

## 代码结构

```
mcp-simulator/
├── cmd/
│   └── mcp-simulator/
│       └── main.go              # 应用入口
├── internal/
│   ├── core/                    # 核心业务逻辑
│   │   ├── interfaces.go        # 接口定义
│   │   ├── config.go            # LLM 配置
│   │   ├── server.go            # 虚拟服务器实现
│   │   ├── registry.go          # 工具注册表
│   │   └── strategies.go        # Mock 策略
│   ├── infra/                   # 基础设施层
│   │   ├── manager/
│   │   │   └── server_manager.go  # 服务器管理器
│   │   └── transport/
│   │       └── sse.go           # SSE 适配器
│   └── service/                 # 应用服务层
│       ├── admin/
│       │   └── handler.go       # Admin API 处理器
│       └── ai/
│           ├── generator.go     # AI 生成器
│           ├── providers.go     # 提供商预设
│           └── dynamic_models.go # 动态模型获取
├── web/                         # 前端代码
│   ├── src/
│   │   ├── components/          # React 组件
│   │   │   ├── ServerList.tsx
│   │   │   ├── CreateServerForm.tsx
│   │   │   ├── ToolManager.tsx
│   │   │   └── LLMSettings.tsx
│   │   ├── lib/
│   │   │   ├── api.ts           # API 客户端
│   │   │   └── utils.ts         # 工具函数
│   │   ├── App.tsx              # 主应用组件
│   │   └── main.tsx             # 入口文件
│   ├── package.json
│   └── vite.config.ts
├── docs/                        # 文档
│   ├── USER_GUIDE.md            # 用户手册
│   └── DEVELOPER_GUIDE.md       # 开发者文档
├── Makefile                     # 构建脚本
├── go.mod
└── README.md
```

---

## 核心组件

### 1. VirtualServer 接口

```go
type VirtualServer interface {
    Start() error
    Stop() error
    Config() ServerConfig
    GetRegistry() Registry
    Status() string
    IsRunning() bool
}
```

**实现**: `BaseVirtualServer`

**职责**:
- 管理服务器生命周期
- 处理 HTTP 请求
- 维护工具注册表
- 状态管理

### 2. Registry 接口

```go
type Registry interface {
    RegisterTool(tool Tool) error
    GetTool(name string) (Tool, bool)
    ListTools() []Tool
    UpdateTool(name string, tool Tool) error
    DeleteTool(name string) error
}
```

**实现**: `InMemoryRegistry`

**职责**:
- 存储工具定义
- 提供线程安全的 CRUD 操作
- 支持工具查询

### 3. AI Generator

```go
type Generator struct {
    config core.LLMConfig
    client *http.Client
    mu     sync.RWMutex
}

func (g *Generator) GenerateMockData(ctx context.Context, 
    prompt string, schema map[string]interface{}) (interface{}, error)

func (g *Generator) GenerateMockResponse(ctx context.Context, 
    toolName, toolDescription string, 
    inputSchema, sampleParams map[string]interface{}) (interface{}, error)
```

**职责**:
- 调用 LLM API
- 生成工具 Schema
- 生成模拟响应数据
- 管理配置

### 4. ServerManager

```go
type ServerManager struct {
    servers map[string]core.VirtualServer
    mu      sync.RWMutex
}

func (m *ServerManager) CreateServer(config core.ServerConfig) (core.VirtualServer, error)
func (m *ServerManager) GetServer(id string) (core.VirtualServer, bool)
func (m *ServerManager) ListServers() []core.VirtualServer
```

**职责**:
- 管理所有虚拟服务器
- 提供服务器查找
- 线程安全操作

---

## API 参考

### Admin API

所有 API 端点的基础 URL: `http://localhost:8080/api`

#### 服务器管理

**列出所有服务器**
```
GET /servers
Response: ServerConfig[]
```

**创建服务器**
```
POST /servers
Body: {
  "id": string,
  "port": number,
  "name": string
}
Response: ServerConfig
```

**启动服务器**
```
POST /servers/:id/start
Response: {message: string}
```

**停止服务器**
```
POST /servers/:id/stop
Response: {message: string}
```

#### 工具管理

**列出工具**
```
GET /servers/:id/tools
Response: Tool[]
```

**创建工具**
```
POST /servers/:id/tools
Body: {
  "name": string,
  "description": string,
  "inputSchema": object
}
Response: Tool
```

**更新工具**
```
PUT /servers/:id/tools/:toolName
Body: Tool
Response: Tool
```

**删除工具**
```
DELETE /servers/:id/tools/:toolName
Response: {message: string}
```

**生成模拟响应**
```
POST /servers/:id/tools/:toolName/generate-mock
Body: {
  "params": object
}
Response: object (生成的模拟数据)
```

#### LLM 配置

**获取配置**
```
GET /config/llm
Response: LLMConfig
```

**更新配置**
```
POST /config/llm
Body: LLMConfig
Response: {message: string}
```

**列出提供商**
```
GET /config/llm/providers
Response: Provider[]
```

**列出模型**
```
GET /config/llm/models?provider=xxx&free=true
Response: ModelInfo[]
```

**动态获取模型**
```
GET /config/llm/models/dynamic?provider=xxx
Response: ModelInfo[]
```

---

## 扩展开发

### 添加新的 LLM 提供商

1. **更新提供商预设** (`internal/service/ai/providers.go`):

```go
"newprovider": {
    Name:    "New Provider",
    BaseURL: "https://api.newprovider.com/v1",
    Models: []ModelInfo{
        {Name: "model-1", DisplayName: "Model 1", Free: false},
    },
},
```

2. **如果支持动态模型获取**，添加到 `dynamic_models.go`:

```go
func FetchNewProviderModels(ctx context.Context, apiKey string) ([]ModelInfo, error) {
    // 实现 API 调用逻辑
}
```

3. **更新 Handler** (`internal/service/admin/handler.go`):

```go
case "newprovider":
    models, err = ai.FetchNewProviderModels(c.Request.Context(), config.APIKey)
```

### 添加新的 Mock 策略

1. **实现 MockStrategy 接口** (`internal/core/strategies.go`):

```go
type CustomStrategy struct {
    // 自定义字段
}

func (s *CustomStrategy) GenerateMock(tool Tool, params map[string]interface{}) (interface{}, error) {
    // 实现生成逻辑
}
```

2. **在 VirtualServer 中使用**:

```go
server := &BaseVirtualServer{
    strategy: &CustomStrategy{},
}
```

### 添加新的前端组件

1. **创建组件** (`web/src/components/NewComponent.tsx`):

```typescript
export function NewComponent() {
    return (
        <div className="...">
            {/* 组件内容 */}
        </div>
    );
}
```

2. **在 App.tsx 中使用**:

```typescript
import { NewComponent } from './components/NewComponent';

function App() {
    return (
        <div>
            <NewComponent />
        </div>
    );
}
```

### 添加新的 API 端点

1. **在 Handler 中添加方法**:

```go
func (h *Handler) NewEndpoint(c *gin.Context) {
    // 处理逻辑
    c.JSON(http.StatusOK, result)
}
```

2. **注册路由**:

```go
func (h *Handler) RegisterRoutes(r *gin.Engine) {
    api := r.Group("/api")
    {
        api.GET("/new-endpoint", h.NewEndpoint)
    }
}
```

3. **更新前端 API 客户端**:

```typescript
export const api = {
    newEndpoint: () => axiosInstance.get('/new-endpoint'),
};
```

---

## 测试指南

### 单元测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/core

# 带覆盖率
go test -cover ./...
```

### 集成测试

```bash
# 启动应用
./mcp-simulator

# 运行测试脚本
./scripts/integration-test.sh
```

### 前端测试

```bash
cd web
npm run test
```

### 手动测试

使用提供的测试脚本:

```bash
# 测试服务器创建
curl -X POST http://localhost:8080/api/servers \
  -H "Content-Type: application/json" \
  -d '{"id":"test","port":9999,"name":"Test"}'

# 测试工具创建
curl -X POST http://localhost:8080/api/servers/test/tools \
  -H "Content-Type: application/json" \
  -d '{"name":"test_tool","description":"Test","inputSchema":{}}'
```

---

## 部署说明

### 构建

```bash
make build
```

生成的二进制文件: `./mcp-simulator`

### 运行

```bash
# 直接运行
./mcp-simulator

# 后台运行
nohup ./mcp-simulator > mcp-simulator.log 2>&1 &

# 使用 systemd (Linux)
sudo systemctl start mcp-simulator
```

### 环境变量

```bash
# LLM API Key (可选)
export OPENAI_API_KEY=sk-xxxxx

# 端口配置 (默认 8080)
export PORT=8080
```

### Docker 部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
COPY --from=builder /app/mcp-simulator /usr/local/bin/
EXPOSE 8080
CMD ["mcp-simulator"]
```

```bash
# 构建镜像
docker build -t mcp-simulator .

# 运行容器
docker run -p 8080:8080 mcp-simulator
```

### 性能优化

**后端**:
- 使用 `sync.Pool` 复用对象
- 启用 HTTP/2
- 添加请求缓存

**前端**:
- 代码分割 (Code Splitting)
- 懒加载组件
- 优化打包体积

---

## 贡献指南

### 代码规范

**Go**:
- 遵循 `gofmt` 格式
- 使用 `golint` 检查
- 添加必要的注释

**TypeScript**:
- 遵循 ESLint 规则
- 使用 Prettier 格式化
- 类型定义完整

### 提交规范

```
<type>(<scope>): <subject>

<body>

<footer>
```

类型:
- `feat`: 新功能
- `fix`: 修复
- `docs`: 文档
- `style`: 格式
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

### Pull Request 流程

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到 Fork
5. 创建 Pull Request

---

## 附录

### 依赖列表

**Go 依赖**:
```
github.com/gin-gonic/gin
```

**npm 依赖**:
```
react
typescript
vite
tailwindcss
axios
lucide-react
```

### 性能指标

- 启动时间: <1s
- 内存占用: ~50MB
- 并发请求: 1000+ req/s
- 响应时间: <10ms (不含 AI 调用)

### 已知问题

1. 配置不持久化 (计划在 v1.1 解决)
2. 不支持服务器删除 (计划在 v1.1 解决)
3. MCP 协议支持不完整 (持续开发中)

---

**版本**: 1.0.0  
**更新日期**: 2025-11-20  
**适用对象**: MCP Simulator 开发者
