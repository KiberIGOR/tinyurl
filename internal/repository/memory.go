package repository

import (
	"errors"
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
	url, ok := m.urls[id]
	m.mu.Unlock()
	return url, ok
}

func (m *Memory) Save(id, originalURL string) error {
	if _,ok := m.Get(id); ok {
		return ErrAlreadyExist
	}
	m.mu.Lock()
	m.urls[id] = originalURL
	m.mu.Unlock()
	return nil
}
