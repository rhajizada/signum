package main

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCLIVersion(t *testing.T) {
	previous := Version
	Version = "test-version"
	t.Cleanup(func() { Version = previous })

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var out bytes.Buffer
	if err := runCLI([]string{"-version"}, &out, logger); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out.String()) != "test-version" {
		t.Fatalf("expected version output, got %q", out.String())
	}
}

func TestRunCLIUnknownFlag(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var out bytes.Buffer
	if err := runCLI([]string{"-nope"}, &out, logger); err == nil {
		t.Fatalf("expected flag error")
	}
}

func TestValidateFontPath(t *testing.T) {
	if err := validateFontPath(""); err == nil {
		t.Fatalf("expected error for empty path")
	}

	dir := t.TempDir()
	if err := validateFontPath(dir); err == nil || !strings.Contains(err.Error(), "not a file") {
		t.Fatalf("expected file error, got %v", err)
	}

	path := filepath.Join(dir, "font.ttf")
	if err := os.WriteFile(path, []byte("font"), 0o600); err != nil {
		t.Fatalf("write font: %v", err)
	}
	if err := validateFontPath(path); err != nil {
		t.Fatalf("expected valid font path, got %v", err)
	}
}

func TestRunServerInvalidFontPath(t *testing.T) {
	t.Setenv("SIGNUM_FONT_PATH", filepath.Join(t.TempDir(), "missing.ttf"))
	t.Setenv("SIGNUM_SECRET_KEY", "secret")
	t.Setenv("SIGNUM_POSTGRES_HOST", "db")
	t.Setenv("SIGNUM_POSTGRES_USER", "user")
	t.Setenv("SIGNUM_POSTGRES_PASSWORD", "pass")
	t.Setenv("SIGNUM_POSTGRES_DBNAME", "name")

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := runServer(logger); err == nil || !strings.Contains(err.Error(), "font path is invalid") {
		t.Fatalf("expected font path error, got %v", err)
	}
}
