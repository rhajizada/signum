package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/font/gofont/goregular"
)

func TestMainVersion(t *testing.T) {
	var buf bytes.Buffer
	if err := run([]string{"-version"}, &buf, os.Getenv); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := buf.String(); got == "" {
		t.Fatalf("expected version output")
	}
}

func TestMainRendersBadge(t *testing.T) {
	dir := t.TempDir()
	fontPath := filepath.Join(dir, "goregular.ttf")
	if err := os.WriteFile(fontPath, goregular.TTF, 0o600); err != nil {
		t.Fatalf("write font: %v", err)
	}
	outPath := filepath.Join(dir, "badge.svg")
	if err := run([]string{
		"-font", fontPath,
		"-subject", "build",
		"-status", "passing",
		"-color", "green",
		"-style", "flat",
		"-out", outPath,
	}, &bytes.Buffer{}, os.Getenv); err != nil {
		t.Fatalf("run: %v", err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}

func TestMainNoArgsShowsUsage(t *testing.T) {
	var buf bytes.Buffer
	if err := run(nil, &buf, os.Getenv); err != nil {
		t.Fatalf("run: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("expected usage output")
	}
}

func TestMainMissingArgsExits(t *testing.T) {
	if err := run([]string{"-subject", "build"}, &bytes.Buffer{}, os.Getenv); err == nil {
		t.Fatalf("expected error for missing args")
	}
}

func TestMainWritesStdoutWhenNoOut(t *testing.T) {
	dir := t.TempDir()
	fontPath := filepath.Join(dir, "goregular.ttf")
	if err := os.WriteFile(fontPath, goregular.TTF, 0o600); err != nil {
		t.Fatalf("write font: %v", err)
	}
	var buf bytes.Buffer
	if err := run([]string{
		"-font", fontPath,
		"-subject", "build",
		"-status", "passing",
		"-color", "green",
		"-style", "flat",
	}, &buf, os.Getenv); err != nil {
		t.Fatalf("run: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("build")) {
		t.Fatalf("expected svg output in stdout")
	}
}
