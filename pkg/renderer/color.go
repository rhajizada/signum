package renderer

import "regexp"

// Color represents color of the badge.
type Color string

// ColorScheme contains named colors that could be used to render the badge.
var ColorScheme = map[string]string{
	"brightgreen": "#4c1",
	"green":       "#97ca00",
	"yellow":      "#dfb317",
	"yellowgreen": "#a4a61d",
	"orange":      "#fe7d37",
	"red":         "#e05d44",
	"blue":        "#007ec6",
	"grey":        "#555",
	"gray":        "#555",
	"lightgrey":   "#9f9f9f",
	"lightgray":   "#9f9f9f",
}

// Standard colors.
const (
	ColorBrightgreen = Color("brightgreen")
	ColorGreen       = Color("green")
	ColorYellow      = Color("yellow")
	ColorYellowgreen = Color("yellowgreen")
	ColorOrange      = Color("orange")
	ColorRed         = Color("red")
	ColorBlue        = Color("blue")
	ColorGrey        = Color("grey")
	ColorGray        = Color("gray")
	ColorLightgrey   = Color("lightgrey")
	ColorLightgray   = Color("lightgray")
)

func (c Color) String() string {
	color, ok := ColorScheme[string(c)]
	if ok {
		return color
	} else {
		return string(c)
	}
}

var hexColorRe = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// IsValid reports whether the color is a named scheme color or a hex code.
// Empty string is treated as valid to allow the template default.
func (c Color) IsValid() bool {
	if c == "" {
		return true
	}
	if _, ok := ColorScheme[string(c)]; ok {
		return true
	}
	return hexColorRe.MatchString(string(c))
}
