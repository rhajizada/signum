package main

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCLI(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		assertions func(t *testing.T, out string, err error)
	}{
		{
			name: "prints version",
			args: []string{"-version"},
			assertions: func(t *testing.T, out string, err error) {
				require.NoError(t, err)
				assert.Equal(t, "test-version\n", out)
			},
		},
		{
			name: "rejects unknown flag",
			args: []string{"-nope"},
			assertions: func(t *testing.T, _ string, err error) {
				require.Error(t, err)
			},
		},
	}

	previous := Version
	Version = "test-version"
	t.Cleanup(func() { Version = previous })
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			err := runCLI(tc.args, &out, logger)
			tc.assertions(t, out.String(), err)
		})
	}
}

func TestValidateFontPath(t *testing.T) {
	tests := []struct {
		name       string
		prepare    func(t *testing.T) string
		assertions func(t *testing.T, err error)
	}{
		{
			name:    "rejects empty path",
			prepare: func(*testing.T) string { return "" },
			assertions: func(t *testing.T, err error) {
				require.EqualError(t, err, "font path is required")
			},
		},
		{
			name:    "rejects directory",
			prepare: func(t *testing.T) string { return t.TempDir() },
			assertions: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not a file")
			},
		},
		{
			name: "accepts file",
			prepare: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "font.ttf")
				require.NoError(t, os.WriteFile(path, []byte("font"), 0o600))
				return path
			},
			assertions: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.assertions(t, validateFontPath(tc.prepare(t)))
		})
	}
}

func TestRunServerInvalidFontPath(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "fails on missing font", path: filepath.Join(t.TempDir(), "missing.ttf")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SIGNUM_FONT_PATH", tc.path)
			t.Setenv("SIGNUM_SECRET_KEY", "secret")
			t.Setenv("SIGNUM_POSTGRES_HOST", "db")
			t.Setenv("SIGNUM_POSTGRES_USER", "user")
			t.Setenv("SIGNUM_POSTGRES_PASSWORD", "pass")
			t.Setenv("SIGNUM_POSTGRES_DBNAME", "name")

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			err := runServer(logger)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "font path is invalid")
		})
	}
}
