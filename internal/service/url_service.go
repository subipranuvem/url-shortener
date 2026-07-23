package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/subipraNuvem/url-shortener/internal/cache"
	"github.com/subipraNuvem/url-shortener/internal/config"
	"github.com/subipraNuvem/url-shortener/internal/database"
	"github.com/subipraNuvem/url-shortener/internal/domain"
)

const cacheKeyPrefix = "url:"

var (
	ErrNotFound   = errors.New("url not found")
	ErrInactive   = errors.New("url inactive")
	ErrAliasTaken = errors.New("alias already taken")
)

type URLService struct {
	repo   database.URLRepository
	cache  cache.Cache
	hasher HashService
	cfg    *config.Config
}

type URLServiceParams struct {
	Repo   database.URLRepository
	Cache  cache.Cache
	Hasher HashService
	Config *config.Config
}

// Params struct chosen over functional options or builder: all fields are required,
// no optional config, and the service scope doesn't justify the extra boilerplate.
func NewURLService(params URLServiceParams) (*URLService, error) {
	if params.Repo == nil {
		return nil, errors.New("URLServiceParams.Repo is required")
	}
	if params.Cache == nil {
		return nil, errors.New("URLServiceParams.Cache is required")
	}
	if params.Hasher == nil {
		return nil, errors.New("URLServiceParams.Hasher is required")
	}
	if params.Config == nil {
		return nil, errors.New("URLServiceParams.Config is required")
	}

	return &URLService{
		repo:   params.Repo,
		cache:  params.Cache,
		hasher: params.Hasher,
		cfg:    params.Config,
	}, nil
}

type CreateURLInput struct {
	LongURL string
	Alias   string
}

type CreateURLOutput struct {
	Code     string
	ShortURL string
}

func (s *URLService) Create(ctx context.Context, input CreateURLInput) (*CreateURLOutput, error) {
	code := input.Alias
	if code == "" {
		var err error
		code, err = s.generateUniqueCode(ctx)
		if err != nil {
			return nil, fmt.Errorf("create url: %w", err)
		}
	} else {
		existingCode, err := s.repo.GetByCode(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("create url: %w", err)
		}
		if existingCode != nil {
			return nil, ErrAliasTaken
		}
	}

	now := time.Now().UTC()
	url := &domain.URL{
		Code:      code,
		LongURL:   input.LongURL,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.repo.Create(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("create url: %w", err)
	}

	cacheKey := cacheKeyPrefix + code
	err = s.cache.Set(ctx, cacheKey, input.LongURL, cache.WithTTL(s.cfg.CacheDefaultTTL()))
	if err != nil {
		slog.WarnContext(ctx, "cache set failed after create", "code", code, "error", err)
	}

	return &CreateURLOutput{Code: code, ShortURL: s.cfg.BaseURL + "/" + code}, nil
}

func (s *URLService) Resolve(ctx context.Context, code string) (string, error) {
	cacheKey := cacheKeyPrefix + code
	cached, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		slog.WarnContext(ctx, "cache get failed", "code", code, "error", err)
	}

	if cached != "" {
		go s.incrementClicks(code)
		return cached, nil
	}

	url, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("resolve: %w", err)
	}
	if url == nil {
		return "", ErrNotFound
	}
	if !url.IsActive {
		return "", ErrInactive
	}

	err = s.cache.Set(ctx, cacheKey, url.LongURL, cache.WithTTL(s.cfg.CacheDefaultTTL()))
	if err != nil {
		slog.WarnContext(ctx, "cache set failed after resolve", "code", code, "error", err)
	}

	go s.incrementClicks(code)

	return url.LongURL, nil
}

func (s *URLService) GetByCode(ctx context.Context, code string) (*domain.URL, error) {
	url, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("get by code: %w", err)
	}
	if url == nil {
		return nil, ErrNotFound
	}
	return url, nil
}

func (s *URLService) GetStats(ctx context.Context, code string) (int64, error) {
	if _, err := s.GetByCode(ctx, code); err != nil {
		return 0, err
	}
	clicks, err := s.repo.GetClicks(ctx, code)
	if err != nil {
		return 0, fmt.Errorf("get stats: %w", err)
	}
	return clicks, nil
}

func (s *URLService) Deactivate(ctx context.Context, code string) error {
	url, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("deactivate: %w", err)
	}
	if url == nil {
		return ErrNotFound
	}

	err = s.repo.Deactivate(ctx, code)
	if err != nil {
		return fmt.Errorf("deactivate: %w", err)
	}

	err = s.cache.Delete(ctx, cacheKeyPrefix+code)
	if err != nil {
		slog.Warn("cache delete failed on deactivate", "code", code, "error", err)
	}

	return nil
}

func (s *URLService) generateUniqueCode(ctx context.Context) (string, error) {
	const maxAttempts = 5
	for range maxAttempts {
		code, err := s.hasher.GenerateCode(ctx)
		if err != nil {
			return "", err
		}
		existing, err := s.repo.GetByCode(ctx, code)
		if err != nil {
			return "", err
		}
		if existing == nil {
			return code, nil
		}
		slog.WarnContext(ctx, "code collision, retrying", "code", code)
	}
	return "", fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
}

func (s *URLService) incrementClicks(code string) {
	err := s.repo.IncrementClicks(context.Background(), code)
	if err != nil {
		slog.Warn("increment clicks failed", "code", code, "error", err)
	}
}
