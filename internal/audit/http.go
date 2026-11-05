package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/KiberIGOR/tinyurl/internal/logger"
	"go.uber.org/zap"
)

// HTTPObserver отправляет событие POST-запросом на удалённый приёмник.
type HTTPObserver struct {
	url    string
	client *http.Client
}

func NewHTTPObserver(url string, client *http.Client) *HTTPObserver {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &HTTPObserver{url: url, client: client}
}

func (o *HTTPObserver) Notify(ctx context.Context, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		logger.Log.Error("audit http: marshal event", zap.Error(err))
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.url, bytes.NewReader(data))
	if err != nil {
		logger.Log.Error("audit http: new request", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		logger.Log.Error("audit http: do request", zap.Error(err), zap.String("url", o.url))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		logger.Log.Error("audit http: bad status", zap.Int("status", resp.StatusCode), zap.String("url", o.url))
	}
}
