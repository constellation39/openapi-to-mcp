# openapi-to-mcp

`openapi-to-mcp` is a Go language implementation that converts OpenAPI specifications into MCP (Model Context Protocol) tools.

## Overview

This project aims to provide a flexible framework that allows developers to automatically expose existing OpenAPI interfaces as MCP tools through simple configuration, enabling AI models to directly call these interfaces.

## Features

- **OpenAPI to MCP Tool Conversion**: Automatically parses OpenAPI specifications and creates corresponding MCP tools based on the definitions.
- **Multiple Transport Support**: Supports `stdio` (standard input/output), `sse` (Server-Sent Events), and `stream` (HTTP stream) as transport protocols for MCP communication.
- **Configurable Middleware**: Built-in logging and rate-limiting middleware, with support for custom middleware to meet different business requirements.
- **Session Management**: Provides session management functionality, allowing custom logic to be executed at the beginning and end of an MCP session.
- **Environment Variable Configuration**: Flexible configuration through `.env` files or system environment variables.

## Quick Start

### 1. Configure Environment Variables

Create a `.env` file or set the following environment variables:

```dotenv
# MCP Transport method: stdio, sse, stream (default: stdio)
MCP_TRANSPORT="stdio"
MCP_BASE_URL="127.0.0.1:8081"

# OpenAPI specification file path (can be a local file path or URL)
OPENAPI_SRC="./example/openapi.yaml"
# Base URL for API requests
OPENAPI_BASE_URL=

# Additional HTTP headers (JSON format), e.g., '{"X-API-Key": "your-api-key"}'
EXTRA_HEADERS='{"X-API-Key": "your-api-key"}'

# Whether to use cookies (true/false, default: true)
USE_COOKIE=true

# Whether to output logs to standard output (true/false, default: false)
LOG_OUTPUT=false

# Rate limit: requests allowed per second (default: 1)
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

Depending on your chosen `MCP_TRANSPORT` method, use the appropriate MCP client to connect to `openapi-to-mcp`.

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

Or, if you have built the binary:

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

If `MCP_TRANSPORT` is set to `sse`, the server will start on `BASE_URL` (e.g., `http://localhost:8080`). You can configure your MCP client to connect to the SSE endpoint as follows:

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

If `MCP_TRANSPORT` is set to `stream`, the server will start on `BASE_URL` (e.g., `http://localhost:8080`). You can configure your MCP client to connect to the Stream endpoint as follows:

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
  ├── middleware.go       # MCP middleware definition
  └── utils.go            # Common utility functions
example/
  ├── openapi.yaml        # Example OpenAPI specification file
main.go                   # Main application entry point
README.md
README_zh.md
go.mod
go.sum
```

## Contribution

Contributions are welcome! Feel free to submit Pull Requests or report Issues.

## License

This project is released under the MIT License.
