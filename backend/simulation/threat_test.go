package simulation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/prng"
)

func TestThreatSpawnEdgeProducesPositionOnMapBounds(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	env := NewSimulator([32]byte{9, 9, 9}, ts).env
	set := NewThreatSet()

	threat, ok := set.TrySpawnEdge(env, ThreatOptions{SpawnChance: prng.New(1, 1), MaxActive: 16}, 0)
	require.True(t, ok)
	require.NotZero(t, threat.Position.X+threat.Position.Y)
	threat.AssertInvariants()
}

func TestThreatSpawnEdgeRespectsMaxActive(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	env := NewSimulator([32]byte{5, 5, 5}, ts).env
	set := NewThreatSet()

	for i := uint64(0); i < 3; i++ {
		_, ok := set.TrySpawnEdge(env, ThreatOptions{SpawnChance: prng.New(1, 1), MaxActive: 3}, i)
		require.True(t, ok)
	}

	_, ok := set.TrySpawnEdge(env, ThreatOptions{SpawnChance: prng.New(1, 1), MaxActive: 3}, 3)
	require.False(t, ok)
}

func TestThreatDespawnActiveRemovesExpiredThreats(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	env := NewSimulator([32]byte{7, 7, 7}, ts).env
	set := NewThreatSet()

	threat, ok := set.TrySpawnEdge(env, ThreatOptions{SpawnChance: prng.New(1, 1), MaxActive: 4}, 0)
	require.True(t, ok)
	set.Activate(threat)

	despawned := set.DespawnActive(200, 100)
	require.Len(t, despawned, 1)
	require.Equal(t, threat.ID, despawned[0].ID)
	require.False(t, set.IsActive(threat.ID))
}

func TestThreatNextTargetUsesRoundRobinWithoutRemovingThreats(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	env := NewSimulator([32]byte{3, 3, 3}, ts).env
	set := NewThreatSet()

	first, ok := set.TrySpawnEdge(env, ThreatOptions{SpawnChance: prng.New(1, 1), MaxActive: 4}, 0)
	require.True(t, ok)
	second, ok := set.TrySpawnEdge(env, ThreatOptions{SpawnChance: prng.New(1, 1), MaxActive: 4}, 1)
	require.True(t, ok)

	target1, ok := set.NextTarget()
	require.True(t, ok)
	target2, ok := set.NextTarget()
	require.True(t, ok)
	target3, ok := set.NextTarget()
	require.True(t, ok)

	require.Equal(t, first.ID, target1.ID)
	require.Equal(t, second.ID, target2.ID)
	require.Equal(t, first.ID, target3.ID)
	require.Len(t, set.Pending(), 2)
}
