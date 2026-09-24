package repository

import (
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
	mu   sync.RWMutex
	urls map[string]string
	fileName string
}

func NewFile(fileName string) (*FileMemory,error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err !=nil {
		return nil, err
	}
	defer file.Close()
	memorybyte, err := io.ReadAll(file)
	if err !=nil {
		return nil, err
	}
	var memory []model.MemoryString 
	if len(memorybyte) == 0 {
    memorybyte = []byte("[]")
	}
	err = json.Unmarshal(memorybyte, &memory)
	if err !=nil {
		return nil, err
	}
	urls := make(map[string]string)
	for _, i:=range memory {
		urls[i.ShortURL] = i.OriginalURL
	}
	return &FileMemory{
		urls: urls,
		fileName: fileName,
	}, nil
}

func (m *FileMemory) Get(ctx context.Context, id string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	url, ok := m.get(id)
	return url, ok
}

func (m *FileMemory) Save(ctx context.Context, id, originalURL string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.get(id); ok {
		return "", fmt.Errorf("%w: %q", ErrAlreadyExist, id)
	}

	file, err := os.OpenFile(m.fileName, os.O_RDWR, 0666)
	if err != nil {
		return "", fmt.Errorf("%w: %q", ErrOpenFile, m.fileName)
  }
	defer file.Close()

	memorybyte, err := io.ReadAll(file)
	if err != nil {
			return "", err
		}
	var memory []model.MemoryString 
	if len(memorybyte) == 0 {
    memorybyte = []byte("[]")
	}
	err = json.Unmarshal(memorybyte, &memory)
	if err != nil {
			return "", fmt.Errorf("%w: %w", ErrParsingJSON, err)
	}
	
	memory = append(memory,model.MemoryString{
		ID: strconv.Itoa(len(m.urls)+1),
		ShortURL: id,
		OriginalURL: originalURL,
	})
	writebyte, err := json.Marshal(memory)
	if err != nil {
        return "", fmt.Errorf("%w: %w", ErrMakingJSON, err)
  }

	err = os.WriteFile(m.fileName, writebyte, 0666)
	if err != nil {
        return "", fmt.Errorf("%w: %w", ErrWritingJSON, err)
  }
	m.urls[id] = originalURL
	return "", nil
}

func (m *FileMemory) BatchSave(ctx context.Context, batch []model.BatchRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	file, err := os.OpenFile(m.fileName, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrOpenFile, m.fileName)
  }
	defer file.Close()

	memorybyte, err := io.ReadAll(file)
	if err != nil {
			return err
		}
	var memory []model.MemoryString 
	if len(memorybyte) == 0 {
    memorybyte = []byte("[]")
	}
	err = json.Unmarshal(memorybyte, &memory)
	if err != nil {
			return fmt.Errorf("%w: %w", ErrParsingJSON, err)
	}

	for _, item := range batch {
		memory = append(memory,model.MemoryString{
			ID: strconv.Itoa(len(m.urls)+1),
			ShortURL: item.ShortURL,
			OriginalURL: item.OriginalURL,
		})
		m.urls[item.ShortURL] = item.OriginalURL
	}
	writebyte, err := json.Marshal(memory)
	if err != nil {
        return fmt.Errorf("%w: %w", ErrMakingJSON, err)
  }

	err = os.WriteFile(m.fileName, writebyte, 0666)
	if err != nil {
        return fmt.Errorf("%w: %w", ErrWritingJSON, err)
  }

	return nil
}

func (m *FileMemory) get(id string) (string, bool)  {
	originalURL, ok := m.urls[id]
	return originalURL, ok
}