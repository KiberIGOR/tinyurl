package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/KiberIGOR/tinyurl/internal/model"
)

var ErrAlreadyExist = errors.New("URL already exist")
var ErrOpenFile = errors.New("Error while seving memory in file")
var ErrMakingJson = errors.New("Error while making JSON")
var ErrParsingJson = errors.New("Error while parsing JSON")
var ErrWritingJson = errors.New("Error while writing JSON")
type Memory struct {
	mu   sync.RWMutex
	urls map[string]string
	fileName string
}

func NewMemory(fileName string) (*Memory,error) {
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
	return &Memory{
		urls: urls,
		fileName: fileName,
	}, nil
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
			return fmt.Errorf("%w: %w", ErrParsingJson, err)
	}
	
	memory = append(memory,model.MemoryString{
		ID: strconv.Itoa(len(m.urls)+1),
		ShortURL: id,
		OriginalURL: originalURL,
	})
	writebyte, err := json.Marshal(memory)
	if err != nil {
        return fmt.Errorf("%w: %w", ErrMakingJson, err)
  }

	err = os.WriteFile(m.fileName, writebyte, 0666)
	if err != nil {
        return fmt.Errorf("%w: %w", ErrWritingJson, err)
  }
	m.urls[id] = originalURL
	return nil
}

func (m *Memory) get(id string) (string, bool)  {
	originalURL, ok := m.urls[id]
	return originalURL, ok
}