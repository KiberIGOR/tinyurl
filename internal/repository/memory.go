package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrAlreadyExist = errors.New("short URL already exist")
var ErrNotFound = errors.New("URL not found")
var ErrDeleted = errors.New("URL deleted")

type Memory struct {
	mu         sync.RWMutex
	byShort    map[string]string // short_url → original_url
	byUser     map[int][]string  // userID → список short_url
	deleted    map[string]struct{}
	lastUserID int
}

func NewMemory() *Memory {
	return &Memory{
		byShort: make(map[string]string),
		byUser:  make(map[int][]string),
		deleted: make(map[string]struct{}),
	}
}

func (m *Memory) Get(ctx context.Context, id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.isDeleted(id) {
		return "", ErrDeleted
	}
	originalURL, ok := m.byShort[id]
	if !ok {
		return "", ErrNotFound
	}
	return originalURL, nil
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
		if m.isDeleted(short) {
			continue
		}
		originalURL, ok := m.byShort[short]
		if !ok {
			continue
		}
		result = append(result, URLEntry{
			ShortURL:    short,
			OriginalURL: originalURL,
		})
	}
	if len(result) == 0 {
		return nil, ErrNoContent
	}
	return result, nil
}

func (m *Memory) GetLastID(ctx context.Context) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastUserID++
	return m.lastUserID
}

func (m *Memory) DeleteByShortURLs(ctx context.Context, userID int, shortURLs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, short := range shortURLs {
		if !m.userOwnsShort(userID, short) {
			continue
		}
		if _, ok := m.byShort[short]; !ok {
			continue
		}
		m.deleted[short] = struct{}{}
	}
	return nil
}

func (m *Memory) get(id string) (string, bool) {
	if m.isDeleted(id) {
		return "", false
	}
	originalURL, ok := m.byShort[id]
	return originalURL, ok
}

func (m *Memory) isDeleted(short string) bool {
	_, ok := m.deleted[short]
	return ok
}

func (m *Memory) userOwnsShort(userID int, short string) bool {
	for _, s := range m.byUser[userID] {
		if s == short {
			return true
		}
	}
	return false
}

func (m *Memory) findByOriginal(originalURL string) (string, bool) {
	for shortURL, url := range m.byShort {
		if url == originalURL {
			return shortURL, true
		}
	}
	return "", false
}
