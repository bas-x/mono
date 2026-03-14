package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bas-x/basex/services"
)

func main() {
	options, err := parseLaunchOptions(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid text-based options: %v\n", err)
		os.Exit(1)
	}

	shell, err := newTextShell(options, services.NewSimulationService(services.SimulationServiceConfig{}))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create text shell: %v\n", err)
		os.Exit(1)
	}
	defer shell.Close()

	ebiten.SetWindowSize(options.windowWidth, options.windowHeight)
	ebiten.SetWindowTitle(options.windowTitle)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.MaximizeWindow()

	if err := ebiten.RunGame(shell); err != nil && err != ebiten.Termination {
		fmt.Fprintf(os.Stderr, "text-based exited with error: %v\n", err)
		os.Exit(1)
	}
}
