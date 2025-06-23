# openapi-to-mcp

[简体中文](README_zh.md)

`openapi-to-mcp` is a Go implementation for converting OpenAPI specifications into MCP (Model Context Protocol) tools.

## Overview

This project aims to provide a flexible framework that allows developers to automatically expose existing OpenAPI interfaces as MCP tools through simple configuration, enabling AI models to call these interfaces directly.

## Features

- **OpenAPI to MCP Tool Conversion**: Automatically parses OpenAPI specifications and creates corresponding MCP tools based on the definitions.
- **Multiple Transport Support**: Supports `stdio` (Standard I/O), `sse` (Server-Sent Events), and `stream` (HTTP Stream) as transport protocols for MCP communication.
- **State Tracking & Authentication**: Supports cookie-based state tracking and JWT (JSON Web Token) handling.
- **Rate Limiting**: Built-in rate limiting to prevent high-frequency calls to the Large Language Model (LLM).
- **Environment Variable Configuration**: Flexible configuration via `.env` file or system environment variables.

## Quick Start

### 1. Configure Environment Variables

Create a `.env` file or set the following environment variables:

```dotenv
# MCP transport: stdio, sse, stream (default: stdio)
MCP_TRANSPORT="stdio"
# MCP BASE URL
MCP_BASE_URL="http://localhost:8080"

# OpenAPI specification file path (can be a local file path or a URL)
OPENAPI_SRC="./example/openapi.yaml"
# Base URL for API requests
OPENAPI_BASE_URL=

# Extra HTTP headers (JSON format), e.g., '{"X-API-Key": "your-api-key"}'
EXTRA_HEADERS='{"X-API-Key": "your-api-key"}'

# Use cookies (true/false, default: true)
USE_COOKIE=true

# Output logs to standard output (true/false, default: false)
LOG_OUTPUT=false

# Rate limit: allowed requests per second (default: 1)
RATE_LIMIT_PER_SECOND=1

# Authorization header, e.g., "Basic xxxx"
AUTHORIZATION_HEADERS="Basic xxxx"
```

### 2. Run the Project

```bash
git clone https://github.com/constellation39/openapi-to-mcp
cd openapi-to-mcp
copy .env.example .env

go run main.go
#or
go build .
./openapi-to-mcp
```

### 3. Connect with an MCP Client

Depending on the `MCP_TRANSPORT` you've chosen, connect to `openapi-to-mcp` using a corresponding MCP client.

**Stdio Mode (Default)**:

If `MCP_TRANSPORT` is set to `stdio`, you can configure your MCP client to run `openapi-to-mcp` as follows:

```json
{
    "mcpServers": {
        "openapi-to-mcp": {
            "command": "go",
            "args": [
                "run",
                "main.go"
            ]
        }
    }
}
```

Alternatively, if you have already built the binary, you can use it directly:

```json
{
    "mcpServers": {
        "openapi-to-mcp": {
            "command": "./openapi-to-mcp" 
        }
    }
}
```

**SSE Mode**:

If `MCP_TRANSPORT` is set to `sse`, the server will start on the `BASE_URL` (e.g., `http://localhost:8080`). You can configure your MCP client to connect to the SSE endpoint as follows:

```json
{
    "mcpServers": {
        "openapi-to-mcp": {
            "url": "http://localhost:8080/sse"
        }
    }
}
```

**Stream Mode**:

If `MCP_TRANSPORT` is set to `stream`, the server will start on the `BASE_URL` (e.g., `http://localhost:8080`). You can configure your MCP client to connect to the Stream endpoint as follows:

```json
{
    "mcpServers": {
        "openapi-to-mcp": {
            "url": "http://localhost:8080/stream"
        }
    }
}
```

Please note that `http://localhost:8080` should be replaced with your actual configured `BASE_URL`.

## Project Structure

```
.github/
core/
  ├── openapi.go          # OpenAPI specification parsing and tool generation logic
  ├── session/            # Session management
  ├── middleware.go       # MCP middleware definitions
  └── utils.go            # Common utility functions
example/
  ├── openapi.yaml        # Example OpenAPI specification file
main.go                   # Main application entry point
README.md
README_zh.md
go.mod
go.sum
```

## Contributing

Contributions are welcome! Feel free to submit a Pull Request or report an Issue.

## License

This project is released under the MIT License.
