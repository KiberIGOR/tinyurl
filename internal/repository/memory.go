package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrAlreadyExist = errors.New("short URL already exist")

type Memory struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewMemory() *Memory {
	return &Memory{
		urls: make(map[string]string),
	}
}

func (m *Memory) Get(ctx context.Context, id string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.get(id)
}

func (m *Memory) Save(ctx context.Context, id, originalURL string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.get(id); ok {
		return "", fmt.Errorf("%w: %q", ErrAlreadyExist, id)
	}
	m.urls[id] = originalURL
	return "", nil
}

func (m *Memory) BatchSave(ctx context.Context, entries []URLEntry) (BatchSaveResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := BatchSaveResult{
		Existing: make(map[string]string),
	}

	for _, item := range entries {
		if shortURL, ok := m.findByOriginal(item.OriginalURL); ok {
			result.Existing[item.OriginalURL] = shortURL
			continue
		}
		if _, ok := m.get(item.ShortURL); ok {
			result.Retries = append(result.Retries, item)
			continue
		}
		m.urls[item.ShortURL] = item.OriginalURL
	}

	return result, nil
}

func (m *Memory) get(id string) (string, bool) {
	originalURL, ok := m.urls[id]
	return originalURL, ok
}

func (m *Memory) findByOriginal(originalURL string) (string, bool) {
	for shortURL, url := range m.urls {
		if url == originalURL {
			return shortURL, true
		}
	}
	return "", false
}
