package main

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

const (
	defaultWindowWidth  = 1280
	defaultWindowHeight = 800
	defaultWindowTitle  = "Text-Based Local Tester"
)

type launchOptions struct {
	windowWidth   int
	windowHeight  int
	windowTitle   string
	screenshotOut string
	exitAfterQA   bool
}

func parseLaunchOptions(args []string) (launchOptions, error) {
	flags := flag.NewFlagSet("text-based", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	width := flags.Int("width", defaultWindowWidth, "window width")
	height := flags.Int("height", defaultWindowHeight, "window height")
	title := flags.String("title", defaultWindowTitle, "window title")
	screenshotOut := flags.String("screenshot-out", "", "optional path for rendered screenshot output")
	exitAfterQA := flags.Bool("exit-after-qa", false, "exit after completing QA flow")

	if err := flags.Parse(args); err != nil {
		return launchOptions{}, err
	}
	if *width <= 0 {
		return launchOptions{}, fmt.Errorf("window width must be positive")
	}
	if *height <= 0 {
		return launchOptions{}, fmt.Errorf("window height must be positive")
	}

	return launchOptions{
		windowWidth:   *width,
		windowHeight:  *height,
		windowTitle:   strings.TrimSpace(*title),
		screenshotOut: strings.TrimSpace(*screenshotOut),
		exitAfterQA:   *exitAfterQA,
	}, nil
}
