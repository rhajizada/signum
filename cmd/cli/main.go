package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
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

	if err := run(os.Args[1:], os.Stdout, os.Getenv); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, getenv func(string) string) error {
	fs := flag.NewFlagSet("signum", flag.ContinueOnError)
	fs.SetOutput(stdout)

	showVersion := fs.Bool("version", false, "Print version and exit")
	fontPath := fs.String("font", "", "Path to a .ttf font file (or set SIGNUM_FONT_PATH)")
	subject := fs.String("subject", "", "Badge subject text")
	status := fs.String("status", "", "Badge status text")
	color := fs.String("color", "", "Badge color (named or hex)")
	style := fs.String("style", "flat", "Badge style (flat, flat-square, plastic)")
	output := fs.String("out", "", "Output SVG file path")

	if len(args) == 0 {
		fs.Usage()
		return nil
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *showVersion {
		_, err := fmt.Fprintln(stdout, Version)
		return err
	}

	if *fontPath == "" {
		*fontPath = getenv("SIGNUM_FONT_PATH")
	}
	if *fontPath == "" {
		return errors.New("font is required (set -font or SIGNUM_FONT_PATH)")
	}
	if *subject == "" {
		return errors.New("subject is required")
	}
	if *status == "" {
		return errors.New("status is required")
	}
	if *color == "" {
		return errors.New("color is required")
	}

	badgeStyle := renderer.Style(*style)
	if !badgeStyle.IsValid() {
		return fmt.Errorf("invalid style: %q", *style)
	}

	r, err := renderer.NewRenderer(*fontPath)
	if err != nil {
		return fmt.Errorf("init renderer: %w", err)
	}

	outputBytes, err := r.Render(renderer.Badge{
		Subject: *subject,
		Status:  *status,
		Color:   renderer.Color(*color),
		Style:   badgeStyle,
	})
	if err != nil {
		return fmt.Errorf("render badge: %w", err)
	}

	if *output == "" {
		_, err = stdout.Write(outputBytes)
		if err != nil {
			return fmt.Errorf("write stdout: %w", err)
		}
		return nil
	}

	if err = os.WriteFile(*output, outputBytes, 0o600); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
