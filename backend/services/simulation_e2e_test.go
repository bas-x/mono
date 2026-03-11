package services_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/prng"
	"github.com/bas-x/basex/services"
	"github.com/bas-x/basex/simulation"
)

func TestSimulationServiceEndToEnd_BaseReadModels(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(2, 2)})
	require.NoError(t, err)

	airbases, err := svc.Airbases(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, airbases, 2)
	for _, airbase := range airbases {
		require.NotEmpty(t, airbase.RegionID)
		require.NotEmpty(t, airbase.Region)
		require.Len(t, airbase.ID, 16)
		_, decodeErr := hex.DecodeString(airbase.ID)
		require.NoError(t, decodeErr)
	}

	aircrafts, err := svc.Aircrafts(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, aircrafts, 2)
	for _, aircraft := range aircrafts {
		require.Len(t, aircraft.TailNumber, 16)
		_, decodeErr := hex.DecodeString(aircraft.TailNumber)
		require.NoError(t, decodeErr)
		require.NotEmpty(t, aircraft.State)
		require.NotNil(t, aircraft.Needs)
		require.NotEmpty(t, aircraft.Needs)
	}
}

func TestSimulationServiceEndToEnd_LifecycleAndSimulationIDHandling(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{})

	_, err := svc.Airbases(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)

	_, err = svc.Aircrafts("branch-a")
	require.ErrorIs(t, err, services.ErrSimulationNotFound)

	_, err = svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	_, err = svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.ErrorIs(t, err, services.ErrBaseAlreadyExists)

	_, err = svc.Airbases("branch-a")
	require.ErrorIs(t, err, services.ErrSimulationNotFound)

	svc.Reset()

	_, err = svc.Aircrafts(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)
}

func TestSimulationServiceEndToEnd_EmitsEvents(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{})
	_, events := svc.Broadcaster().Subscribe()
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	base, ok := svc.Base()
	require.True(t, ok)

	var sawStep bool
	var sawStateChange bool

	for range 12 {
		base.Step()
	drainLoop:
		for {
			select {
			case raw := <-events:
				switch event := raw.(type) {
				case services.SimulationStepEvent:
					sawStep = true
					require.Equal(t, services.BaseSimulationID, event.SimulationID)
				case services.AircraftStateChangeEvent:
					sawStateChange = true
					require.Equal(t, services.BaseSimulationID, event.SimulationID)
					require.NotEmpty(t, event.Aircraft.TailNumber)
				}
			case <-time.After(10 * time.Millisecond):
				break drainLoop
			}
		}
		if sawStep && sawStateChange {
			break
		}
	}

	require.True(t, sawStep)
	require.True(t, sawStateChange)
}

func TestSimulationServiceEndToEnd_StartSimulationAndStatus(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{
		RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 128},
	})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	list := svc.Simulations()
	require.Len(t, list, 1)
	require.Equal(t, services.BaseSimulationID, list[0].ID)
	require.False(t, list[0].Running)

	_, events := svc.Broadcaster().Subscribe()
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))
	require.ErrorIs(t, svc.StartSimulation(services.BaseSimulationID), services.ErrSimulationRunning)

	deadline := time.After(2 * time.Second)
	for {
		select {
		case raw := <-events:
			step, ok := raw.(services.SimulationStepEvent)
			if !ok {
				continue
			}
			require.Equal(t, services.BaseSimulationID, step.SimulationID)
			require.Greater(t, step.Tick, uint64(0))
			require.True(t, svc.Simulations()[0].Running)
			return
		case <-deadline:
			t.Fatal("expected simulation step event from running simulation")
		}
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
