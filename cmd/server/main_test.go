package main

import (
	"bytes"
	"flag"
	"io"
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
	origArgs := os.Args
	origStdout := os.Stdout
	origFlags := flag.CommandLine
	defer func() {
		os.Args = origArgs
		os.Stdout = origStdout
		flag.CommandLine = origFlags
	}()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	flag.CommandLine = flag.NewFlagSet("signum-server", flag.ContinueOnError)
	os.Args = []string{"signum-server", "-version"}
	main()

	_ = w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if got := buf.String(); got == "" {
		t.Fatalf("expected version output")
	}
}
