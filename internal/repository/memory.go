package repository

import (
	"errors"
	"fmt"
	"sync"
)

var ErrAlreadyExist = errors.New("URL already exist")
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
	defer m.mu.Unlock()
	url, ok := m.get(id)
	return url, ok
}

func (m *Memory) Save(id, originalURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.get(id); ok {
		return fmt.Errorf("%w: %q", ErrAlreadyExist, id)
	}
	m.urls[id] = originalURL
	return nil
}

func (m *Memory) get(id string) (string, bool)  {
	originalURL, ok := m.urls[id]
	return originalURL, ok
}