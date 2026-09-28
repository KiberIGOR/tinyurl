package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/KiberIGOR/tinyurl/internal/model"
)

var ErrOpenFile = errors.New("error while saving memory in file")
var ErrMakingJSON = errors.New("error while making JSON")
var ErrParsingJSON = errors.New("error while parsing JSON")
var ErrWritingJSON = errors.New("error while writing JSON")

type FileMemory struct {
	mu         sync.RWMutex
	byShort    map[string]string
	byUser     map[int][]string
	file       *os.File
	enc        *json.Encoder
	nextID     int
	lastUserID int
}

func NewFile(fileName string) (*FileMemory, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("%w: %q", ErrOpenFile, fileName)
	}

	records, legacy, err := loadRecords(file)
	if err != nil {
		file.Close()
		return nil, err
	}

	byShort, byUser, maxUserID := indexesFromRecords(records)

	if legacy {
		if err := writeRecords(file, records); err != nil {
			file.Close()
			return nil, err
		}
	} else if _, err := file.Seek(0, io.SeekEnd); err != nil {
		file.Close()
		return nil, err
	}

	return &FileMemory{
		byShort:    byShort,
		byUser:     byUser,
		file:       file,
		enc:        json.NewEncoder(file),
		nextID:     len(records) + 1,
		lastUserID: maxUserID,
	}, nil
}

func loadRecords(file *os.File) ([]model.MemoryString, bool, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, false, err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, false, err
	}
	if len(data) == 0 {
		return nil, false, nil
	}

	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, false, nil
	}

	if trimmed[0] == '[' {
		var records []model.MemoryString
		if err := json.Unmarshal(trimmed, &records); err != nil {
			return nil, false, fmt.Errorf("%w: %w", ErrParsingJSON, err)
		}
		return records, true, nil
	}

	var records []model.MemoryString
	for line := range bytes.SplitSeq(data, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var record model.MemoryString
		if err := json.Unmarshal(line, &record); err != nil {
			return nil, false, fmt.Errorf("%w: %w", ErrParsingJSON, err)
		}
		records = append(records, record)
	}
	return records, false, nil
}

func writeRecords(file *os.File, records []model.MemoryString) error {
	if err := file.Truncate(0); err != nil {
		return err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	enc := json.NewEncoder(file)
	for _, record := range records {
		if err := enc.Encode(record); err != nil {
			return fmt.Errorf("%w: %w", ErrWritingJSON, err)
		}
	}
	return nil
}

func indexesFromRecords(records []model.MemoryString) (map[string]string, map[int][]string, int) {
	byShort := make(map[string]string, len(records))
	byUser := make(map[int][]string)
	maxUserID := 0

	for _, record := range records {
		byShort[record.ShortURL] = record.OriginalURL
		userID := parseUserID(record.UserID)
		byUser[userID] = append(byUser[userID], record.ShortURL)
		if userID > maxUserID {
			maxUserID = userID
		}
	}

	return byShort, byUser, maxUserID
}

func parseUserID(value string) int {
	if value == "" {
		return 0
	}
	userID, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return userID
}

func (m *FileMemory) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.file == nil {
		return nil
	}
	err := m.file.Close()
	m.file = nil
	m.enc = nil
	return err
}

func (m *FileMemory) Get(ctx context.Context, id string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.get(id)
}

func (m *FileMemory) Save(ctx context.Context, shortURL, originalURL string, userID int) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.get(shortURL); ok {
		return "", fmt.Errorf("%w: %q", ErrAlreadyExist, shortURL)
	}

	record := model.MemoryString{
		ID:          strconv.Itoa(m.nextID),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      strconv.Itoa(userID),
	}
	if err := m.enc.Encode(record); err != nil {
		return "", fmt.Errorf("%w: %w", ErrWritingJSON, err)
	}

	m.nextID++
	m.byShort[shortURL] = originalURL
	m.byUser[userID] = append(m.byUser[userID], shortURL)
	return "", nil
}

func (m *FileMemory) BatchSave(ctx context.Context, entries []URLEntry, userID int) (BatchSaveResult, error) {
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

		record := model.MemoryString{
			ID:          strconv.Itoa(m.nextID),
			ShortURL:    item.ShortURL,
			OriginalURL: item.OriginalURL,
			UserID:      strconv.Itoa(userID),
		}
		if err := m.enc.Encode(record); err != nil {
			return result, fmt.Errorf("%w: %w", ErrWritingJSON, err)
		}

		m.nextID++
		m.byShort[item.ShortURL] = item.OriginalURL
		m.byUser[userID] = append(m.byUser[userID], item.ShortURL)
	}

	return result, nil
}

func (m *FileMemory) GetURLsByUserID(ctx context.Context, userID int) ([]URLEntry,error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shorts := m.byUser[userID]
	if len(shorts)==0 {
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

func (m *FileMemory) GetLastId(ctx context.Context) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastUserID++
	return m.lastUserID
}

func (m *FileMemory) get(id string) (string, bool) {
	originalURL, ok := m.byShort[id]
	return originalURL, ok
}

func (m *FileMemory) findByOriginal(originalURL string) (string, bool) {
	for shortURL, url := range m.byShort {
		if url == originalURL {
			return shortURL, true
		}
	}
	return "", false
}
