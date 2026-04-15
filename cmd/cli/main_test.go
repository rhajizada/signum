package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/font/gofont/goregular"
)

func writeTempFont(tb testing.TB) string {
	tb.Helper()
	path := filepath.Join(tb.TempDir(), "font.ttf")
	require.NoError(tb, os.WriteFile(path, goregular.TTF, 0o600))
	return path
}

type failingWriter struct{ err error }

func (w failingWriter) Write(_ []byte) (int, error) { return 0, w.err }

func TestRun(t *testing.T) {
	previous := Version
	Version = "test-version"
	t.Cleanup(func() { Version = previous })

	tests := []struct {
		name       string
		args       []string
		writer     any
		envFont    func(string) string
		assertions func(t *testing.T, out string, err error, outputPath string)
	}{
		{
			name: "prints usage",
			assertions: func(t *testing.T, out string, err error, _ string) {
				require.NoError(t, err)
				assert.Contains(t, out, "Usage of")
			},
		},
		{
			name: "prints version",
			args: []string{"-version"},
			assertions: func(t *testing.T, out string, err error, _ string) {
				require.NoError(t, err)
				assert.Equal(t, "test-version\n", out)
			},
		},
		{
			name: "requires font",
			args: []string{"-subject", "build", "-status", "passing", "-color", "green"},
			assertions: func(t *testing.T, _ string, err error, _ string) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "font is required")
			},
		},
		{
			name: "rejects invalid style",
			args: func() []string {
				fontPath := writeTempFont(t)
				return []string{
					"-font",
					fontPath,
					"-subject",
					"build",
					"-status",
					"passing",
					"-color",
					"green",
					"-style",
					"nope",
				}
			}(),
			assertions: func(t *testing.T, _ string, err error, _ string) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid style")
			},
		},
		{
			name: "renders to stdout",
			args: func() []string {
				fontPath := writeTempFont(t)
				return []string{"-font", fontPath, "-subject", "build", "-status", "passing", "-color", "green"}
			}(),
			assertions: func(t *testing.T, out string, err error, _ string) {
				require.NoError(t, err)
				assert.Contains(t, out, "<svg")
			},
		},
		{
			name: "renders to file",
			args: func() []string {
				fontPath := writeTempFont(t)
				outputPath := filepath.Join(t.TempDir(), "badge.svg")
				return []string{
					"-font",
					fontPath,
					"-subject",
					"build",
					"-status",
					"passing",
					"-color",
					"green",
					"-out",
					outputPath,
				}
			}(),
			assertions: func(t *testing.T, _ string, err error, outputPath string) {
				require.NoError(t, err)
				contents, readErr := os.ReadFile(outputPath)
				require.NoError(t, readErr)
				assert.Contains(t, string(contents), "<svg")
			},
		},
		{
			name:    "uses env font path",
			args:    []string{"-subject", "build", "-status", "passing", "-color", "green"},
			envFont: func(string) string { return writeTempFont(t) },
			assertions: func(t *testing.T, out string, err error, _ string) {
				require.NoError(t, err)
				assert.Contains(t, out, "<svg")
			},
		},
		{
			name: "requires subject",
			args: func() []string {
				fontPath := writeTempFont(t)
				return []string{"-font", fontPath, "-status", "passing", "-color", "green"}
			}(),
			assertions: func(t *testing.T, _ string, err error, _ string) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "subject is required")
			},
		},
		{
			name: "requires status",
			args: func() []string {
				fontPath := writeTempFont(t)
				return []string{"-font", fontPath, "-subject", "build", "-color", "green"}
			}(),
			assertions: func(t *testing.T, _ string, err error, _ string) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "status is required")
			},
		},
		{
			name: "requires color",
			args: func() []string {
				fontPath := writeTempFont(t)
				return []string{"-font", fontPath, "-subject", "build", "-status", "passing"}
			}(),
			assertions: func(t *testing.T, _ string, err error, _ string) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "color is required")
			},
		},
		{
			name: "rejects unknown flag",
			args: []string{"-nope"},
			assertions: func(t *testing.T, _ string, err error, _ string) {
				require.Error(t, err)
			},
		},
		{
			name: "reports stdout write failure",
			args: func() []string {
				fontPath := writeTempFont(t)
				return []string{"-font", fontPath, "-subject", "build", "-status", "passing", "-color", "green"}
			}(),
			writer: failingWriter{err: os.ErrInvalid},
			assertions: func(t *testing.T, _ string, err error, _ string) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "write stdout")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			writer := any(&out)
			if tc.writer != nil {
				writer = tc.writer
			}

			outputPath := ""
			for i := range tc.args {
				if tc.args[i] == "-out" && i+1 < len(tc.args) {
					outputPath = tc.args[i+1]
				}
			}

			envFont := func(string) string { return "" }
			if tc.envFont != nil {
				envFont = tc.envFont
			}

			err := run(tc.args, writer.(interface{ Write([]byte) (int, error) }), envFont)
			tc.assertions(t, out.String(), err, outputPath)
		})
	}
}
