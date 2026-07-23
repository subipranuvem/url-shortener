package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/subipraNuvem/url-shortener/internal/api"
	"github.com/subipraNuvem/url-shortener/internal/api/handler"
	redisclient "github.com/subipraNuvem/url-shortener/internal/cache/redis"
	"github.com/subipraNuvem/url-shortener/internal/config"
	"github.com/subipraNuvem/url-shortener/internal/database/scylla"
	"github.com/subipraNuvem/url-shortener/internal/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dbClient := scylla.NewClient(cfg)
	err = dbClient.Connect(ctx)
	if err != nil {
		if ctx.Err() != nil {
			slog.Warn("scylla connect cancelled", "reason", err)
			os.Exit(0)
		}
		slog.Error("scylla connect failed", "error", err)
		os.Exit(1)
	}
	defer dbClient.Close()

	cacheClient := redisclient.NewClient(cfg)
	err = cacheClient.Connect(ctx, cfg.RedisAddr)
	if err != nil {
		slog.Error("redis connect failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		err := cacheClient.Close()
		if err != nil {
			slog.Warn("redis close error", "error", err)
		}
	}()

	dbClient.PingDatabasePeriodically(ctx)

	urlRepo := scylla.NewURLRepository(dbClient)
	hasher := service.NewRandomHashService()
	urlSvc, err := service.NewURLService(service.URLServiceParams{
		Repo:   urlRepo,
		Cache:  cacheClient,
		Hasher: hasher,
		Config: cfg,
	})
	if err != nil {
		slog.Error("url service init failed", "error", err)
		os.Exit(1)
	}

	urlHandler := handler.NewURLHandler(urlSvc)
	redirectHandler := handler.NewRedirectHandler(urlSvc)
	healthHandler := handler.NewHealthHandler(dbClient, cacheClient)

	router := api.NewRouter(urlHandler, redirectHandler, healthHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		slog.Info("starting server", "port", cfg.Port)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	slog.Info("server stopped")
}
