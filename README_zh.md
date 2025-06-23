# openapi-to-mcp

[English](README.md)

`openapi-to-mcp` 是一个用于将 OpenAPI 规范转换为 MCP (模型上下文协议) 工具的 Go 实现。

## 概述

本项目旨在提供一个灵活的框架，允许开发者通过简单的配置，将现有的 OpenAPI 接口自动暴露为 MCP 工具，从而使 AI 模型能够直接调用这些接口。

## 功能特性

- **OpenAPI 到 MCP 工具转换**：自动解析 OpenAPI 规范，并根据定义创建相应的 MCP 工具。
- **多种传输支持**：支持 `stdio`（标准输入/输出）、`sse`（服务器发送事件）和 `stream`（HTTP 流）作为 MCP 通信的传输协议。
- **状态跟踪与认证**：支持基于 Cookie 的状态跟踪和 JWT (JSON Web Token) 处理。
- **速率限制**：内置速率限制，以防止对大语言模型 (LLM) 的高频调用。
- **环境变量配置**：通过 `.env` 文件或系统环境变量进行灵活配置。
- **更加严格的MCPTool定义**：使得LLM能够更加好的使用TOOL工具

## 安装

您可以通过两种方式安装 `openapi-to-mcp`。

### 方式一：使用 `go install` (推荐)

安装和运行该工具最简单的方法是使用 `go install`。安装后，`openapi-to-mcp` 命令将在您的 shell 中可用。

```bash
go install github.com/constellation39/openapi-to-mcp@latest
```

### 方式二：从源码构建

如果您想修改代码或参与贡献，可以克隆仓库并构建项目。

```bash
git clone https://github.com/constellation39/openapi-to-mcp
cd openapi-to-mcp
go build .
# 可执行文件将是 ./openapi-to-mcp
```

## 使用方法

请按照以下步骤配置和运行该工具。

### 步骤一：配置环境变量

在项目的根目录中创建一个 `.env` 文件，或在您的系统中设置以下环境变量。您可以从复制示例文件开始：

```bash
# 在 Windows 上
copy .env.example .env

# 在 macOS/Linux 上
cp .env.example .env
```

然后，编辑 `.env` 文件以进行您期望的配置：

```dotenv
# MCP 传输协议: stdio, sse, stream (默认为 stdio)
MCP_TRANSPORT="stdio"
# MCP 基础 URL (用于 sse 和 stream 模式)
MCP_BASE_URL="http://localhost:8080"

# OpenAPI 规范文件路径 (可以是本地文件路径或 URL)
OPENAPI_SRC="./example/openapi.yaml"
# API 请求的基础 URL
OPENAPI_BASE_URL=

# 额外的 HTTP 头 (JSON 格式), 例如：'{"X-API-Key": "your-api-key"}'
EXTRA_HEADERS='{"X-API-Key": "your-api-key"}'

# 是否使用 Cookie (true/false, 默认为 true)
USE_COOKIE=true

# 是否将日志输出到标准输出 (true/false, 默认为 false)
LOG_OUTPUT=false

# 速率限制: 每秒允许的请求数 (默认为 1)
RATE_LIMIT_PER_SECOND=1

# 授权头, 例如："Basic xxxx"
AUTHORIZATION_HEADERS="Basic xxxx"
```

### 步骤二：运行应用程序

从您的终端执行应用程序。

如果您使用 `go install` 安装：
```bash
openapi-to-mcp
```

如果您从源码构建：
```bash
./openapi-to-mcp
```

在开发过程中，您也可以直接使用 `go run` 运行：
```bash
go run main.go
```

### 步骤三：连接 MCP 客户端

根据您选择的 `MCP_TRANSPORT`，配置您的 MCP 客户端以连接到 `openapi-to-mcp`。

**- Stdio 模式 (默认)**:
如果 `MCP_TRANSPORT` 是 `stdio`，请配置您的客户端直接执行命令。`env` 块可用于传递或覆盖环境变量。

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
*注意：如果从源码运行，您可以将 "command" 设置为 "./openapi-to-mcp" 或使用 `go run main.go`。*

**- SSE 模式**:
如果 `MCP_TRANSPORT` 是 `sse`，服务器将在 `MCP_BASE_URL` 上启动。请配置您的客户端连接到 `/sse` 端点。

```json
{
    "mcpServers": {
        "openapi-to-mcp": {
            "url": "http://localhost:8080/sse"
        }
    }
}
```

**- Stream 模式**:
如果 `MCP_TRANSPORT` 是 `stream`，服务器将在 `MCP_BASE_URL` 上启动。请配置您的客户端连接到 `/stream` 端点。

```json
{
    "mcpServers": {
        "openapi-to-mcp": {
            "url": "http://localhost:8080/stream"
        }
    }
}
```
*请确保 http://localhost:8080 与您配置中的 MCP_BASE_URL 一致。*

## 项目结构

```
.github/
core/
  ├── openapi.go          # OpenAPI 规范解析和工具生成逻辑
  ├── session/            # 会话管理
  ├── middleware.go       # MCP 中间件定义
  └── utils.go            # 通用工具函数
example/
  ├── openapi.yaml        # OpenAPI 规范示例文件
main.go                   # 应用程序主入口点
README.md
README_zh.md
go.mod
go.sum
```

## 贡献

欢迎贡献！随时可以提交 Pull Request 或报告问题 (Issue)。

## 许可证

本项目根据 MIT 许可证发布。