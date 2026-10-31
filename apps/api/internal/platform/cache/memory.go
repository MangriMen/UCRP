package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     any
	expiresAt time.Time
}

type Memory struct {
	mu      sync.RWMutex
	entries map[string]entry
}

func NewMemory() *Memory {
	return &Memory{entries: make(map[string]entry)}
}

func (m *Memory) Get(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.entries[key]
	if !ok {
		return nil, false
	}
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

func (m *Memory) Set(key string, value any, ttlSeconds int) {
	var expires time.Time
	if ttlSeconds > 0 {
		expires = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	}
	m.mu.Lock()
	m.entries[key] = entry{value: value, expiresAt: expires}
	m.mu.Unlock()
}

func (m *Memory) Delete(key string) {
	m.mu.Lock()
	delete(m.entries, key)
	m.mu.Unlock()
}
