package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"portfolio/internal/adapters/cache"
	"portfolio/internal/adapters/config"
	"portfolio/internal/adapters/handler/rest"
	"portfolio/internal/adapters/ipfetcher"
	"portfolio/internal/adapters/logger"
	"portfolio/internal/adapters/ratelimiter"
	"portfolio/internal/adapters/repo"
	"portfolio/internal/core/service"
	"portfolio/pkg/tuning"
	"syscall"
	"time"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer done()

	logger, err := logger.NewZapLogger()
	if err != nil {
		logger.Error("failed create logger", "error", err)
		return
	}

	tuning.Tune(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		return
	}

	repo, err := repo.NewPostgresRepo(cfg.DB_HOST, cfg.DB_PORT, cfg.DB_USER, cfg.DB_PASS, cfg.DB_NAME)
	if err != nil {
		logger.Error("failed to create repo", "error", err)
		return
	}

	logger.Info("successfully created repo")

	ipfetcher := ipfetcher.NewIPWhoisAdapter()
	metaIpCache := cache.NewMetaIPCache(ctx, 32)
	svc := service.NewService(repo, logger, ipfetcher, metaIpCache, cfg.QUERY_PASS)
	rl := ratelimiter.NewRateLimiter(ctx, 10)
	handler := rest.NewRestHandler(svc, logger, rl)

	server := &http.Server{
		Addr:         fmt.Sprint(":", cfg.HTTP_PORT),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      handler.WireRoutes(),
	}

	go func() {
		logger.Info("starting http server", "port", cfg.HTTP_PORT)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed to start", "error", err)
			done()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down http server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server forced to shutdown", "error", err)
	}

	logger.Info("http server exited cleanly")
}
