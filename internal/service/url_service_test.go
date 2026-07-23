package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cachemock "github.com/subipraNuvem/url-shortener/internal/cache/mock"
	"github.com/subipraNuvem/url-shortener/internal/config"
	dbmock "github.com/subipraNuvem/url-shortener/internal/database/mock"
	"github.com/subipraNuvem/url-shortener/internal/domain"
	"github.com/subipraNuvem/url-shortener/internal/service"
	svcmock "github.com/subipraNuvem/url-shortener/internal/service/mock"
)

// ── helper ────────────────────────────────────────────────────────────────────

var testConfig = &config.Config{
	BaseURL:               "https://short.ly",
	CacheDefaultTTLMillis: 3_600_000,
}

func newSvc(t *testing.T, repo *dbmock.MockURLRepository, cache *cachemock.MockCache, hasher *svcmock.MockHashService) *service.URLService {
	t.Helper()
	svc, err := service.NewURLService(service.URLServiceParams{
		Repo:   repo,
		Cache:  cache,
		Hasher: hasher,
		Config: testConfig,
	})
	require.NoError(t, err)
	return svc
}

// ── NewURLService ─────────────────────────────────────────────────────────────

func TestNewURLService(t *testing.T) {
	t.Run("repo nil returns error", func(t *testing.T) {
		_, err := service.NewURLService(service.URLServiceParams{
			Cache:  &cachemock.MockCache{},
			Hasher: &svcmock.MockHashService{},
			Config: testConfig,
		})
		require.Error(t, err)
	})

	t.Run("cache nil returns error", func(t *testing.T) {
		_, err := service.NewURLService(service.URLServiceParams{
			Repo:   &dbmock.MockURLRepository{},
			Hasher: &svcmock.MockHashService{},
			Config: testConfig,
		})
		require.Error(t, err)
	})

	t.Run("hasher nil returns error", func(t *testing.T) {
		_, err := service.NewURLService(service.URLServiceParams{
			Repo:   &dbmock.MockURLRepository{},
			Cache:  &cachemock.MockCache{},
			Config: testConfig,
		})
		require.Error(t, err)
	})

	t.Run("config nil returns error", func(t *testing.T) {
		_, err := service.NewURLService(service.URLServiceParams{
			Repo:   &dbmock.MockURLRepository{},
			Cache:  &cachemock.MockCache{},
			Hasher: &svcmock.MockHashService{},
		})
		require.Error(t, err)
	})
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestCreate(t *testing.T) {
	t.Run("auto code success", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		mockedHash.On("GenerateCode", mock.Anything).Return("abc12", nil).Once()
		mockedRepository.On("GetByCode", mock.Anything, "abc12").Return(nil, nil).Once()
		mockedRepository.On("Create", mock.Anything, mock.AnythingOfType("*domain.URL")).Return(nil).Once()
		mockedCacheClient.On("Set", mock.Anything, "url:abc12", "https://example.com", mock.Anything).Return(nil).Once()

		out, err := svc.Create(context.Background(), service.CreateURLInput{LongURL: "https://example.com"})

		require.NoError(t, err)
		require.Equal(t, "abc12", out.Code)
		require.Equal(t, "https://short.ly/abc12", out.ShortURL)
		mockedRepository.AssertExpectations(t)
		mockedCacheClient.AssertExpectations(t)
		mockedHash.AssertExpectations(t)
	})

	t.Run("with alias success", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		mockedRepository.On("GetByCode", mock.Anything, "mygoogle").Return(nil, nil).Once()
		mockedRepository.On("Create", mock.Anything, mock.AnythingOfType("*domain.URL")).Return(nil).Once()
		mockedCacheClient.On("Set", mock.Anything, "url:mygoogle", "https://google.com", mock.Anything).Return(nil).Once()

		out, err := svc.Create(context.Background(), service.CreateURLInput{
			LongURL: "https://google.com",
			Alias:   "mygoogle",
		})

		require.NoError(t, err)
		require.Equal(t, "mygoogle", out.Code)
		require.Equal(t, "https://short.ly/mygoogle", out.ShortURL)
		mockedRepository.AssertExpectations(t)
		mockedCacheClient.AssertExpectations(t)
	})

	t.Run("alias already taken returns ErrAliasTaken", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		mockedRepository.On("GetByCode", mock.Anything, "taken").Return(&domain.URL{Code: "taken"}, nil).Once()

		_, err := svc.Create(context.Background(), service.CreateURLInput{
			LongURL: "https://google.com",
			Alias:   "taken",
		})

		require.ErrorIs(t, err, service.ErrAliasTaken)
		mockedRepository.AssertExpectations(t)
		_ = mockedCacheClient
		_ = mockedHash
	})

	t.Run("collision retry succeeds on third attempt", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		taken := &domain.URL{Code: "aaaaa"}
		mockedHash.On("GenerateCode", mock.Anything).Return("aaaaa", nil).Once()
		mockedHash.On("GenerateCode", mock.Anything).Return("bbbbb", nil).Once()
		mockedHash.On("GenerateCode", mock.Anything).Return("ccccc", nil).Once()
		mockedRepository.On("GetByCode", mock.Anything, "aaaaa").Return(taken, nil).Once()
		mockedRepository.On("GetByCode", mock.Anything, "bbbbb").Return(taken, nil).Once()
		mockedRepository.On("GetByCode", mock.Anything, "ccccc").Return(nil, nil).Once()
		mockedRepository.On("Create", mock.Anything, mock.AnythingOfType("*domain.URL")).Return(nil).Once()
		mockedCacheClient.On("Set", mock.Anything, "url:ccccc", "https://example.com", mock.Anything).Return(nil).Once()

		out, err := svc.Create(context.Background(), service.CreateURLInput{LongURL: "https://example.com"})

		require.NoError(t, err)
		require.Equal(t, "ccccc", out.Code)
		mockedRepository.AssertExpectations(t)
		mockedCacheClient.AssertExpectations(t)
		mockedHash.AssertExpectations(t)
	})

	t.Run("max collisions exceeded returns error", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		taken := &domain.URL{Code: "xxxxx"}
		mockedHash.On("GenerateCode", mock.Anything).Return("xxxxx", nil).Times(5)
		mockedRepository.On("GetByCode", mock.Anything, "xxxxx").Return(taken, nil).Times(5)

		_, err := svc.Create(context.Background(), service.CreateURLInput{LongURL: "https://example.com"})

		require.Error(t, err)
		mockedRepository.AssertExpectations(t)
		mockedHash.AssertExpectations(t)
		_ = mockedCacheClient
	})

	t.Run("repo create fails returns error", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		mockedHash.On("GenerateCode", mock.Anything).Return("abc12", nil).Once()
		mockedRepository.On("GetByCode", mock.Anything, "abc12").Return(nil, nil).Once()
		mockedRepository.On("Create", mock.Anything, mock.AnythingOfType("*domain.URL")).Return(errors.New("write timeout")).Once()

		_, err := svc.Create(context.Background(), service.CreateURLInput{LongURL: "https://example.com"})

		require.Error(t, err)
		mockedRepository.AssertExpectations(t)
		mockedHash.AssertExpectations(t)
		_ = mockedCacheClient
	})
}

// ── Resolve ───────────────────────────────────────────────────────────────────

func TestResolve(t *testing.T) {
	t.Run("cache hit returns long url", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		mockedCacheClient.On("Get", mock.Anything, "url:abc12").Return("https://example.com", nil).Once()
		// IncrementClicks runs in a goroutine — Maybe() prevents failure if it resolves after test ends
		mockedRepository.On("IncrementClicks", mock.Anything, "abc12").Return(nil).Maybe()

		longURL, err := svc.Resolve(context.Background(), "abc12")

		require.NoError(t, err)
		require.Equal(t, "https://example.com", longURL)
		mockedCacheClient.AssertExpectations(t)
		_ = mockedHash
	})

	t.Run("db hit populates cache and returns long url", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		url := &domain.URL{Code: "abc12", LongURL: "https://example.com", IsActive: true}
		mockedCacheClient.On("Get", mock.Anything, "url:abc12").Return("", nil).Once()
		mockedRepository.On("GetByCode", mock.Anything, "abc12").Return(url, nil).Once()
		mockedCacheClient.On("Set", mock.Anything, "url:abc12", "https://example.com", mock.Anything).Return(nil).Once()
		mockedRepository.On("IncrementClicks", mock.Anything, "abc12").Return(nil).Maybe()

		longURL, err := svc.Resolve(context.Background(), "abc12")

		require.NoError(t, err)
		require.Equal(t, "https://example.com", longURL)
		mockedCacheClient.AssertExpectations(t)
		mockedRepository.AssertExpectations(t)
		_ = mockedHash
	})

	t.Run("url not found returns ErrNotFound", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		mockedCacheClient.On("Get", mock.Anything, "url:missing").Return("", nil).Once()
		mockedRepository.On("GetByCode", mock.Anything, "missing").Return(nil, nil).Once()

		_, err := svc.Resolve(context.Background(), "missing")

		require.ErrorIs(t, err, service.ErrNotFound)
		mockedCacheClient.AssertExpectations(t)
		mockedRepository.AssertExpectations(t)
		_ = mockedHash
	})

	t.Run("url inactive returns ErrInactive", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		url := &domain.URL{Code: "abc12", LongURL: "https://example.com", IsActive: false}
		mockedCacheClient.On("Get", mock.Anything, "url:abc12").Return("", nil).Once()
		mockedRepository.On("GetByCode", mock.Anything, "abc12").Return(url, nil).Once()

		_, err := svc.Resolve(context.Background(), "abc12")

		require.ErrorIs(t, err, service.ErrInactive)
		mockedCacheClient.AssertExpectations(t)
		mockedRepository.AssertExpectations(t)
		_ = mockedHash
	})

	t.Run("repo error is propagated", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		mockedCacheClient.On("Get", mock.Anything, "url:abc12").Return("", nil).Once()
		mockedRepository.On("GetByCode", mock.Anything, "abc12").Return(nil, errors.New("connection refused")).Once()

		_, err := svc.Resolve(context.Background(), "abc12")

		require.Error(t, err)
		mockedCacheClient.AssertExpectations(t)
		mockedRepository.AssertExpectations(t)
		_ = mockedHash
	})
}

// ── Deactivate ────────────────────────────────────────────────────────────────

func TestDeactivate(t *testing.T) {
	t.Run("success invalidates cache", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		url := &domain.URL{Code: "abc12", LongURL: "https://example.com", IsActive: true}
		mockedRepository.On("GetByCode", mock.Anything, "abc12").Return(url, nil).Once()
		mockedRepository.On("Deactivate", mock.Anything, "abc12").Return(nil).Once()
		mockedCacheClient.On("Delete", mock.Anything, "url:abc12").Return(nil).Once()

		err := svc.Deactivate(context.Background(), "abc12")

		require.NoError(t, err)
		mockedRepository.AssertExpectations(t)
		mockedCacheClient.AssertExpectations(t)
		_ = mockedHash
	})

	t.Run("url not found returns ErrNotFound", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		mockedRepository.On("GetByCode", mock.Anything, "missing").Return(nil, nil).Once()

		err := svc.Deactivate(context.Background(), "missing")

		require.ErrorIs(t, err, service.ErrNotFound)
		mockedRepository.AssertExpectations(t)
		_ = mockedCacheClient
		_ = mockedHash
	})

	t.Run("repo deactivate error is propagated", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		url := &domain.URL{Code: "abc12", LongURL: "https://example.com", IsActive: true}
		mockedRepository.On("GetByCode", mock.Anything, "abc12").Return(url, nil).Once()
		mockedRepository.On("Deactivate", mock.Anything, "abc12").Return(errors.New("write timeout")).Once()

		err := svc.Deactivate(context.Background(), "abc12")

		require.Error(t, err)
		mockedRepository.AssertExpectations(t)
		_ = mockedCacheClient
		_ = mockedHash
	})
}

// ── GetStats ──────────────────────────────────────────────────────────────────

func TestGetStats(t *testing.T) {
	t.Run("success returns click count", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		url := &domain.URL{Code: "abc12", LongURL: "https://example.com", IsActive: true}
		mockedRepository.On("GetByCode", mock.Anything, "abc12").Return(url, nil).Once()
		mockedRepository.On("GetClicks", mock.Anything, "abc12").Return(int64(42), nil).Once()

		clicks, err := svc.GetStats(context.Background(), "abc12")

		require.NoError(t, err)
		require.Equal(t, int64(42), clicks)
		mockedRepository.AssertExpectations(t)
		_ = mockedCacheClient
		_ = mockedHash
	})

	t.Run("url not found returns ErrNotFound", func(t *testing.T) {
		mockedRepository, mockedCacheClient, mockedHash := &dbmock.MockURLRepository{}, &cachemock.MockCache{}, &svcmock.MockHashService{}
		svc := newSvc(t, mockedRepository, mockedCacheClient, mockedHash)

		mockedRepository.On("GetByCode", mock.Anything, "missing").Return(nil, nil).Once()

		_, err := svc.GetStats(context.Background(), "missing")

		require.ErrorIs(t, err, service.ErrNotFound)
		mockedRepository.AssertExpectations(t)
		_ = mockedCacheClient
		_ = mockedHash
	})
}
