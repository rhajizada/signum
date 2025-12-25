package renderer

import (
	"fmt"
	"html/template"
	"strings"
	"sync"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

func testRenderer(tmplText string) (*Renderer, error) {
	tmpl, err := template.New("mock-template").Parse(tmplText)
	if err != nil {
		return nil, err
	}
	return &Renderer{
		fd:    &font.Drawer{Face: basicfont.Face7x13},
		tmpl:  tmpl,
		mutex: &sync.Mutex{},
	}, nil
}

func TestRendererRender(t *testing.T) {
	mockTemplate := strings.TrimSpace(`
	{{.Subject}},{{.Status}},{{.Color}},{{with .Bounds}}{{.SubjectX}},{{.SubjectDx}},{{.StatusX}},{{.StatusDx}},{{.Dx}}{{end}}
	`)

	r, err := testRenderer(mockTemplate)
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}
	badge := Badge{
		Subject: "XXX",
		Status:  "YYY",
		Color:   Color("#c0c0c0"),
	}

	subjectDx := r.measureString(badge.Subject)
	statusDx := r.measureString(badge.Status)
	bounds := bounds{
		SubjectDx: subjectDx,
		SubjectX:  subjectDx/2.0 + 1,
		StatusDx:  statusDx,
		StatusX:   subjectDx + statusDx/2.0 - 1,
	}
	expected := fmt.Sprintf(
		"%s,%s,%s,%v,%v,%v,%v,%v",
		badge.Subject,
		badge.Status,
		badge.Color,
		bounds.SubjectX,
		bounds.SubjectDx,
		bounds.StatusX,
		bounds.StatusDx,
		bounds.Dx(),
	)

	output, err := r.Render(badge)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if string(output) != expected {
		t.Fatalf("expect %q got %q", expected, string(output))
	}
}

func TestNewRendererEmptyPath(t *testing.T) {
	_, err := NewRenderer("")
	if err == nil {
		t.Fatalf("expected error for empty font path")
	}
}

func TestRendererRenderInvalidColor(t *testing.T) {
	mockTemplate := strings.TrimSpace(`
	{{.Subject}},{{.Status}},{{.Color}},{{with .Bounds}}{{.SubjectX}},{{.SubjectDx}},{{.StatusX}},{{.StatusDx}},{{.Dx}}{{end}}
	`)

	r, err := testRenderer(mockTemplate)
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	_, err = r.Render(Badge{
		Subject: "XXX",
		Status:  "YYY",
		Color:   Color("not-a-color"),
	})
	if err == nil {
		t.Fatalf("expected error for invalid color")
	}
}

func BenchmarkRender(b *testing.B) {
	mockTemplate := strings.TrimSpace(`
	{{.Subject}},{{.Status}},{{.Color}},{{with .Bounds}}{{.SubjectX}},{{.SubjectDx}},{{.StatusX}},{{.StatusDx}},{{.Dx}}{{end}}
	`)
	r, err := testRenderer(mockTemplate)
	if err != nil {
		b.Fatalf("parse template: %v", err)
	}
	badge := Badge{Subject: "XXX", Status: "YYY", Color: ColorBlue}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := r.Render(badge)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderParallel(b *testing.B) {
	mockTemplate := strings.TrimSpace(`
	{{.Subject}},{{.Status}},{{.Color}},{{with .Bounds}}{{.SubjectX}},{{.SubjectDx}},{{.StatusX}},{{.StatusDx}},{{.Dx}}{{end}}
	`)
	r, err := testRenderer(mockTemplate)
	if err != nil {
		b.Fatalf("parse template: %v", err)
	}
	badge := Badge{Subject: "XXX", Status: "YYY", Color: ColorBlue}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := r.Render(badge)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
