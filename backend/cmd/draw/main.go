package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bas-x/basex/prng"
	"github.com/bas-x/basex/simulation"
)

var (
	outputPath   = flag.String("out", "simulation.png", "output PNG file path")
	seedHex      = flag.String("seed", "", "optional 32-byte seed as hex")
	includeFlags = flag.String("regions", "", "comma-separated region names to include")
	minPerRegion = flag.Uint("min-per-region", 1, "minimum airbases per included region")
	maxPerRegion = flag.Uint("max-per-region", 3, "maximum airbases per included region")
	maxTotal     = flag.Uint("max-total", 32, "maximum total airbases")
	openBrowser  = flag.Bool("open", false, "open the generated image in the default viewer")
)

func main() {
	flag.Parse()

	seed, err := parseSeed(*seedHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid seed: %v\n", err)
		os.Exit(1)
	}

	include := parseRegions(*includeFlags)

	if *maxPerRegion < *minPerRegion {
		*maxPerRegion = *minPerRegion
	}
	if *maxTotal == 0 {
		*maxTotal = 1
	}

	ts := simulation.New(time.Millisecond, simulation.WithEpoch(time.Unix(0, 1)))
	sim := simulation.NewSimulator(seed, ts)

	options := &simulation.SimulationOptions{
		Airbases: simulation.AirbasesOptions{
			IncludeRegions:    include,
			MinPerRegion:      uint(*minPerRegion),
			MaxPerRegion:      uint(*maxPerRegion),
			MaxTotal:          uint(*maxTotal),
			RegionProbability: prng.New(1, 1),
		},
	}

	if err := sim.Init(options); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize simulation: %v\n", err)
		os.Exit(1)
	}

	img := simulation.Draw(sim)
	if err := writePNG(*outputPath, img); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write image: %v\n", err)
		os.Exit(1)
	}

	if *openBrowser {
		if err := openFile(*outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to open image: %v\n", err)
			os.Exit(1)
		}
	}
}

func parseSeed(hexSeed string) ([32]byte, error) {
	var seed [32]byte
	if hexSeed == "" {
		return seed, nil
	}
	bytes, err := hex.DecodeString(strings.TrimSpace(hexSeed))
	if err != nil {
		return seed, err
	}
	if len(bytes) != len(seed) {
		return seed, fmt.Errorf("expected %d bytes, got %d", len(seed), len(bytes))
	}
	copy(seed[:], bytes)
	return seed, nil
}

func parseRegions(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	regions := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		regions = append(regions, trimmed)
	}
	return regions
}

func writePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func openFile(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", abs)
	case "darwin":
		cmd = exec.Command("open", abs)
	default:
		cmd = exec.Command("xdg-open", abs)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}
