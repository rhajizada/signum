package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/rhajizada/signum/pkg/renderer"
)

// Version is overridden at build time via -ldflags.
//
//nolint:gochecknoglobals // required for build-time version injection
var Version = "dev"

func main() {
	logger := slog.Default()

	showVersion := flag.Bool("version", false, "Print version and exit")
	fontPath := flag.String("font", "", "Path to a .ttf font file (or set SIGNUM_FONT_PATH)")
	subject := flag.String("subject", "", "Badge subject text")
	status := flag.String("status", "", "Badge status text")
	color := flag.String("color", "", "Badge color (named or hex)")
	style := flag.String("style", "flat", "Badge style (flat, flat-square, plastic)")
	output := flag.String("out", "", "Output SVG file path")

	if len(os.Args) == 1 {
		flag.Usage()
		return
	}
	flag.Parse()

	if *showVersion {
		_, _ = os.Stdout.WriteString(Version + "\n")
		return
	}

	if *fontPath == "" {
		*fontPath = os.Getenv("SIGNUM_FONT_PATH")
	}
	if *fontPath == "" {
		fatal(logger, "font is required (set -font or SIGNUM_FONT_PATH)")
	}
	if *subject == "" {
		fatal(logger, "subject is required")
	}
	if *status == "" {
		fatal(logger, "status is required")
	}
	if *color == "" {
		fatal(logger, "color is required")
	}
	if *output == "" {
		fatal(logger, "out is required")
	}

	badgeStyle := renderer.Style(*style)
	if !badgeStyle.IsValid() {
		fatal(logger, fmt.Sprintf("invalid style: %q", *style))
	}

	r, err := renderer.NewRenderer(*fontPath)
	if err != nil {
		fatal(logger, err.Error())
	}

	outputBytes, err := r.Render(renderer.Badge{
		Subject: *subject,
		Status:  *status,
		Color:   renderer.Color(*color),
		Style:   badgeStyle,
	})
	if err != nil {
		fatal(logger, err.Error())
	}

	err = os.WriteFile(*output, outputBytes, 0o600)
	if err != nil {
		fatal(logger, err.Error())
	}
}

func fatal(logger *slog.Logger, message string) {
	logger.Error(message)
	os.Exit(1)
}
