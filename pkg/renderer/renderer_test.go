package renderer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rhajizada/signum/pkg/renderer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/goregular"
)

func newRenderer(tb testing.TB) *renderer.Renderer {
	tb.Helper()
	r, err := renderer.NewRendererWithFontFace(basicfont.Face7x13)
	require.NoError(tb, err)
	return r
}

func TestRendererRender(t *testing.T) {
	r := newRenderer(t)
	tests := []struct {
		name        string
		badge       renderer.Badge
		wantErr     bool
		contains    []string
		notContains []string
	}{
		{
			name:     "renders valid badge",
			badge:    renderer.Badge{Subject: "XXX", Status: "YYY", Color: renderer.Color("#c0c0c0")},
			contains: []string{"XXX", "YYY", "#c0c0c0"},
		},
		{
			name:    "rejects invalid color",
			badge:   renderer.Badge{Subject: "XXX", Status: "YYY", Color: renderer.Color("not-a-color")},
			wantErr: true,
		},
		{
			name: "rejects invalid style",
			badge: renderer.Badge{
				Subject: "XXX",
				Status:  "YYY",
				Color:   renderer.Color("#c0c0c0"),
				Style:   renderer.Style("unknown"),
			},
			wantErr: true,
		},
		{
			name:     "uses default flat style",
			badge:    renderer.Badge{Subject: "build", Status: "passing", Color: renderer.ColorBrightgreen},
			contains: []string{"url(#smooth-"},
		},
		{
			name: "renders flat style",
			badge: renderer.Badge{
				Subject: "build",
				Status:  "passing",
				Color:   renderer.ColorBrightgreen,
				Style:   renderer.StyleFlat,
			},
			contains: []string{"url(#smooth-"},
		},
		{
			name: "renders flat square style",
			badge: renderer.Badge{
				Subject: "build",
				Status:  "passing",
				Color:   renderer.ColorBrightgreen,
				Style:   renderer.StyleFlatSquare,
			},
			contains: []string{"url(#square-"},
		},
		{
			name: "renders plastic style",
			badge: renderer.Badge{
				Subject: "build",
				Status:  "passing",
				Color:   renderer.ColorBrightgreen,
				Style:   renderer.StylePlastic,
			},
			contains: []string{"url(#shine-"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := r.Render(tc.badge)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			for _, expected := range tc.contains {
				assert.Contains(t, string(output), expected)
			}
			for _, unexpected := range tc.notContains {
				assert.NotContains(t, string(output), unexpected)
			}
		})
	}
}

func TestNewRenderer(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T) string
		wantErr bool
	}{
		{name: "rejects empty path", prepare: func(*testing.T) string { return "" }, wantErr: true},
		{
			name:    "rejects missing path",
			prepare: func(t *testing.T) string { return filepath.Join(t.TempDir(), "missing.ttf") },
			wantErr: true,
		},
		{
			name: "rejects invalid font bytes",
			prepare: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "bad.ttf")
				require.NoError(t, os.WriteFile(path, []byte("not a font"), 0o600))
				return path
			},
			wantErr: true,
		},
		{
			name: "accepts valid font bytes",
			prepare: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "goregular.ttf")
				require.NoError(t, os.WriteFile(path, goregular.TTF, 0o600))
				return path
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := renderer.NewRenderer(tc.prepare(t))
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestRendererNilHandling(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "nil renderer rejects render",
			run: func(t *testing.T) {
				var r *renderer.Renderer
				_, err := r.Render(renderer.Badge{Subject: "a", Status: "b", Color: renderer.ColorBrightgreen})
				require.Error(t, err)
			},
		},
		{
			name: "nil font face rejected",
			run: func(t *testing.T) {
				_, err := renderer.NewRendererWithFontFace(nil)
				require.Error(t, err)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}

func BenchmarkRender(b *testing.B) {
	r := newRenderer(b)
	badge := renderer.Badge{Subject: "XXX", Status: "YYY", Color: renderer.ColorBlue}

	for b.Loop() {
		_, err := r.Render(badge)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderParallel(b *testing.B) {
	r := newRenderer(b)
	badge := renderer.Badge{Subject: "XXX", Status: "YYY", Color: renderer.ColorBlue}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := r.Render(badge)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
