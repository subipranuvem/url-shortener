package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	base62Chars       = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	defaultCodeLength = 5
)

type HashService interface {
	GenerateCode(ctx context.Context) (string, error)
}

type RandomHashService struct{}

func NewRandomHashService() *RandomHashService {
	return &RandomHashService{}
}

func (s *RandomHashService) GenerateCode(_ context.Context) (string, error) {
	result := make([]byte, defaultCodeLength)

	for i := range result {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", fmt.Errorf("generate code: %w", err)
		}
		result[i] = base62Chars[idx.Int64()]
	}

	resultString := string(result)
	return resultString, nil
}
