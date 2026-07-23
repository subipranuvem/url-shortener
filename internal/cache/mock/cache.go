package mock

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/subipraNuvem/url-shortener/internal/cache"
)

type MockCache struct{ mock.Mock }

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key, value string, opts ...cache.SetOption) error {
	return m.Called(ctx, key, value, opts).Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	return m.Called(ctx, key).Error(0)
}

func (m *MockCache) IncrementCounter(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCache) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
