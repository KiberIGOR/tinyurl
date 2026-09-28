package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"

	"github.com/KiberIGOR/tinyurl/internal/model"
	"github.com/KiberIGOR/tinyurl/internal/repository"
)

var ErrNotFound = errors.New("URL not found")
var ErrCollision = errors.New("not found correct id for URL(collision)")
var ErrConflict = errors.New("original URL already exists")
var ErrNoContent = errors.New("URLs not found")

type URLRepository interface {
	Get(ctx context.Context, id string) (string, bool)
	Save(ctx context.Context, id, originalURL string, userID int) (string, error)
	BatchSave(ctx context.Context, entries []repository.URLEntry, userID int) (repository.BatchSaveResult, error)
	GetURLsByUserID(ctx context.Context, userID int) ([]repository.URLEntry, error)
}

type Shortener struct {
	repo    URLRepository
	baseURL string
}

func NewShortener(repo URLRepository, baseURL string) *Shortener {
	return &Shortener{
		repo:    repo,
		baseURL: baseURL,
	}
}

func (s *Shortener) Shorten(ctx context.Context, originalURL string, userID int) (string, error) {
	const n int = 5
	for i := 0; i < n; i++ {
		id, err := generateID()
		if err != nil {
			return "", err
		}
		shortID, err := s.repo.Save(ctx, id, originalURL, userID)
		if err != nil {
			if errors.Is(err, repository.ErrAlreadyExist) {
				continue
			}
			if errors.Is(err, repository.ErrConflict) {
				shortURL, err2 := url.JoinPath(s.baseURL, shortID)
				if err2 != nil {
					return "", err2
				}
				return shortURL, ErrConflict
			}
			return "", fmt.Errorf("failed to store the URL: %w", err)
		}
		return url.JoinPath(s.baseURL, id)
	}
	return "", ErrCollision
}

func (s *Shortener) Resolve(ctx context.Context, id string) (string, error) {
	originalURL, ok := s.repo.Get(ctx, id)
	if !ok {
		return "", ErrNotFound
	}
	return originalURL, nil
}

func (s *Shortener) BatchShorten(ctx context.Context, batch []model.BatchRequest, userID int) ([]model.BatchResponse, error) {
	const maxAttempts = 5

	byOriginal := make(map[string]string, len(batch))
	for i := range batch {
		if _, ok := byOriginal[batch[i].OriginalURL]; ok {
			continue
		}
		id, err := generateID()
		if err != nil {
			return nil, err
		}
		byOriginal[batch[i].OriginalURL] = id
	}

	pending := make([]repository.URLEntry, 0, len(byOriginal))
	for originalURL, shortID := range byOriginal {
		pending = append(pending, repository.URLEntry{
			ShortURL:    shortID,
			OriginalURL: originalURL,
		})
	}

	for attempt := 0; len(pending) > 0 && attempt < maxAttempts; attempt++ {
		result, err := s.repo.BatchSave(ctx, pending, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to store the URL's: %w", err)
		}

		for originalURL, shortID := range result.Existing {
			byOriginal[originalURL] = shortID
		}

		if len(result.Retries) == 0 {
			pending = nil
			break
		}

		pending = pending[:0]
		for _, item := range result.Retries {
			id, err := generateID()
			if err != nil {
				return nil, err
			}
			byOriginal[item.OriginalURL] = id
			pending = append(pending, repository.URLEntry{
				ShortURL:    id,
				OriginalURL: item.OriginalURL,
			})
		}
	}

	if len(pending) > 0 {
		return nil, ErrCollision
	}

	out := make([]model.BatchResponse, 0, len(batch))
	for _, item := range batch {
		shortID := byOriginal[item.OriginalURL]
		result, err := url.JoinPath(s.baseURL, shortID)
		if err != nil {
			return nil, fmt.Errorf("failed to Join ShortURL's: %w", err)
		}
		out = append(out, model.BatchResponse{
			ID:       item.ID,
			ShortURL: result,
		})
	}
	return out, nil
}

func generateID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:8], nil
}

func (s *Shortener) ResolveByUserID(ctx context.Context, userID int) ([]model.UserResponse, error) {
	urls, err := s.repo.GetURLsByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNoContent) {
			return nil, ErrNoContent
		}
		return nil, ErrNotFound
	}
	out := make([]model.UserResponse, 0, len(urls))
	for _, item := range urls {
		result, err := url.JoinPath(s.baseURL, item.ShortURL)
		if err != nil {
			return nil, fmt.Errorf("failed to Join ShortURL's: %w", err)
		}
		out = append(out, model.UserResponse{
			OriginalURL: item.OriginalURL,
			ShortURL:    result,
		})
	}
	return out, nil
}
