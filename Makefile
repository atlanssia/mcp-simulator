.PHONY: all build clean install-web build-web build-server run dev help

# Default target
all: build

# Help target
help:
	@echo "MCP Simulator - Build Targets"
	@echo "=============================="
	@echo "make build        - Build complete application (web + server) into single binary"
	@echo "make build-linux  - Build for Linux/amd64"
	@echo "make clean        - Clean all build artifacts"
	@echo "make install-web  - Install web dependencies"
	@echo "make build-web    - Build web frontend only"
	@echo "make build-server - Build Go server only (requires web/dist)"
	@echo "make run          - Build and run the application"
	@echo "make dev          - Run web dev server (for development)"

# Install web dependencies
install-web:
	@echo "Installing web dependencies..."
	cd web && npm install

# Build web frontend
build-web: install-web
	@echo "Building web frontend..."
	cd web && npm run build

# Build Go server with embedded frontend
build-server:
	@echo "Building Go server with embedded frontend..."
	go build -o mcp-simulator .

# Complete build: web + server
build: build-web build-server
	@echo "Build complete! Binary: ./mcp-simulator"

# Build for Linux (amd64)
build-linux: build-web
	@echo "Building for Linux/amd64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o mcp-simulator-linux-amd64 .
	@echo "Build complete! Binary: ./mcp-simulator-linux-amd64"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f mcp-simulator
	rm -f mcp-simulator-linux-amd64
	rm -rf web/dist
	rm -rf web/node_modules
	@echo "Clean complete!"

# Build and run
run: build
	@echo "Starting MCP Simulator..."
	./mcp-simulator

# Development mode (web only)
dev:
	@echo "Starting web dev server..."
	cd web && npm run dev
