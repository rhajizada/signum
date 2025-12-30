package main

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFontPath(t *testing.T) {
	if err := validateFontPath(""); err == nil {
		t.Fatalf("expected error for empty path")
	}
	if err := validateFontPath(filepath.Join(t.TempDir(), "missing.ttf")); err == nil {
		t.Fatalf("expected error for missing file")
	}

	dir := t.TempDir()
	if err := validateFontPath(dir); err == nil {
		t.Fatalf("expected error for directory path")
	}

	path := filepath.Join(t.TempDir(), "font.ttf")
	if err := os.WriteFile(path, []byte("font"), 0o600); err != nil {
		t.Fatalf("write font: %v", err)
	}
	if err := validateFontPath(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMainVersion(t *testing.T) {
	var buf bytes.Buffer
	if err := runCLI([]string{"-version"}, &buf, slog.Default()); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := buf.String(); got == "" {
		t.Fatalf("expected version output")
	}
}
