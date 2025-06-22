# openapi-to-mcp

`openapi-to-mcp` 是一个将 OpenAPI 规范转换为 MCP (Model Context Protocol) 工具的 Go 语言实现。

## 概述

该项目旨在提供一个灵活的框架，允许开发者通过简单的配置将现有的 OpenAPI 接口自动暴露为 MCP 工具，从而使得 AI 模型能够直接调用这些接口。

## 特性

- **OpenAPI 转换为 MCP 工具**: 自动解析 OpenAPI 规范，并根据规范定义创建相应的 MCP 工具。
- **多种传输方式支持**: 支持 `stdio` (标准输入输出)、`sse` (Server-Sent Events) 和 `stream` (HTTP 流) 作为 MCP 通信的传输协议。
- **可配置的中间件**: 内置日志记录和速率限制中间件，并支持自定义中间件，以满足不同的业务需求。
- **会话管理**: 提供会话管理功能，允许在 MCP 会话开始和结束时执行自定义逻辑。
- **环境变量配置**: 通过 `.env` 文件或系统环境变量进行灵活配置。

## 快速开始

### 1. 配置环境变量

创建 `.env` 文件或设置以下环境变量：

```dotenv
# MCP 传输方式：stdio, sse, stream (默认: stdio)
MCP_TRANSPORT="stdio"
# MCP BASE URL
MCP_BASE_URL="http://localhost:8080"

# OpenAPI 规范文件路径（可以是本地文件路径或 URL）
OPENAPI_SRC="./example/openapi.yaml"
# # Base URL for API requests
OPENAPI_BASE_URL=

# 额外的 HTTP 头（JSON 格式），例如 '{"X-API-Key": "your-api-key"}'
EXTRA_HEADERS='{"X-API-Key": "your-api-key"}'

# 是否使用 Cookie (true/false, 默认: true)
USE_COOKIE=true

# 是否输出日志到标准输出 (true/false, 默认: false)
LOG_OUTPUT=false

# 速率限制：每秒允许的请求数 (默认: 1)
RATE_LIMIT_PER_SECOND=1

# 授权头，例如 "Basic xxxx"
AUTHORIZATION_HEADERS="Basic xxxx"
```

### 2. 运行项目

```bash
git clone https://github.com/constellation39/openapi-to-mcp
cd openapi-to-mcp

go run main.go
#or
go build .
./openapi-to-mcp
```

### 3. 使用 MCP 客户端连接

根据您选择的 `MCP_TRANSPORT` 方式，使用相应的 MCP 客户端连接到 `openapi-to-mcp`。

**Stdio 模式 (默认)**:

如果 `MCP_TRANSPORT` 设置为 `stdio`，您可以通过以下方式配置 MCP 客户端来运行 `openapi-to-mcp`：

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

或者，如果您已经构建了二进制文件，可以直接使用：

```json
{
    "mcpServers": {
        "openapi-to-mcp": {
            "command": "./openapi-to-mcp" 
        }
    }
}
```

**SSE 模式**:

如果 `MCP_TRANSPORT` 设置为 `sse`，服务器将在 `BASE_URL` 上启动（例如 `http://localhost:8080`）。您可以通过以下方式配置 MCP 客户端连接到 SSE 端点：

```json
{
    "mcpServers": {
        "openapi-to-mcp": {
            "url": "http://localhost:8080/sse"
        }
    }
}
```

**Stream 模式**:

如果 `MCP_TRANSPORT` 设置为 `stream`，服务器将在 `BASE_URL` 上启动（例如 `http://localhost:8080`）。您可以通过以下方式配置 MCP 客户端连接到 Stream 端点：

```json
{
    "mcpServers": {
        "openapi-to-mcp": {
            "url": "http://localhost:8080/stream"
        }
    }
}
```

请注意，`http://localhost:8080` 应该替换为您实际配置的 `BASE_URL`。
## 项目结构

```
.github/
core/
  ├── openapi.go          # OpenAPI 规范解析和工具生成逻辑
  ├── session/            # 会话管理
  ├── middleware.go       # MCP 中间件定义
  └── utils.go            # 常用工具函数
example/
  ├── openapi.yaml        # 示例 OpenAPI 规范文件
main.go                   # 主应用程序入口
README.md
README_zh.md
go.mod
go.sum
```

## 贡献

欢迎贡献！请随意提交 Pull Request 或报告 Issue。

## 许可证

本项目根据 MIT 许可证发布。