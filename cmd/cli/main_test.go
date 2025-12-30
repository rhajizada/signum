package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/image/font/gofont/goregular"
)

func writeTempFont(tb testing.TB) string {
	tb.Helper()
	dir := tb.TempDir()
	path := filepath.Join(dir, "font.ttf")
	if err := os.WriteFile(path, goregular.TTF, 0o600); err != nil {
		tb.Fatalf("write font: %v", err)
	}
	return path
}

func TestRunUsageOutput(t *testing.T) {
	var out bytes.Buffer
	if err := run(nil, &out, func(string) string { return "" }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Usage of") {
		t.Fatalf("expected usage output, got %q", out.String())
	}
}

func TestRunVersion(t *testing.T) {
	previous := Version
	Version = "test-version"
	t.Cleanup(func() { Version = previous })

	var out bytes.Buffer
	if err := run([]string{"-version"}, &out, func(string) string { return "" }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out.String()) != "test-version" {
		t.Fatalf("expected version output, got %q", out.String())
	}
}

func TestRunMissingFont(t *testing.T) {
	var out bytes.Buffer
	err := run(
		[]string{"-subject", "build", "-status", "passing", "-color", "green"},
		&out,
		func(string) string { return "" },
	)
	if err == nil || !strings.Contains(err.Error(), "font is required") {
		t.Fatalf("expected font error, got %v", err)
	}
}

func TestRunInvalidStyle(t *testing.T) {
	fontPath := writeTempFont(t)
	var out bytes.Buffer
	err := run([]string{
		"-font", fontPath,
		"-subject", "build",
		"-status", "passing",
		"-color", "green",
		"-style", "nope",
	}, &out, func(string) string { return "" })
	if err == nil || !strings.Contains(err.Error(), "invalid style") {
		t.Fatalf("expected style error, got %v", err)
	}
}

func TestRunRendersToStdout(t *testing.T) {
	fontPath := writeTempFont(t)
	var out bytes.Buffer
	if err := run([]string{
		"-font", fontPath,
		"-subject", "build",
		"-status", "passing",
		"-color", "green",
	}, &out, func(string) string { return "" }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "<svg") {
		t.Fatalf("expected svg output, got %q", out.String())
	}
}

func TestRunRendersToFile(t *testing.T) {
	fontPath := writeTempFont(t)
	outputPath := filepath.Join(t.TempDir(), "badge.svg")
	var out bytes.Buffer
	if err := run([]string{
		"-font", fontPath,
		"-subject", "build",
		"-status", "passing",
		"-color", "green",
		"-out", outputPath,
	}, &out, func(string) string { return "" }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	contents, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(contents), "<svg") {
		t.Fatalf("expected svg file, got %q", string(contents))
	}
}

func TestRunUsesEnvFontPath(t *testing.T) {
	fontPath := writeTempFont(t)
	var out bytes.Buffer
	if err := run([]string{
		"-subject", "build",
		"-status", "passing",
		"-color", "green",
	}, &out, func(string) string { return fontPath }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "<svg") {
		t.Fatalf("expected svg output, got %q", out.String())
	}
}

func TestRunMissingSubject(t *testing.T) {
	fontPath := writeTempFont(t)
	var out bytes.Buffer
	err := run([]string{
		"-font", fontPath,
		"-status", "passing",
		"-color", "green",
	}, &out, func(string) string { return "" })
	if err == nil || !strings.Contains(err.Error(), "subject is required") {
		t.Fatalf("expected subject error, got %v", err)
	}
}

func TestRunMissingStatus(t *testing.T) {
	fontPath := writeTempFont(t)
	var out bytes.Buffer
	err := run([]string{
		"-font", fontPath,
		"-subject", "build",
		"-color", "green",
	}, &out, func(string) string { return "" })
	if err == nil || !strings.Contains(err.Error(), "status is required") {
		t.Fatalf("expected status error, got %v", err)
	}
}

func TestRunMissingColor(t *testing.T) {
	fontPath := writeTempFont(t)
	var out bytes.Buffer
	err := run([]string{
		"-font", fontPath,
		"-subject", "build",
		"-status", "passing",
	}, &out, func(string) string { return "" })
	if err == nil || !strings.Contains(err.Error(), "color is required") {
		t.Fatalf("expected color error, got %v", err)
	}
}

func TestRunUnknownFlag(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"-nope"}, &out, func(string) string { return "" }); err == nil {
		t.Fatalf("expected flag error")
	}
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestRunStdoutWriteFailure(t *testing.T) {
	fontPath := writeTempFont(t)
	err := run([]string{
		"-font", fontPath,
		"-subject", "build",
		"-status", "passing",
		"-color", "green",
	}, failingWriter{err: os.ErrInvalid}, func(string) string { return "" })
	if err == nil || !strings.Contains(err.Error(), "write stdout") {
		t.Fatalf("expected stdout error, got %v", err)
	}
}
