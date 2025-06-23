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

## Installation

You can install `openapi-to-mcp` in two ways.

### Option 1: Using `go install` (Recommended)

The easiest way to install and run the tool is using `go install`. After installation, the `openapi-to-mcp` command will be available in your shell.

```bash
go install github.com/constellation39/openapi-to-mcp@latest
```

### Option 2: From Source

If you want to modify the code or contribute, clone the repository and build the project.

```bash
git clone https://github.com/constellation39/openapi-to-mcp
cd openapi-to-mcp
go build .
# The executable will be ./openapi-to-mcp
```

## Usage

Follow these steps to configure and run the tool.

### Step 1: Configure Environment Variables

Create a `.env` file in the project's root directory or set the following environment variables in your system. You can start by copying the example file:

```bash
# On Windows
copy .env.example .env

# On macOS/Linux
cp .env.example .env
```

Then, edit the `.env` file with your desired configuration:

```dotenv
# MCP transport: stdio, sse, stream (default: stdio)
MCP_TRANSPORT="stdio"
# MCP BASE URL (used for sse and stream modes)
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

### Step 2: Run the Application

Execute the application from your terminal.

If you installed using `go install`:
```bash
openapi-to-mcp
```

If you built from source:
```bash
./openapi-to-mcp
```

For development, you can also run directly using `go run`:
```bash
go run main.go
```

### Step 3: Connect with an MCP Client

Configure your MCP client to connect to `openapi-to-mcp` based on the `MCP_TRANSPORT` you've chosen.

**- Stdio Mode (Default)**:
If `MCP_TRANSPORT` is `stdio`, configure your client to execute the command directly. The `env` block can be used to pass or override environment variables.

```json
{
    "mcpServers": {
        "openapi-to-mcp": {
            "command": "openapi-to-mcp",
            "env": {
                "MCP_BASE_URL": "http://localhost:8080",
                "OPENAPI_SRC": "./example/openapi.yaml"
            }
        }
    }
}
```
*Note: If running from source, you might set `"command"` to `"./openapi-to-mcp"` or use `go run main.go`.*

**- SSE Mode**:
If `MCP_TRANSPORT` is `sse`, the server will start on the `MCP_BASE_URL`. Configure your client to connect to the `/sse` endpoint.

```json
{
    "mcpServers": {
        "openapi-to-mcp": {
            "url": "http://localhost:8080/sse"
        }
    }
}
```

**- Stream Mode**:
If `MCP_TRANSPORT` is `stream`, the server will start on the `MCP_BASE_URL`. Configure your client to connect to the `/stream` endpoint.

```json
{
    "mcpServers": {
        "openapi-to-mcp": {
            "url": "http://localhost:8080/stream"
        }
    }
}
```
*Please ensure `http://localhost:8080` matches the `MCP_BASE_URL` in your configuration.*

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