package session

import (
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"
)

type State struct {
	mu          sync.RWMutex   // 保护下面两个可变字段
	Permissions []string       // 可能在会话生命周期中被修改
	Settings    map[string]any // 同上
	StartTime   time.Time
	Client      *http.Client // 每个会话自己的 HTTP Client（带 CookieJar）
}

func (s *State) GetSetting(k string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Settings[k]
	return v, ok
}
func (s *State) SetSetting(k string, v any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Settings[k] = v
}

type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*State
}

var (
	instance *Manager
	once     sync.Once
)

func Instance() *Manager {
	once.Do(func() {
		instance = &Manager{
			sessions: make(map[string]*State),
		}
	})
	return instance
}

func (sm *Manager) CreateSession(id string, p []string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	jar, _ := cookiejar.New(nil) // 标准库实现，已并发安全
	cl := &http.Client{
		Jar:     jar,
		Timeout: 60 * time.Second, // 按需设置
	}

	sm.sessions[id] = &State{
		Permissions: append([]string(nil), p...),
		Settings:    make(map[string]any),
		StartTime:   time.Now(),
		Client:      cl,
	}
}

func (sm *Manager) GetSession(id string) (*State, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	s, ok := sm.sessions[id]
	return s, ok
}

func (sm *Manager) RemoveSession(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, id)
}

func (sm *Manager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}
