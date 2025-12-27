package renderer

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"sync"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

type bounds struct {
	// SubjectDx is the width of subject string of the badge.
	SubjectDx float64
	SubjectX  float64
	// StatusDx is the width of status string of the badge.
	StatusDx float64
	StatusX  float64
}

func (b bounds) Dx() float64 {
	return b.SubjectDx + b.StatusDx
}

type badgeTemplateData struct {
	Subject string
	Status  string
	Color   string
	Bounds  bounds
}

type Renderer struct {
	fd    *font.Drawer
	tmpls map[Style]*template.Template
	mutex *sync.Mutex
}

// shield.io uses Verdana.ttf to measure text width with an extra 10px.
// This value keeps output widths aligned with shield-style badges.
const extraDx = 13

const (
	dpi          = 72
	fontsize     = 11
	measureShift = 6
)

func NewRenderer(fontPath string) (*Renderer, error) {
	if fontPath == "" {
		return nil, errors.New("font path is required")
	}
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}
	fd, err := newFontDrawerFromBytes(fontBytes, fontsize, dpi)
	if err != nil {
		return nil, err
	}
	tmpls, err := parseTemplates()
	if err != nil {
		return nil, err
	}
	return &Renderer{
		fd:    fd,
		tmpls: tmpls,
		mutex: &sync.Mutex{},
	}, nil
}

func NewRendererWithFontFace(face font.Face) (*Renderer, error) {
	if face == nil {
		return nil, errors.New("font face is required")
	}
	tmpls, err := parseTemplates()
	if err != nil {
		return nil, err
	}
	return &Renderer{
		fd:    &font.Drawer{Face: face},
		tmpls: tmpls,
		mutex: &sync.Mutex{},
	}, nil
}

func (r *Renderer) Render(b Badge) ([]byte, error) {
	if r == nil {
		return nil, errors.New("renderer is nil")
	}
	if !b.Color.IsValid() {
		return nil, fmt.Errorf("invalid color: %q", b.Color)
	}
	style := b.Style
	if style == "" {
		style = StyleFlat
	}
	if !style.IsValid() {
		return nil, fmt.Errorf("invalid style: %q", style)
	}
	tmpl, ok := r.tmpls[style]
	if !ok {
		return nil, fmt.Errorf("missing template for style: %q", style)
	}
	resolvedColor := b.Color.String()
	r.mutex.Lock()
	subjectDx := r.measureString(b.Subject)
	statusDx := r.measureString(b.Status)
	r.mutex.Unlock()

	renderData := badgeTemplateData{
		Subject: b.Subject,
		Status:  b.Status,
		Color:   resolvedColor,
		Bounds: bounds{
			SubjectDx: subjectDx,
			SubjectX:  subjectDx/2.0 + 1,
			StatusDx:  statusDx,
			StatusX:   subjectDx + statusDx/2.0 - 1,
		},
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, renderData); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *Renderer) measureString(s string) float64 {
	return float64(r.fd.MeasureString(s)>>measureShift) + extraDx
}

func newFontDrawerFromBytes(ttfBytes []byte, size, dpi float64) (*font.Drawer, error) {
	ttf, err := truetype.Parse(ttfBytes)
	if err != nil {
		return nil, err
	}
	return &font.Drawer{
		Face: truetype.NewFace(ttf, &truetype.Options{
			Size:    size,
			DPI:     dpi,
			Hinting: font.HintingFull,
		}),
	}, nil
}
