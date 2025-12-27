package renderer_test

import (
	"testing"

	"github.com/rhajizada/badger/pkg/renderer"
)

func TestColorString(t *testing.T) {
	cases := map[renderer.Color]string{
		renderer.ColorBrightgreen: "#4c1",
		renderer.ColorGreen:       "#97ca00",
		renderer.ColorYellow:      "#dfb317",
		renderer.ColorYellowgreen: "#a4a61d",
		renderer.ColorOrange:      "#fe7d37",
		renderer.ColorRed:         "#e05d44",
		renderer.ColorBlue:        "#007ec6",
		renderer.ColorGrey:        "#555",
		renderer.ColorGray:        "#555",
		renderer.ColorLightgrey:   "#9f9f9f",
		renderer.ColorLightgray:   "#9f9f9f",
	}
	for input, expected := range cases {
		if got := input.String(); got != expected {
			t.Fatalf("expected %q to map to %q, got %q", input, expected, got)
		}
	}
	custom := renderer.Color("magenta")
	if got := custom.String(); got != "magenta" {
		t.Fatalf("expected custom color to return itself, got %q", got)
	}
}

func TestColorIsValid(t *testing.T) {
	valid := []renderer.Color{
		"",
		renderer.ColorBrightgreen,
		renderer.ColorGreen,
		renderer.ColorYellow,
		renderer.ColorYellowgreen,
		renderer.ColorOrange,
		renderer.ColorRed,
		renderer.ColorBlue,
		renderer.ColorGrey,
		renderer.ColorGray,
		renderer.ColorLightgrey,
		renderer.ColorLightgray,
		renderer.Color("#fff"),
		renderer.Color("#abc"),
		renderer.Color("#abcdef"),
		renderer.Color("#ABCDEF"),
	}
	for _, c := range valid {
		if !c.IsValid() {
			t.Fatalf("expected %q to be valid", c)
		}
	}

	invalid := []renderer.Color{
		renderer.Color("#ff"),
		renderer.Color("#ffff"),
		renderer.Color("#fffff"),
		renderer.Color("#gggggg"),
		renderer.Color("not-a-color"),
	}
	for _, c := range invalid {
		if c.IsValid() {
			t.Fatalf("expected %q to be invalid", c)
		}
	}
}
