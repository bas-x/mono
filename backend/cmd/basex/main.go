package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bas-x/basex/cmd/basex/app"
	"github.com/bas-x/basex/services"
)

func main() {
	cfg, err := app.ParseConfig(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid basex config: %v\n", err)
		os.Exit(1)
	}

	runtime, err := app.New(app.RuntimeConfig{
		Config:  cfg,
		Service: services.NewSimulationService(services.SimulationServiceConfig{}),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create basex app: %v\n", err)
		os.Exit(1)
	}
	defer runtime.Close()

	ebiten.SetWindowSize(cfg.WindowWidth, cfg.WindowHeight)
	ebiten.SetWindowTitle(cfg.WindowTitle)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(runtime); err != nil && err != ebiten.Termination {
		fmt.Fprintf(os.Stderr, "basex exited with error: %v\n", err)
		os.Exit(1)
	}
}
