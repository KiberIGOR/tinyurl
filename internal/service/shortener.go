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

type URLRepository interface {
	Get(ctx context.Context, id string) (string, bool)
	Save(ctx context.Context, id, originalURL string) (string, error)
	MassiveSave(ctx context.Context, MassiveURLs []model.MassiveRequest) error
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

func (s *Shortener) Shorten(ctx context.Context,originalURL string) (string, error) {
	const n int = 5
	for i:=0; i<n; i++ {
		id, err := generateID()
		if err != nil {
			return "", err
		}
		shortID, err := s.repo.Save(ctx, id, originalURL)
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

func (s *Shortener) Resolve(ctx context.Context,id string) (string, error) {
	originalURL, ok := s.repo.Get(ctx, id)
	if !ok {
		return "", ErrNotFound
	}
	return originalURL, nil
}

func (s *Shortener) MassiveShorten(ctx context.Context, MassiveOriginalURL []model.MassiveRequest) ([]model.MassiveResponse, error) {
	const n int = 5
	for i := range MassiveOriginalURL {
		for j:=0; j<n; j++ {
			//генерируем ShortURL
			id, err := generateID()
			if err != nil {
				return nil, err
			}
			//проверка, что в store нет подобных id
			_, ok := s.repo.Get(ctx, id)
			if ok {
				if j==(n-1){
						return nil, ErrCollision
				}
				continue
			}
			//проверка, что в уже сгенерированных ShortURL нет id
			for k:=0; k<i; k++ {
				if MassiveOriginalURL[k].ShortURL == id {
					if j==(n-1){
						return nil, ErrCollision
					}
					continue
				}
			}
			//записываем id
			MassiveOriginalURL[i].ShortURL = id
			break
		}
	}
	err := s.repo.MassiveSave(ctx, MassiveOriginalURL)
	if err!=nil {
		return nil, fmt.Errorf("failed to store the URL's: %w", err)
	}

	out := make([]model.MassiveResponse, 0, len(MassiveOriginalURL))
	for _,item := range MassiveOriginalURL {
		result, err := url.JoinPath(s.baseURL,item.ShortURL)
		if err!=nil {
			return nil, fmt.Errorf("failed to Join ShortURL's: %w", err)
		}
		out = append(out, model.MassiveResponse{
        ID:       item.ID,
        ShortURL: result, // полный short URL
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

