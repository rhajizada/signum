package renderer_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rhajizada/badger/pkg/renderer"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/goregular"
)

func newRenderer(tb testing.TB) *renderer.Renderer {
	tb.Helper()
	r, err := renderer.NewRendererWithFontFace(basicfont.Face7x13)
	if err != nil {
		tb.Fatalf("new renderer: %v", err)
	}
	return r
}

func TestRendererRender(t *testing.T) {
	r := newRenderer(t)
	badge := renderer.Badge{
		Subject: "XXX",
		Status:  "YYY",
		Color:   renderer.Color("#c0c0c0"),
	}

	output, err := r.Render(badge)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	result := string(output)
	if !strings.Contains(result, "XXX") || !strings.Contains(result, "YYY") {
		t.Fatalf("output missing badge text: %s", result)
	}
	if !strings.Contains(result, "#c0c0c0") {
		t.Fatalf("output missing color: %s", result)
	}
}

func TestNewRendererEmptyPath(t *testing.T) {
	_, err := renderer.NewRenderer("")
	if err == nil {
		t.Fatalf("expected error for empty font path")
	}
}

func TestNewRendererMissingPath(t *testing.T) {
	_, err := renderer.NewRenderer(filepath.Join(t.TempDir(), "missing.ttf"))
	if err == nil {
		t.Fatalf("expected error for missing font file")
	}
}

func TestNewRendererInvalidFontBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.ttf")
	if err := os.WriteFile(path, []byte("not a font"), 0o600); err != nil {
		t.Fatalf("write temp font: %v", err)
	}
	_, err := renderer.NewRenderer(path)
	if err == nil {
		t.Fatalf("expected error for invalid font bytes")
	}
}

func TestNewRendererValidFontBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "goregular.ttf")
	if err := os.WriteFile(path, goregular.TTF, 0o600); err != nil {
		t.Fatalf("write temp font: %v", err)
	}
	if _, err := renderer.NewRenderer(path); err != nil {
		t.Fatalf("expected valid renderer: %v", err)
	}
}

func TestRendererRenderInvalidColor(t *testing.T) {
	r := newRenderer(t)
	_, err := r.Render(renderer.Badge{
		Subject: "XXX",
		Status:  "YYY",
		Color:   renderer.Color("not-a-color"),
	})
	if err == nil {
		t.Fatalf("expected error for invalid color")
	}
}

func TestRendererRenderInvalidStyle(t *testing.T) {
	r := newRenderer(t)
	_, err := r.Render(renderer.Badge{
		Subject: "XXX",
		Status:  "YYY",
		Color:   renderer.Color("#c0c0c0"),
		Style:   renderer.Style("unknown"),
	})
	if err == nil {
		t.Fatalf("expected error for invalid style")
	}
}

func TestRendererRenderNilRenderer(t *testing.T) {
	var r *renderer.Renderer
	_, err := r.Render(renderer.Badge{Subject: "a", Status: "b", Color: renderer.ColorBrightgreen})
	if err == nil {
		t.Fatalf("expected error for nil renderer")
	}
}

func TestNewRendererWithFontFaceNil(t *testing.T) {
	_, err := renderer.NewRendererWithFontFace(nil)
	if err == nil {
		t.Fatalf("expected error for nil font face")
	}
}

func TestRendererRenderDefaultStyle(t *testing.T) {
	r := newRenderer(t)
	output, err := r.Render(renderer.Badge{
		Subject: "build",
		Status:  "passing",
		Color:   renderer.ColorBrightgreen,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(output), "url(#smooth)") {
		t.Fatalf("expected flat template output")
	}
}

func TestRendererRenderStyleVariants(t *testing.T) {
	r := newRenderer(t)
	tests := []struct {
		name          string
		style         renderer.Style
		containString string
	}{
		{name: "flat", style: renderer.StyleFlat, containString: "url(#smooth)"},
		{name: "flat-square", style: renderer.StyleFlatSquare, containString: "mask=\"url(#square)\""},
		{name: "plastic", style: renderer.StylePlastic, containString: "url(#shine)"},
	}

	for _, tc := range tests {
		output, err := r.Render(renderer.Badge{
			Subject: "build",
			Status:  "passing",
			Color:   renderer.ColorBrightgreen,
			Style:   tc.style,
		})
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		if !strings.Contains(string(output), tc.containString) {
			t.Fatalf("%s: expected output to contain %q", tc.name, tc.containString)
		}
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
