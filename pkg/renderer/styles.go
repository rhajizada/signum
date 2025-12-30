package renderer

import (
	_ "embed"
	"fmt"
	"html/template"
)

//go:embed templates/flat.svg.tmpl
var flatTemplate string

//go:embed templates/flat-square.svg.tmpl
var flatSquareTemplate string

//go:embed templates/plastic.svg.tmpl
var plasticTemplate string

type Style string

const (
	StyleFlat       Style = "flat"
	StyleFlatSquare Style = "flat-square"
	StylePlastic    Style = "plastic"
)

func (s Style) IsValid() bool {
	switch s {
	case StyleFlat, StyleFlatSquare, StylePlastic:
		return true
	default:
		return false
	}
}

func parseTemplates() (map[Style]*template.Template, error) {
	templates := map[Style]string{
		StyleFlat:       flatTemplate,
		StyleFlatSquare: flatSquareTemplate,
		StylePlastic:    plasticTemplate,
	}
	parsed := make(map[Style]*template.Template, len(templates))
	for style, tmplText := range templates {
		if tmplText == "" {
			return nil, fmt.Errorf("empty template for style: %q", style)
		}
		tmpl, err := template.New(string(style)).Parse(stripXMLWhitespace(tmplText))
		if err != nil {
			return nil, err
		}
		parsed[style] = tmpl
	}
	return parsed, nil
}
