package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/constellation39/openapi-to-mcp/core"
	"github.com/constellation39/openapi-to-mcp/core/session"
	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	transport := strings.ToLower(core.LoadEnv("MCP_TRANSPORT", "stdio"))
	if err := start(transport); err != nil {
		log.Fatal(err)
	}
}

func start(transport string) error {
	var logger *log.Logger
	logOutput := core.LoadEnv("LOG_OUTPUT", "false")
	if logOutput == "false" {
		logger = log.New(io.Discard, "[MCP] ", log.LstdFlags|log.Lshortfile)
	} else {
		logger = log.New(os.Stdout, "[MCP] ", log.LstdFlags|log.Lshortfile)
	}

	sessionMgr := session.Instance()

	hooks := &server.Hooks{}
	hooks.AddOnRegisterSession(func(ctx context.Context, s server.ClientSession) {
		sessionMgr.CreateSession(s.SessionID(), []string{"read"})
		logger.Printf(">>> session start %s", s.SessionID())
	})
	hooks.AddOnUnregisterSession(func(ctx context.Context, s server.ClientSession) {
		sessionMgr.RemoveSession(s.SessionID())
		logger.Printf("<<< session end   %s", s.SessionID())
	})
	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		logger.Printf("afterCallTool: %v, %v, %v\n", id, message, result)
	})
	hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		logger.Printf("beforeCallTool: %v, %v\n", id, message)
	})

	serverOptions := []server.ServerOption{
		server.WithHooks(hooks),
		server.WithToolHandlerMiddleware(core.NewLoggingMiddleware(logger).ToolMiddleware),
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, true),
		server.WithPromptCapabilities(true),
		server.WithRecovery(),
		server.WithLogging(),
	}

	rpsStr := core.LoadEnv("RATE_LIMIT_PER_SECOND", "")
	if rpsStr != "" {
		rateLimit, err := strconv.ParseFloat(rpsStr, 64)
		if err != nil {
			return err
		}
		server.WithToolHandlerMiddleware(core.NewRateLimitMiddleware(rateLimit, 1).ToolMiddleware)
	}

	mcpServer := server.NewMCPServer(
		core.ServerName,
		core.ServerVersion,
		serverOptions...,
	)

	extraHeadersStr := core.LoadEnv("EXTRA_HEADERS", "")
	hdr := map[string]string{}
	if extraHeadersStr != "" {
		if err := json.Unmarshal([]byte(extraHeadersStr), &hdr); err != nil {
			logger.Printf("Failed to parse EXTRA_HEADERS: %v", err)
		}
	}

	src := core.LoadEnv("OPENAPI_SRC", "")
	if src != "" {
		doc, err := core.LoadOpenAPIDoc(src)
		if err != nil {
			logger.Fatalf("openapi load error: %v", err)
		}
		err = core.AddToolFromOpenAPI(
			mcpServer,
			core.LoadEnv("OPENAPI_BASE_URL", ""),
			hdr,
			doc,
		)
		if err != nil {
			logger.Fatal(err)
		}
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		logger.Println("<< shutting down ...")
		os.Exit(0)
	}()

	switch transport {
	case "stdio":
		return server.ServeStdio(mcpServer)
	case "sse":
		baseURL := core.LoadEnv("MCP_BASE_URL", ":8080")
		httpServer := server.NewSSEServer(mcpServer)
		return httpServer.Start(baseURL)
	case "stream":
		baseURL := core.LoadEnv("MCP_BASE_URL", ":8080")
		httpServer := server.NewStreamableHTTPServer(mcpServer, server.WithStateLess(true))
		return httpServer.Start(baseURL)
	default:
		return fmt.Errorf("unknown MCP_TRANSPORT=%s", transport)
	}
}
