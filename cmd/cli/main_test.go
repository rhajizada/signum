package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"os/exec"

	"golang.org/x/image/font/gofont/goregular"
)

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

	flag.CommandLine = flag.NewFlagSet("signum", flag.ContinueOnError)
	os.Args = []string{"signum", "-version"}
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

func TestMainRendersBadge(t *testing.T) {
	origArgs := os.Args
	origFlags := flag.CommandLine
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origFlags
	}()

	dir := t.TempDir()
	fontPath := filepath.Join(dir, "goregular.ttf")
	if err := os.WriteFile(fontPath, goregular.TTF, 0o600); err != nil {
		t.Fatalf("write font: %v", err)
	}
	outPath := filepath.Join(dir, "badge.svg")

	flag.CommandLine = flag.NewFlagSet("signum", flag.ContinueOnError)
	os.Args = []string{
		"signum",
		"-font", fontPath,
		"-subject", "build",
		"-status", "passing",
		"-color", "green",
		"-style", "flat",
		"-out", outPath,
	}
	main()

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}

func TestMainNoArgsShowsUsage(t *testing.T) {
	origArgs := os.Args
	origFlags := flag.CommandLine
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origFlags
	}()

	flag.CommandLine = flag.NewFlagSet("signum", flag.ContinueOnError)
	os.Args = []string{"signum"}
	main()
}

func TestMainMissingArgsExits(t *testing.T) {
	cmd := execCommand("-subject", "build")
	if err := cmd.Run(); err == nil {
		t.Fatalf("expected non-zero exit")
	}
}

func TestMainWritesStdoutWhenNoOut(t *testing.T) {
	origArgs := os.Args
	origStdout := os.Stdout
	origFlags := flag.CommandLine
	defer func() {
		os.Args = origArgs
		os.Stdout = origStdout
		flag.CommandLine = origFlags
	}()

	dir := t.TempDir()
	fontPath := filepath.Join(dir, "goregular.ttf")
	if err := os.WriteFile(fontPath, goregular.TTF, 0o600); err != nil {
		t.Fatalf("write font: %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	flag.CommandLine = flag.NewFlagSet("signum", flag.ContinueOnError)
	os.Args = []string{
		"signum",
		"-font", fontPath,
		"-subject", "build",
		"-status", "passing",
		"-color", "green",
		"-style", "flat",
	}
	main()

	_ = w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("build")) {
		t.Fatalf("expected svg output in stdout")
	}
}

func execCommand(args ...string) *exec.Cmd {
	cmdArgs := append([]string{"-test.run=TestHelperProcess", "--"}, args...)
	cmd := exec.Command(os.Args[0], cmdArgs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"signum"}, os.Args[i+1:]...)
			break
		}
	}
	flag.CommandLine = flag.NewFlagSet("signum", flag.ContinueOnError)
	main()
}
