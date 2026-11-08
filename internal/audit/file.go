package audit

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/KiberIGOR/tinyurl/internal/logger"
	"go.uber.org/zap"
)

// FileObserver дописывает JSON-события в конец файла (по одному на строку).
type FileObserver struct {
	path string
	mu   sync.Mutex
	file *os.File
}

func NewFileObserver(path string) (*FileObserver, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	return &FileObserver{path: path, file: f}, nil
}

func (o *FileObserver) Notify(ctx context.Context, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		logger.Log.Error("audit file: marshal event", zap.Error(err))
		return
	}
	data = append(data, '\n')

	o.mu.Lock()
	defer o.mu.Unlock()

	if o.file == nil {
		var f *os.File
		f, err = os.OpenFile(o.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			logger.Log.Error("audit file: open", zap.Error(err), zap.String("path", o.path))
			return
		}
		o.file = f
	}

	if _, err = o.file.Write(data); err != nil {
		logger.Log.Error("audit file: write", zap.Error(err), zap.String("path", o.path))
	}
}
