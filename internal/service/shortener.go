package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"

	"github.com/KiberIGOR/tinyurl/internal/repository"
)

var ErrNotFound = errors.New("URL not found")
var ErrCollision = errors.New("not found correct id for URL(collision)")

type URLRepository interface {
	Get(id string) (string, bool)
	Save(id, originalURL string) error
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

func (s *Shortener) Shorten(originalURL string) (string, error) {
	const n int = 5
	for i:=0; i<n; i++ {
		id, err := generateID()
		if err != nil {
			return "", err
		}
		err = s.repo.Save(id, originalURL)
		if err !=nil {
			if errors.Is(err,repository.ErrAlreadyExist) {
				continue
			}
			return "", fmt.Errorf("failed to store the URL: %w", err)
		}
		return url.JoinPath(s.baseURL, id)
	}
	return "", ErrCollision
}

func (s *Shortener) Resolve(id string) (string, error) {
	originalURL, ok := s.repo.Get(id)
	if !ok {
		return "", ErrNotFound
	}
	return originalURL, nil
}

func generateID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:8], nil
}
