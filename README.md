# MCP Simulator

A comprehensive Model Context Protocol (MCP) simulator built in Go, allowing developers to mock multiple MCP servers with dynamic configuration and AI-generated data.

## Features

- **Multi-Server Support**: Run multiple virtual MCP servers on different ports
- **Dynamic Configuration**: Add tools, resources, and prompts at runtime
- **AI-Powered Mocking**: Generate realistic mock data using OpenAI/Anthropic APIs
- **SSE Support**: Built-in Server-Sent Events adapter for MCP communication
- **Web UI**: Modern React-based interface for managing servers and generating mocks
- **Single Binary**: Frontend embedded in Go binary for easy deployment

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 18+ and npm
- OpenAI API key (for AI mock generation)

### Build

```bash
# Build everything (web + server) into a single binary
make build

# Or build manually
cd web && npm install && npm run build && cd ..
go build -o mcp-simulator ./cmd/mcp-simulator
```

### Run

```bash
# Set your OpenAI API key
export OPENAI_API_KEY=your-api-key-here

# Run the simulator
./mcp-simulator
```

The application will start on `http://localhost:8080`

## Development

### Web Development

```bash
# Run web dev server with hot reload
make dev
```

### Build Targets

- `make build` - Build complete application (web + server)
- `make clean` - Clean all build artifacts
- `make install-web` - Install web dependencies
- `make build-web` - Build web frontend only
- `make build-server` - Build Go server only
- `make run` - Build and run the application
- `make help` - Show all available targets

## API Usage

### Create a Virtual Server

```bash
curl -X POST http://localhost:8080/api/servers \
  -H "Content-Type: application/json" \
  -d '{"id": "srv1", "name": "GitHub Mock", "port": 8081}'
```

### Start a Server

```bash
curl -X POST http://localhost:8080/api/servers/srv1/start
```

### Generate Mock Data with AI

```bash
curl -X POST http://localhost:8080/api/ai/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Generate a list of 5 recent orders",
    "schema": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": {"type": "string"},
          "total": {"type": "number"}
        }
      }
    }
  }'
```

## Architecture

The system follows Clean Architecture principles:

- **Infrastructure Layer**: Server management, transport adapters (SSE, Stdio), configuration
- **Domain/Engine Layer**: Core logic, dynamic registry, smart router, mock strategies
- **Application Layer**: Admin API, request inspection, AI service integration
- **Web Layer**: React UI with Tailwind CSS

## License

MIT