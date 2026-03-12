package simulation_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sim "github.com/bas-x/basex/simulation"
)

func TestSimulationEndToEndLandingFlow(t *testing.T) {
	t.Parallel()

	ts := sim.New(10*time.Minute, sim.WithEpoch(time.Unix(0, 1)))
	simulation := sim.NewSimulator([32]byte{4}, ts)

	opts := &sim.SimulationOptions{
		ConstellationOpts: sim.ConstellationOptions{
			MaxTotal:     2,
			MinPerRegion: 2,
			MaxPerRegion: 2,
		},
		FleetOpts: sim.FleetOptions{
			AircraftMin: 1,
			AircraftMax: 1,
		},
	}
	require.NoError(t, simulation.Init(opts))

	bases := simulation.Airbases()
	require.GreaterOrEqual(t, len(bases), 2)

	aircraft := simulation.Aircrafts()
	require.Len(t, aircraft, 1)
	tail := aircraft[0].TailNumber

	stepUntil := func(target string, maxSteps int) {
		for i := 0; i < maxSteps; i++ {
			current := simulation.Aircrafts()
			require.NotEmpty(t, current)
			if current[0].State.Name() == target {
				return
			}
			simulation.Step()
		}
		t.Fatalf("timed out waiting for state %s", target)
	}

	stepUntil("Inbound", 50)

	assignment, err := simulation.RequestLanding(tail)
	require.NoError(t, err)
	require.Equal(t, sim.AssignmentSourceAlgorithm, assignment.Source)

	overrideBase := bases[1].ID
	replacement, err := simulation.OverrideLandingAssignment(tail, overrideBase)
	require.NoError(t, err)
	require.Equal(t, overrideBase, replacement.Base)
	require.Equal(t, sim.AssignmentSourceHuman, replacement.Source)

	stepUntil("Committed", 50)
	current := simulation.Aircrafts()
	require.True(t, current[0].HasAssignment)
	require.Equal(t, overrideBase, current[0].AssignedBase)

	stepUntil("Ready", 50)
	final := simulation.Aircrafts()[0]
	require.Equal(t, "Ready", final.State.Name())
	require.True(t, final.HasAssignment)
	require.Equal(t, overrideBase, final.AssignedBase)
}
