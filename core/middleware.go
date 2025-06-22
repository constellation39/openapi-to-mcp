package core

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/time/rate"
	"log"
	"sync"
)

type LoggingMiddleware struct{ logger *log.Logger }

func NewLoggingMiddleware(l *log.Logger) *LoggingMiddleware { return &LoggingMiddleware{l} }

func (m *LoggingMiddleware) ToolMiddleware(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		//start := time.Now()
		//sid := server.GetSessionID(ctx)
		//m.logger.Printf("[TOOL] -> session=%s tool=%s", sid, r.Params.Name)
		//res, err := next(ctx, r)
		//m.logger.Printf("[TOOL] <- session=%s tool=%s dur=%.0fms err=%v",
		//	sid, r.Params.Name, time.Since(start).Seconds()*1000, err)
		//return res, err
		//return nil, nil
		return next(ctx, request)
	}
}

func (m *LoggingMiddleware) ResourceMiddleware(next server.ResourceHandlerFunc) server.ResourceHandlerFunc {
	return func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		//start := time.Now()
		//sid := server.GetSessionID(ctx)
		//m.logger.Printf("[RES] -> session=%s uri=%s", sid, r.Params.URI)
		//res, err := next(ctx, r)
		//m.logger.Printf("[RES] <- session=%s uri=%s dur=%.0fms err=%v",
		//	sid, r.Params.URI, time.Since(start).Seconds()*1000, err)
		//return res, err
		//return nil, nil
		return next(ctx, request)
	}

}

type RateLimitMiddleware struct {
	rate, burst int
	limiters    map[string]*rate.Limiter
	mutex       sync.RWMutex
}

func NewRateLimitMiddleware(rps float64, burst int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		rate:     int(rps),
		burst:    burst,
		limiters: map[string]*rate.Limiter{},
	}
}

func (m *RateLimitMiddleware) limiter(id string) *rate.Limiter {
	m.mutex.RLock()
	l, ok := m.limiters[id]
	m.mutex.RUnlock()
	if ok {
		return l
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	l = rate.NewLimiter(rate.Limit(m.rate), m.burst)
	m.limiters[id] = l
	return l
}

func (m *RateLimitMiddleware) ToolMiddleware(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if cs := server.ClientSessionFromContext(ctx); cs != nil {
			sid := cs.SessionID()
			if !m.limiter(sid).Allow() {
				return nil, fmt.Errorf("rate-limit exceeded for session %s", sid)
			}
		}
		return next(ctx, r)
	}
}
