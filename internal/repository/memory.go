package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/KiberIGOR/tinyurl/internal/model"
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
	m.mu.Lock()
	defer m.mu.Unlock()
	url, ok := m.get(id)
	return url, ok
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

func (m *Memory) BatchSave(ctx context.Context, batch []model.BatchRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range batch {
		m.urls[item.ShortURL] = item.OriginalURL
	}
	return nil
}

func (m *Memory) get(id string) (string, bool) {
	originalURL, ok := m.urls[id]
	return originalURL, ok
}
