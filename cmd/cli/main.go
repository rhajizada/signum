package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/rhajizada/badger/pkg/renderer"
)

func main() {
	fontPath := flag.String("font", "", "Path to a .ttf font file (or set BADGER_CLI_FONT)")
	subject := flag.String("subject", "", "Badge subject text")
	status := flag.String("status", "", "Badge status text")
	color := flag.String("color", "", "Badge color (named or hex)")
	style := flag.String("style", "", "Badge style (flat, flat-square, plastic)")
	output := flag.String("out", "", "Output SVG file path")
	flag.Parse()

	if *fontPath == "" {
		*fontPath = os.Getenv("BADGER_CLI_FONT")
	}
	if *fontPath == "" {
		fatal("font is required (set -font or BADGER_CLI_FONT)")
	}
	if *subject == "" {
		fatal("subject is required")
	}
	if *status == "" {
		fatal("status is required")
	}
	if *color == "" {
		fatal("color is required")
	}
	if *style == "" {
		fatal("style is required")
	}
	if *output == "" {
		fatal("out is required")
	}

	badgeStyle := renderer.Style(*style)
	if !badgeStyle.IsValid() {
		fatal(fmt.Sprintf("invalid style: %q", *style))
	}

	r, err := renderer.NewRenderer(*fontPath)
	if err != nil {
		fatal(err.Error())
	}

	outputBytes, err := r.Render(renderer.Badge{
		Subject: *subject,
		Status:  *status,
		Color:   renderer.Color(*color),
		Style:   badgeStyle,
	})
	if err != nil {
		fatal(err.Error())
	}

	if err := os.WriteFile(*output, outputBytes, 0o600); err != nil {
		fatal(err.Error())
	}
}

func fatal(message string) {
	slog.Error(message)
	os.Exit(1)
}
