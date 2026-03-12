package simulation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/prng"
)

func TestThreatSpawnFromConfiguredRegionsNotLimitedByBases(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{9, 9, 9}, ts)
	regions := threatRegionsAll()
	require.Greater(t, len(regions), 2)

	set := NewThreatSet()
	seen := map[string]struct{}{}
	for i := uint64(0); i < 64; i++ {
		threat, ok := set.TrySpawnFromRegions(sim.env, regions, ThreatOptions{SpawnChance: prng.New(1, 1), MaxActive: 16}, i)
		if !ok {
			break
		}
		seen[threat.Region] = struct{}{}
	}

	require.Greater(t, len(seen), 2)
}

func TestSimulationThreatRegionsRestrictedToConstellation(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{5, 5, 5}, ts)
	require.NoError(t, sim.Init(&SimulationOptions{
		ConstellationOpts: ConstellationOptions{IncludeRegions: []string{"Blekinge"}},
	}))

	require.Len(t, sim.threatRegions, 1)
	require.Equal(t, "Blekinge", sim.threatRegions[0].Region)
}
