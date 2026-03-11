package services

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/prng"
	"github.com/bas-x/basex/simulation"
)

func TestSimulationService_AirbasesAndAircrafts(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{})
	_, err := svc.CreateBaseSimulation(BaseSimulationConfig{
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

	bases, err := svc.Airbases(BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, bases, 1)
	require.Len(t, bases[0].ID, 16)
	_, err = hex.DecodeString(bases[0].ID)
	require.NoError(t, err)

	aircrafts, err := svc.Aircrafts(BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, aircrafts, 1)
	require.Len(t, aircrafts[0].TailNumber, 16)
	_, err = hex.DecodeString(aircrafts[0].TailNumber)
	require.NoError(t, err)
	require.NotEmpty(t, aircrafts[0].State)
}

func TestSimulationService_UnknownSimulationID(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{})

	_, err := svc.Airbases("branch-1")
	require.ErrorIs(t, err, ErrSimulationNotFound)

	_, err = svc.Aircrafts("branch-1")
	require.ErrorIs(t, err, ErrSimulationNotFound)
}
