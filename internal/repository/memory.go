package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/KiberIGOR/tinyurl/internal/model"
)

var ErrAlreadyExist = errors.New("short URL already exist")
var ErrConflict = errors.New("original URL already exists")
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
	if shortURL, ok := m.findByOriginal(originalURL); ok {
		return shortURL, ErrConflict
	}
	if _, ok := m.get(id); ok {
		return "", fmt.Errorf("%w: %q", ErrAlreadyExist, id)
	}
	m.urls[id] = originalURL
	return "", nil
}

func (m *Memory) MassiveSave(ctx context.Context, MassiveURLs []model.MassiveRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _,item :=range MassiveURLs {
		m.urls[item.ShortURL] = item.OriginalURL
	}
	return nil
}

func (m *Memory) get(id string) (string, bool)  {
	originalURL, ok := m.urls[id]
	return originalURL, ok
}

func (m *Memory) findByOriginal(originalURL string) (string, bool) {
	for id, url := range m.urls {
		if url == originalURL {
			return id, true
		}
	}
	return "", false
}