package simulation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/prng"
)

func TestFleetInitGeneratesAircrafts(t *testing.T) {
	t.Parallel()
	ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	seed := [32]byte{0, 1, 2, 3}
	sim := NewSimulator(seed, ts)
	fleet := NewFleet()
	opts := FleetOptions{
		AircraftMin:    4,
		AircraftMax:    4,
		NeedsMin:       1,
		NeedsMax:       2,
		NeedsPool:      []NeedType{NeedFuel, NeedMunitions, NeedRepairs, NeedMaintenance},
		SeverityMin:    10,
		SeverityMax:    40,
		BlockingChance: prng.New(1, 2),
	}

	require.NoError(t, fleet.Init(sim.env, &opts))
	fleet.AssertInvariants()

	aircrafts := fleet.Aircrafts()
	require.Len(t, aircrafts, 4)
	seenTails := make(map[TailNumber]struct{}, len(aircrafts))

	for _, aircraft := range aircrafts {
		aircraft.AssertInvariants()
		require.NotEmpty(t, aircraft.Model)
		_, duplicate := seenTails[aircraft.TailNumber]
		require.False(t, duplicate, "duplicate tail number")
		seenTails[aircraft.TailNumber] = struct{}{}

		require.GreaterOrEqual(t, len(aircraft.Needs), int(opts.NeedsMin))
		require.LessOrEqual(t, len(aircraft.Needs), int(opts.NeedsMax))

		var seenMask uint64
		for _, need := range aircraft.Needs {
			idx, ok := NeedTypeIndex(need.Type)
			require.True(t, ok)
			mask := uint64(1) << idx
			require.Zero(t, seenMask&mask, "duplicate need %s", need.Type)
			seenMask |= mask
			require.Contains(t, opts.NeedsPool, need.Type)
			require.GreaterOrEqual(t, need.Severity, int(opts.SeverityMin))
			require.LessOrEqual(t, need.Severity, int(opts.SeverityMax))
			require.Equal(t, need.Type, need.RequiredCapability)
		}
	}
}

func TestFleetInitDeterministic(t *testing.T) {
	t.Parallel()
	seed := [32]byte{9, 9, 9}
	opts := FleetOptions{
		AircraftMin:    3,
		AircraftMax:    5,
		NeedsMin:       1,
		NeedsMax:       3,
		BlockingChance: prng.New(1, 3),
	}

	ts1 := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim1 := NewSimulator(seed, ts1)
	fleet1 := NewFleet()
	require.NoError(t, fleet1.Init(sim1.env, &opts))

	ts2 := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim2 := NewSimulator(seed, ts2)
	fleet2 := NewFleet()
	require.NoError(t, fleet2.Init(sim2.env, &opts))

	require.Equal(t, fleet1.Aircrafts(), fleet2.Aircrafts())
}

func TestFleetInitDefaultsToHighSeverityNeeds(t *testing.T) {
	t.Parallel()

	ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{4, 5, 6}, ts)
	fleet := NewFleet()
	opts := FleetOptions{
		AircraftMin: 2,
		AircraftMax: 2,
		NeedsMin:    1,
		NeedsMax:    2,
	}

	require.NoError(t, fleet.Init(sim.env, &opts))
	for _, aircraft := range fleet.Aircrafts() {
		require.NotEmpty(t, aircraft.Needs)
		for _, need := range aircraft.Needs {
			require.GreaterOrEqual(t, need.Severity, 60)
			require.LessOrEqual(t, need.Severity, 90)
		}
	}
}
