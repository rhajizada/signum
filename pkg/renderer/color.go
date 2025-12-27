package renderer

// Color represents color of the badge.
type Color string

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
	color, ok := schemeColor(string(c))
	if ok {
		return color
	}
	return string(c)
}

// IsValid reports whether the color is a named scheme color or a hex code.
// Empty string is treated as valid to allow the template default.
func (c Color) IsValid() bool {
	if c == "" {
		return true
	}
	if _, ok := schemeColor(string(c)); ok {
		return true
	}
	return isHexColor(string(c))
}

func schemeColor(name string) (string, bool) {
	switch name {
	case "brightgreen":
		return "#4c1", true
	case "green":
		return "#97ca00", true
	case "yellow":
		return "#dfb317", true
	case "yellowgreen":
		return "#a4a61d", true
	case "orange":
		return "#fe7d37", true
	case "red":
		return "#e05d44", true
	case "blue":
		return "#007ec6", true
	case "grey", "gray":
		return "#555", true
	case "lightgrey", "lightgray":
		return "#9f9f9f", true
	default:
		return "", false
	}
}

func isHexColor(value string) bool {
	if len(value) != 4 && len(value) != 7 {
		return false
	}
	if value[0] != '#' {
		return false
	}
	for i := 1; i < len(value); i++ {
		if !isHexDigit(value[i]) {
			return false
		}
	}
	return true
}

func isHexDigit(b byte) bool {
	switch {
	case b >= '0' && b <= '9':
		return true
	case b >= 'a' && b <= 'f':
		return true
	case b >= 'A' && b <= 'F':
		return true
	default:
		return false
	}
}
