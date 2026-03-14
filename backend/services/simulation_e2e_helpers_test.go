package services_test

import (
	"math/rand/v2"
	"time"

	"github.com/bas-x/basex/prng"
	"github.com/bas-x/basex/services"
	"github.com/bas-x/basex/simulation"
)

const pseudoE2ERunnerTicksPerSecond = 128

func newDeterministicService() *services.SimulationService {
	return services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
}

func newControlledRunnerService() *services.SimulationService {
	return services.NewSimulationService(services.SimulationServiceConfig{
		RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: pseudoE2ERunnerTicksPerSecond},
	})
}

func newPositionTrackingService() *services.SimulationService {
	return services.NewSimulationService(services.SimulationServiceConfig{Resolution: time.Second})
}

func deterministicBaseConfig(numBases, numAircraft uint) services.BaseSimulationConfig {
	return services.BaseSimulationConfig{Options: safeSimulationOptions(numBases, numAircraft)}
}

func deterministicBaseUntilTickConfig(numBases, numAircraft uint, untilTick int64) services.BaseSimulationConfig {
	config := deterministicBaseConfig(numBases, numAircraft)
	config.UntilTick = untilTick
	return config
}

func positionTrackingBaseConfig(seed [32]byte, numAircraft uint) services.BaseSimulationConfig {
	return services.BaseSimulationConfig{
		Seed:    seed,
		Options: positionTrackingSimulationOptions(numAircraft),
	}
}

func safeSimulationOptions(numBases, numAircraft uint) *simulation.SimulationOptions {
	return &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      numBases,
			MaxPerRegion:      numBases,
			MaxTotal:          numBases,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{
			AircraftMin: numAircraft,
			AircraftMax: numAircraft,
			NeedsMin:    1,
			NeedsMax:    2,
		},
	}
}

func positionTrackingSimulationOptions(numAircraft uint) *simulation.SimulationOptions {
	lifecycle := simulation.DefaultLifecycleModel()
	lifecycle.Durations.Ready = 0
	return &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      1,
			MaxPerRegion:      1,
			MaxTotal:          1,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{
			AircraftMin: numAircraft,
			AircraftMax: numAircraft,
			NeedsMin:    1,
			NeedsMax:    1,
			StateFactory: func(*rand.Rand) simulation.AircraftState {
				return &simulation.ReadyState{}
			},
		},
		ThreatOpts: simulation.ThreatOptions{
			SpawnChance: prng.New(1, 1),
			MaxActive:   numAircraft,
		},
		LifecycleOpts: &lifecycle,
	}
}

func inboundOverrideBaseConfig() services.BaseSimulationConfig {
	return services.BaseSimulationConfig{Options: inboundOverrideSimulationOptions()}
}

func inboundOverrideSimulationOptions() *simulation.SimulationOptions {
	options := safeSimulationOptions(2, 1)
	options.FleetOpts.StateFactory = func(_ *rand.Rand) simulation.AircraftState {
		return &simulation.InboundState{}
	}
	options.LifecycleOpts = &simulation.LifecycleModel{
		Durations: simulation.PhaseDurations{
			Outbound:        5 * time.Second,
			Engaged:         5 * time.Second,
			InboundDecision: 3 * time.Second,
			CommitApproach:  4 * time.Second,
			Servicing:       6 * time.Second,
			Ready:           2 * time.Second,
		},
		ReturnThreshold: 80,
		NeedRates: map[simulation.NeedType]simulation.NeedRateModel{
			simulation.NeedFuel:      {OutboundMilliPerHour: 18000000, EngagedMilliPerHour: 18000000, ServicingMilliPerHour: 28800000, VariancePermille: 0},
			simulation.NeedMunitions: {OutboundMilliPerHour: 18000000, EngagedMilliPerHour: 18000000, ServicingMilliPerHour: 18000000, VariancePermille: 0},
		},
	}
	return options
}
