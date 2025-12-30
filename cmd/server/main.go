// @title			signum
// @version		dev
// @description	signum API.
// @BasePath		/
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/rhajizada/signum/docs"
	"github.com/rhajizada/signum/internal/config"
	"github.com/rhajizada/signum/internal/handler"
	"github.com/rhajizada/signum/internal/middleware"
	"github.com/rhajizada/signum/internal/repository"
	"github.com/rhajizada/signum/internal/router"
	"github.com/rhajizada/signum/internal/service"
	"github.com/rhajizada/signum/pkg/renderer"
)

const (
	shutdownTimeout   = 5 * time.Second
	readHeaderTimeout = 10 * time.Second
)

// Version is overridden at build time via -ldflags.
//
//nolint:gochecknoglobals // required for build-time version injection
var Version = "dev"

func main() {
	logger := slog.Default()

	if err := runCLI(os.Args[1:], os.Stdout, logger); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}

func runCLI(args []string, stdout io.Writer, logger *slog.Logger) error {
	fs := flag.NewFlagSet("signum-server", flag.ContinueOnError)
	fs.SetOutput(stdout)

	showVersion := fs.Bool("version", false, "Print version and exit")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *showVersion {
		_, err := fmt.Fprintln(stdout, Version)
		return err
	}
	return runServer(logger)
}

func runServer(logger *slog.Logger) error {
	cfg, err := config.LoadServer()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err = validateFontPath(cfg.FontPath); err != nil {
		return err
	}

	db, err := openDB(context.Background(), cfg.Postgres)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error("failed to close database", "error", closeErr)
		}
	}()

	if err = runMigrations(db); err != nil {
		return err
	}

	rdr, err := renderer.NewRenderer(cfg.FontPath)
	if err != nil {
		return fmt.Errorf("init renderer: %w", err)
	}

	tokenManager, err := service.NewTokenManager(cfg.SecretKey)
	if err != nil {
		return fmt.Errorf("init token manager: %w", err)
	}

	svc, err := service.New(rdr, repository.New(db), tokenManager)
	if err != nil {
		return fmt.Errorf("init service: %w", err)
	}

	h, err := handler.New(svc, logger)
	if err != nil {
		return fmt.Errorf("init handler: %w", err)
	}

	docs.SwaggerInfo.Title = "signum"
	docs.SwaggerInfo.Version = Version

	r := router.New(h)
	handlerWithLogging := middleware.Logging(logger)(r)

	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           handlerWithLogging,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	return serve(logger, srv)
}

func validateFontPath(path string) error {
	if path == "" {
		return errors.New("font path is required")
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("font path is invalid: %w", err)
	}
	if info.IsDir() {
		return errors.New("font path is not a file")
	}
	return nil
}

func openDB(ctx context.Context, cfg config.PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect database: %w", err)
	}
	return db, nil
}

func runMigrations(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.Up(db, "data/sql/migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

func serve(logger *slog.Logger, srv *http.Server) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("signum server listening", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
	case serveErr := <-errCh:
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", serveErr)
		}
	}
	return nil
}
