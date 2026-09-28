package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrAlreadyExist = errors.New("short URL already exist")

type Memory struct {
	mu         sync.RWMutex
	byShort    map[string]string // short_url → original_url
	byUser     map[int][]string  // userID → список short_url
	lastUserID int
}

func NewMemory() *Memory {
	return &Memory{
		byShort: make(map[string]string),
		byUser:  make(map[int][]string),
	}
}

func (m *Memory) Get(ctx context.Context, id string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.get(id)
}

func (m *Memory) Save(ctx context.Context, shortURL, originalURL string, userID int) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.get(shortURL); ok {
		return "", fmt.Errorf("%w: %q", ErrAlreadyExist, shortURL)
	}

	m.byShort[shortURL] = originalURL
	m.byUser[userID] = append(m.byUser[userID], shortURL)
	return "", nil
}

func (m *Memory) BatchSave(ctx context.Context, entries []URLEntry, userID int) (BatchSaveResult, error) {
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
		m.byShort[item.ShortURL] = item.OriginalURL
		m.byUser[userID] = append(m.byUser[userID], item.ShortURL)
	}

	return result, nil
}

func (m *Memory) GetURLsByUserID(ctx context.Context, userID int) ([]URLEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shorts := m.byUser[userID]
	if len(shorts) == 0 {
		return nil, ErrNoContent
	}
	result := make([]URLEntry, 0, len(shorts))
	for _, short := range shorts {
		originalURL, ok := m.byShort[short]
		if !ok {
			continue
		}
		result = append(result, URLEntry{
			ShortURL:    short,
			OriginalURL: originalURL,
		})
	}
	return result, nil
}

func (m *Memory) GetLastID(ctx context.Context) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastUserID++
	return m.lastUserID
}

func (m *Memory) get(id string) (string, bool) {
	originalURL, ok := m.byShort[id]
	return originalURL, ok
}

func (m *Memory) findByOriginal(originalURL string) (string, bool) {
	for shortURL, url := range m.byShort {
		if url == originalURL {
			return shortURL, true
		}
	}
	return "", false
}
