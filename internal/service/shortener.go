package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
)

var ErrNotFound = errors.New("URL not found")

type URLRepository interface {
	Get(id string) (string, bool)
	Save(id, originalURL string)
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
	for {
		id, err := generateID()
		if err != nil {
			return "", err
		}
		if _, exists := s.repo.Get(id); exists {
			continue
		}
		s.repo.Save(id, originalURL)
		return url.JoinPath(s.baseURL, id)
	}
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
