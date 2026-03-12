package app

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

const (
	defaultWindowWidth  = 1280
	defaultWindowHeight = 800
	defaultWindowTitle  = "Basex Local Tester"
)

type Config struct {
	WindowWidth    int
	WindowHeight   int
	WindowTitle    string
	Seed           string
	QASteps        []string
	DumpStatePath  string
	DumpEventsPath string
	ScreenshotOut  string
	ExitAfterQA    bool
}

func ParseConfig(args []string) (Config, error) {
	fs := flag.NewFlagSet("basex", flag.ContinueOnError)
	width := fs.Int("width", defaultWindowWidth, "window width")
	height := fs.Int("height", defaultWindowHeight, "window height")
	title := fs.String("title", defaultWindowTitle, "window title")
	seed := fs.String("seed", "kartick", "optional seed string")
	qaScript := fs.String("qa-script", "", "comma-separated QA steps")
	dumpState := fs.String("dump-state", "", "optional path for JSON state dump")
	dumpEvents := fs.String("dump-events", "", "optional path for full event history dump")
	screenshot := fs.String("screenshot-out", "", "optional path for rendered screenshot output")
	exitAfterQA := fs.Bool("exit-after-qa", false, "exit after completing QA script")
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	if *width <= 0 || *height <= 0 {
		return Config{}, fmt.Errorf("window size must be positive")
	}
	steps := make([]string, 0)
	if trimmed := strings.TrimSpace(*qaScript); trimmed != "" {
		for _, part := range strings.Split(trimmed, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			steps = append(steps, part)
		}
	}
	return Config{
		WindowWidth:    *width,
		WindowHeight:   *height,
		WindowTitle:    *title,
		Seed:           *seed,
		QASteps:        steps,
		DumpStatePath:  strings.TrimSpace(*dumpState),
		DumpEventsPath: strings.TrimSpace(*dumpEvents),
		ScreenshotOut:  strings.TrimSpace(*screenshot),
		ExitAfterQA:    *exitAfterQA,
	}, nil
}
