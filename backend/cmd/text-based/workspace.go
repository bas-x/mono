package main

import (
	"errors"
	"fmt"
	"slices"
	"sort"

	"github.com/bas-x/basex/assets"
	"github.com/bas-x/basex/prng"
	"github.com/bas-x/basex/services"
	"github.com/bas-x/basex/simulation"
)

type simulationTab struct {
	id string
}

type tabSummary struct {
	aircraftCount int
	airbaseCount  int
	threatCount   int
	tick          uint64
	timestamp     string
	running       bool
	paused        bool
}

type shellWorkspace struct {
	tabs           []simulationTab
	activeTabID    string
	activeSummary  tabSummary
	baseInfo       services.SimulationInfo
	allSimulations []services.SimulationInfo
}

func bootstrapWorkspace(service *services.SimulationService) (shellWorkspace, error) {
	return bootstrapWorkspaceWithSeed(service, [32]byte{})
}

func bootstrapWorkspaceWithSeed(service *services.SimulationService, seed [32]byte) (shellWorkspace, error) {
	if service == nil {
		return shellWorkspace{}, fmt.Errorf("simulation service is required")
	}

	opts := defaultSimulationOptions()
	cfg := services.BaseSimulationConfig{Seed: seed, Options: opts}
	if _, err := service.CreateBaseSimulation(cfg); err != nil && !errors.Is(err, services.ErrBaseAlreadyExists) {
		return shellWorkspace{}, fmt.Errorf("create base simulation: %w", err)
	}

	return workspaceForActiveTab(service, services.BaseSimulationID)
}

func defaultSimulationOptions() *simulation.SimulationOptions {
	return &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    slices.Clone(assets.RegionNames),
			MinPerRegion:      1,
			MaxPerRegion:      3,
			MaxTotal:          15,
			RegionProbability: prng.New(1, 2),
		},
		FleetOpts: simulation.FleetOptions{
			AircraftMin: 6,
			AircraftMax: 40,
			NeedsMin:    1,
			NeedsMax:    uint(len(simulation.AllNeedTypes)),
			SeverityMin: 0,
			SeverityMax: 40,
		},
		ThreatOpts: simulation.ThreatOptions{
			SpawnChance: prng.New(1, 4),
			MaxActive:   3,
		},
	}
}

func workspaceForActiveTab(service *services.SimulationService, activeTabID string) (shellWorkspace, error) {
	if service == nil {
		return shellWorkspace{}, fmt.Errorf("simulation service is required")
	}

	baseInfo, err := service.Simulation(services.BaseSimulationID)
	if err != nil {
		return shellWorkspace{}, fmt.Errorf("load base simulation: %w", err)
	}
	if activeTabID == "" {
		activeTabID = baseInfo.ID
	}
	activeInfo, err := service.Simulation(activeTabID)
	if err != nil {
		return shellWorkspace{}, fmt.Errorf("load active simulation %q: %w", activeTabID, err)
	}

	summary, err := buildTabSummary(service, activeTabID, activeInfo)
	if err != nil {
		return shellWorkspace{}, fmt.Errorf("build tab summary for %q: %w", activeTabID, err)
	}

	simulations := service.Simulations()
	tabs := make([]simulationTab, 0, len(simulations))
	for _, sim := range simulations {
		tabs = append(tabs, simulationTab{id: sim.ID})
	}
	if len(tabs) == 0 {
		tabs = append(tabs, simulationTab{id: baseInfo.ID})
	}
	// Sort deterministically: base tab first, then branches alphabetically.
	sort.Slice(tabs, func(i, j int) bool {
		if tabs[i].id == services.BaseSimulationID {
			return true
		}
		if tabs[j].id == services.BaseSimulationID {
			return false
		}
		return tabs[i].id < tabs[j].id
	})

	return shellWorkspace{
		tabs:           tabs,
		activeTabID:    activeTabID,
		activeSummary:  summary,
		baseInfo:       baseInfo,
		allSimulations: simulations,
	}, nil
}

func branchBaseTab(service *services.SimulationService) (shellWorkspace, error) {
	if service == nil {
		return shellWorkspace{}, fmt.Errorf("simulation service is required")
	}

	branchID, err := service.BranchSimulation(services.BaseSimulationID)
	if err != nil {
		return shellWorkspace{}, fmt.Errorf("branch base simulation: %w", err)
	}

	return workspaceForActiveTab(service, branchID)
}

func buildTabSummary(service *services.SimulationService, simID string, info services.SimulationInfo) (tabSummary, error) {
	aircrafts, err := service.Aircrafts(simID)
	if err != nil {
		return tabSummary{}, err
	}
	airbases, err := service.Airbases(simID)
	if err != nil {
		return tabSummary{}, err
	}
	threats, err := service.Threats(simID)
	if err != nil {
		return tabSummary{}, err
	}

	ts := "—"
	if !info.Timestamp.IsZero() {
		ts = info.Timestamp.Format("15:04:05")
	}

	return tabSummary{
		aircraftCount: len(aircrafts),
		airbaseCount:  len(airbases),
		threatCount:   len(threats),
		tick:          info.Tick,
		timestamp:     ts,
		running:       info.Running,
		paused:        info.Paused,
	}, nil
}
