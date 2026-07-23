package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/subipraNuvem/url-shortener/internal/service"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func TestGenerateCode(t *testing.T) {
	t.Run("returns 5 characters", func(t *testing.T) {
		h := service.NewRandomHashService()

		code, err := h.GenerateCode(context.Background())

		require.NoError(t, err)
		require.Len(t, code, 5)
	})

	t.Run("only base62 characters", func(t *testing.T) {
		h := service.NewRandomHashService()

		for range 100 {
			code, err := h.GenerateCode(context.Background())
			require.NoError(t, err)
			for _, c := range code {
				require.True(t, strings.ContainsRune(base62Alphabet, c), "unexpected character: %c", c)
			}
		}
	})

	t.Run("generates unique codes across runs", func(t *testing.T) {
		h := service.NewRandomHashService()
		seen := make(map[string]struct{})

		for range 50 {
			code, err := h.GenerateCode(context.Background())
			require.NoError(t, err)
			seen[code] = struct{}{}
		}

		require.Greater(t, len(seen), 1, "expected diverse codes, got repeated collisions across 50 runs")
	})
}
