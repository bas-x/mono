package services_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/prng"
	"github.com/bas-x/basex/services"
	"github.com/bas-x/basex/simulation"
)

func TestSimulationServiceEndToEnd_BaseReadModels(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{
		Options: &simulation.SimulationOptions{
			ConstellationOpts: simulation.ConstellationOptions{
				IncludeRegions:    []string{"Blekinge"},
				MinPerRegion:      2,
				MaxPerRegion:      2,
				MaxTotal:          2,
				RegionProbability: prng.New(1, 1),
			},
			FleetOpts: simulation.FleetOptions{
				AircraftMin: 2,
				AircraftMax: 2,
				NeedsMin:    1,
				NeedsMax:    2,
			},
		},
	})
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
	}
}

func TestSimulationServiceEndToEnd_LifecycleAndSimulationIDHandling(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{})

	_, err := svc.Airbases(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)

	_, err = svc.Aircrafts("branch-a")
	require.ErrorIs(t, err, services.ErrSimulationNotFound)

	_, err = svc.CreateBaseSimulation(services.BaseSimulationConfig{
		Options: &simulation.SimulationOptions{
			ConstellationOpts: simulation.ConstellationOptions{
				IncludeRegions:    []string{"Blekinge"},
				MinPerRegion:      1,
				MaxPerRegion:      1,
				MaxTotal:          1,
				RegionProbability: prng.New(1, 1),
			},
			FleetOpts: simulation.FleetOptions{
				AircraftMin: 1,
				AircraftMax: 1,
			},
		},
	})
	require.NoError(t, err)

	_, err = svc.CreateBaseSimulation(services.BaseSimulationConfig{
		Options: &simulation.SimulationOptions{
			ConstellationOpts: simulation.ConstellationOptions{
				IncludeRegions:    []string{"Blekinge"},
				MinPerRegion:      1,
				MaxPerRegion:      1,
				MaxTotal:          1,
				RegionProbability: prng.New(1, 1),
			},
			FleetOpts: simulation.FleetOptions{
				AircraftMin: 1,
				AircraftMax: 1,
			},
		},
	})
	require.ErrorIs(t, err, services.ErrBaseAlreadyExists)

	_, err = svc.Airbases("branch-a")
	require.ErrorIs(t, err, services.ErrSimulationNotFound)

	svc.Reset()

	_, err = svc.Aircrafts(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)
}
