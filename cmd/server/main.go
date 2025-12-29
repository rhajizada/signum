package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/rhajizada/signum/docs"
	"github.com/rhajizada/signum/internal/config"
	"github.com/rhajizada/signum/internal/handler"
	"github.com/rhajizada/signum/internal/middleware"
	"github.com/rhajizada/signum/internal/router"
	"github.com/rhajizada/signum/internal/service"
	"github.com/rhajizada/signum/pkg/renderer"
)

const (
	shutdownTimeout   = 5 * time.Second
	readHeaderTimeout = 10 * time.Second
)

func main() {
	logger := slog.Default()

	cfg, err := config.LoadServer()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if cfg.FontPath == "" {
		logger.Error("font path is required", "env", "SIGNUM_FONT_PATH")
		os.Exit(1)
	}
	if info, err := os.Stat(cfg.FontPath); err != nil || info.IsDir() {
		if err == nil {
			err = os.ErrInvalid
		}
		logger.Error("font path is not a file", "path", cfg.FontPath, "error", err)
		os.Exit(1)
	}

	rdr, err := renderer.NewRenderer(cfg.FontPath)
	if err != nil {
		logger.Error("failed to init renderer", "error", err)
		os.Exit(1)
	}

	svc := service.New(rdr)
	h, err := handler.New(svc, logger)
	if err != nil {
		logger.Error("failed to init handler", "error", err)
		os.Exit(1)
	}

	r := router.New(h)
	handlerWithLogging := middleware.Logging(logger)(r)

	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           handlerWithLogging,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("signum server listening", "addr", cfg.Address)
		if err := srv.ListenAndServe(); err != nil {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown failed", "error", err)
			os.Exit(1)
		}
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}
}
