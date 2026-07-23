package mock

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockHashService struct{ mock.Mock }

func (m *MockHashService) GenerateCode(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}
