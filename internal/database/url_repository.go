package database

import (
	"context"

	"github.com/subipraNuvem/url-shortener/internal/domain"
)

type URLRepository interface {
	Create(ctx context.Context, url *domain.URL) error
	GetByCode(ctx context.Context, code string) (*domain.URL, error)
	Deactivate(ctx context.Context, code string) error
	IncrementClicks(ctx context.Context, code string) error
	GetClicks(ctx context.Context, code string) (int64, error)
}
