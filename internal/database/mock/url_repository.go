package mock

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/subipraNuvem/url-shortener/internal/domain"
)

type MockURLRepository struct{ mock.Mock }

func (m *MockURLRepository) Create(ctx context.Context, url *domain.URL) error {
	return m.Called(ctx, url).Error(0)
}

func (m *MockURLRepository) GetByCode(ctx context.Context, code string) (*domain.URL, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.URL), args.Error(1)
}

func (m *MockURLRepository) Deactivate(ctx context.Context, code string) error {
	return m.Called(ctx, code).Error(0)
}

func (m *MockURLRepository) IncrementClicks(ctx context.Context, code string) error {
	return m.Called(ctx, code).Error(0)
}

func (m *MockURLRepository) GetClicks(ctx context.Context, code string) (int64, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(int64), args.Error(1)
}
