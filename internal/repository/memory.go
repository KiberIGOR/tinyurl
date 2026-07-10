package repository

import "sync"

type Memory struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewMemory() *Memory {
	return &Memory{
		urls: make(map[string]string),
	}
}

func (m *Memory) Get(id string) (string, bool) {
	m.mu.Lock()
	url, ok := m.urls[id]
	m.mu.Unlock()
	return url, ok
}

func (m *Memory) Save(id, originalURL string) {
	m.mu.Lock()
	m.urls[id] = originalURL
	m.mu.Unlock()
}
