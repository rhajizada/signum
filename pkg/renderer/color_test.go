package renderer_test

import (
	"testing"

	"github.com/rhajizada/signum/pkg/renderer"
	"github.com/stretchr/testify/assert"
)

func TestColorString(t *testing.T) {
	tests := []struct {
		name     string
		input    renderer.Color
		expected string
	}{
		{name: "brightgreen", input: renderer.ColorBrightgreen, expected: "#4c1"},
		{name: "green", input: renderer.ColorGreen, expected: "#97ca00"},
		{name: "yellow", input: renderer.ColorYellow, expected: "#dfb317"},
		{name: "yellowgreen", input: renderer.ColorYellowgreen, expected: "#a4a61d"},
		{name: "orange", input: renderer.ColorOrange, expected: "#fe7d37"},
		{name: "red", input: renderer.ColorRed, expected: "#e05d44"},
		{name: "blue", input: renderer.ColorBlue, expected: "#007ec6"},
		{name: "grey", input: renderer.ColorGrey, expected: "#555"},
		{name: "gray", input: renderer.ColorGray, expected: "#555"},
		{name: "lightgrey", input: renderer.ColorLightgrey, expected: "#9f9f9f"},
		{name: "lightgray", input: renderer.ColorLightgray, expected: "#9f9f9f"},
		{name: "custom", input: renderer.Color("magenta"), expected: "magenta"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.input.String())
		})
	}
}

func TestColorIsValid(t *testing.T) {
	tests := []struct {
		name  string
		input renderer.Color
		valid bool
	}{
		{name: "empty", input: "", valid: true},
		{name: "brightgreen", input: renderer.ColorBrightgreen, valid: true},
		{name: "green", input: renderer.ColorGreen, valid: true},
		{name: "yellow", input: renderer.ColorYellow, valid: true},
		{name: "yellowgreen", input: renderer.ColorYellowgreen, valid: true},
		{name: "orange", input: renderer.ColorOrange, valid: true},
		{name: "red", input: renderer.ColorRed, valid: true},
		{name: "blue", input: renderer.ColorBlue, valid: true},
		{name: "grey", input: renderer.ColorGrey, valid: true},
		{name: "gray", input: renderer.ColorGray, valid: true},
		{name: "lightgrey", input: renderer.ColorLightgrey, valid: true},
		{name: "lightgray", input: renderer.ColorLightgray, valid: true},
		{name: "hex short", input: renderer.Color("#fff"), valid: true},
		{name: "hex short mixed", input: renderer.Color("#abc"), valid: true},
		{name: "hex long", input: renderer.Color("#abcdef"), valid: true},
		{name: "hex uppercase", input: renderer.Color("#ABCDEF"), valid: true},
		{name: "too short", input: renderer.Color("#ff"), valid: false},
		{name: "too long short form", input: renderer.Color("#ffff"), valid: false},
		{name: "wrong length", input: renderer.Color("#fffff"), valid: false},
		{name: "non hex", input: renderer.Color("#gggggg"), valid: false},
		{name: "named invalid", input: renderer.Color("not-a-color"), valid: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.valid, tc.input.IsValid())
		})
	}
}
