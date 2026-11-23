// Package facade — общая бизнес-логика для HTTP- и gRPC-обработчиков.
package facade

import (
	"context"
	"time"

	"github.com/KiberIGOR/tinyurl/internal/audit"
	"github.com/KiberIGOR/tinyurl/internal/model"
	"github.com/KiberIGOR/tinyurl/internal/service"
)

// Shortener делегирует операции сервису сокращения и аудиту.
type Shortener struct {
	svc   *service.Shortener
	audit *audit.Broker
}

func NewShortener(svc *service.Shortener, auditBroker *audit.Broker) *Shortener {
	return &Shortener{svc: svc, audit: auditBroker}
}

func (f *Shortener) ShortenURL(ctx context.Context, userID int, originalURL string) (string, error) {
	shortURL, err := f.svc.Shorten(ctx, originalURL, userID)
	if err != nil {
		return "", err
	}
	f.publishAudit(ctx, audit.NewShortenEvent(time.Now().Unix(), userID, originalURL))
	return shortURL, nil
}

func (f *Shortener) ExpandURL(ctx context.Context, userID int, id string) (string, error) {
	originalURL, err := f.svc.Resolve(ctx, id)
	if err != nil {
		return "", err
	}
	f.publishAudit(ctx, audit.NewFollowEvent(time.Now().Unix(), userID, originalURL))
	return originalURL, nil
}

func (f *Shortener) ListUserURLs(ctx context.Context, userID int) ([]model.UserResponse, error) {
	return f.svc.ResolveByUserID(ctx, userID)
}

func (f *Shortener) publishAudit(ctx context.Context, event audit.Event) {
	if f.audit == nil {
		return
	}
	f.audit.Publish(ctx, event)
}
