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
	mu     sync.RWMutex
	urls   map[string]string
	file   *os.File
	enc    *json.Encoder
	nextID int
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

	urls := make(map[string]string, len(records))
	for _, record := range records {
		urls[record.ShortURL] = record.OriginalURL
	}

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
		urls:   urls,
		file:   file,
		enc:    json.NewEncoder(file),
		nextID: len(records) + 1,
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

func (m *FileMemory) Save(ctx context.Context, id, originalURL string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.get(id); ok {
		return "", fmt.Errorf("%w: %q", ErrAlreadyExist, id)
	}

	record := model.MemoryString{
		ID:          strconv.Itoa(m.nextID),
		ShortURL:    id,
		OriginalURL: originalURL,
	}
	if err := m.enc.Encode(record); err != nil {
		return "", fmt.Errorf("%w: %w", ErrWritingJSON, err)
	}

	m.nextID++
	m.urls[id] = originalURL
	return "", nil
}

func (m *FileMemory) BatchSave(ctx context.Context, entries []URLEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, item := range entries {
		record := model.MemoryString{
			ID:          strconv.Itoa(m.nextID),
			ShortURL:    item.ShortURL,
			OriginalURL: item.OriginalURL,
		}
		if err := m.enc.Encode(record); err != nil {
			return fmt.Errorf("%w: %w", ErrWritingJSON, err)
		}
		m.nextID++
		m.urls[item.ShortURL] = item.OriginalURL
	}
	return nil
}

func (m *FileMemory) get(id string) (string, bool) {
	originalURL, ok := m.urls[id]
	return originalURL, ok
}
