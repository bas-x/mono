package simulation

import (
	"math/rand/v2"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/assets"
	"github.com/bas-x/basex/geometry"
	"github.com/bas-x/basex/prng"
)

func TestRandSourceCopy(t *testing.T) {
	src := rand.NewChaCha8([32]byte{1, 2, 3})

	_ = src.Uint64()
	_ = src.Uint64()
	_ = src.Uint64()

	srcCopy := *src

	v1 := src.Uint64()
	v2 := srcCopy.Uint64()

	if v1 != v2 {
		t.Errorf("expected both sources to produce the same value, got %d and %d", v1, v2)
	}
}

func TestSimulation_SimpleRun(t *testing.T) {
	ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{1, 2, 3}, ts)
	runner := NewBasicRunner(BasicRunnerConfig{})
	runner.untilTick = 10

	runner.Run(t.Context(), sim)
	runner.AssertInvariants()
	sim.AssertInvariants()

	require.False(t, runner.active.Load())
}

func TestSimulation_IdenticalStepTagsAfterClone(t *testing.T) {
	t.Run("branch before start", func(t *testing.T) {
		ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
		sim := NewSimulator([32]byte{1, 2, 3}, ts)
		clone := sim.Clone()

		sim.step()
		clone.step()

		require.Equal(t, sim.stepTag, clone.stepTag)
	})

	t.Run("branch after start", func(t *testing.T) {
		ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
		sim := NewSimulator([32]byte{1, 2, 3}, ts)
		sim.step()

		clone := sim.Clone()
		require.Equal(t, sim.stepTag, clone.stepTag)

		sim.step()
		clone.step()

		require.Equal(t, sim.stepTag, clone.stepTag)
	})
}

func TestSimulation_RunnerPauseAndBranch(t *testing.T) {
	t.Run("branch while paused", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
			runnerConfig := ControlledRunnerConfig{
				TicksPerSecond: 64,
				UntilTick:      128,
			}

			sim := NewSimulator([32]byte{1, 2, 3}, ts)
			simRunner := NewControlledRunner(runnerConfig)
			go simRunner.Run(t.Context(), sim)

			time.Sleep(100 * time.Millisecond)
			simRunner.Pause()

			clone := sim.Clone()
			cloneRunner := NewControlledRunner(runnerConfig)
			go cloneRunner.Run(t.Context(), clone)

			simRunner.Unpause()
			time.Sleep(3 * time.Second)

			sim.AssertInvariants()
			simRunner.AssertInvariants()
			clone.AssertInvariants()
			cloneRunner.AssertInvariants()

			require.Equal(t, sim.stepTag, clone.stepTag)
			require.False(t, simRunner.active.Load())
			require.False(t, cloneRunner.active.Load())
		})
	})
}

func TestSimulationInitAirbases(t *testing.T) {
	ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{9, 9, 9}, ts)
	options := &SimulationOptions{
		Airbases: AirbasesOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      2,
			MaxPerRegion:      3,
			MaxTotal:          5,
			RegionProbability: prng.New(1, 1),
		},
	}

	require.NoError(t, sim.Init(options))
	bases := sim.Airbases()
	require.NotEmpty(t, bases)
	require.LessOrEqual(t, len(bases), 5)

	region := findRegionByName(t, "Blekinge")
	for _, base := range bases {
		require.Equal(t, "Blekinge", base.Region)
		require.Truef(t, pointInsideRegion(base.Location, region), "base %+v not inside region", base)
	}
}

func TestSimulationInitDeterministic(t *testing.T) {
	seed := [32]byte{1, 2, 3}
	opts := &SimulationOptions{
		Airbases: AirbasesOptions{
			IncludeRegions:    []string{"Blekinge", "Gotland"},
			MinPerRegion:      1,
			MaxPerRegion:      2,
			MaxTotal:          6,
			RegionProbability: prng.New(1, 1),
		},
	}

	ts1 := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim1 := NewSimulator(seed, ts1)
	require.NoError(t, sim1.Init(opts))

	ts2 := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim2 := NewSimulator(seed, ts2)
	require.NoError(t, sim2.Init(opts))

	require.Equal(t, sim1.Airbases(), sim2.Airbases())
}

func TestSimulationInitRespectsMaxTotal(t *testing.T) {
	ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{5, 5, 5}, ts)
	opts := &SimulationOptions{
		Airbases: AirbasesOptions{
			IncludeRegions:    []string{"Blekinge", "Gotland", "Halland"},
			MinPerRegion:      1,
			MaxPerRegion:      5,
			MaxTotal:          3,
			RegionProbability: prng.New(1, 1),
		},
	}
	require.NoError(t, sim.Init(opts))
	require.LessOrEqual(t, len(sim.Airbases()), 3)
}

func pointInsideRegion(point geometry.Point, region assets.Region) bool {
	for _, area := range region.Areas {
		poly := toGeometryPolygon(area)
		if pointInPolygonGeometry(point, poly) {
			return true
		}
	}
	return false
}

func findRegionByName(t *testing.T, name string) assets.Region {
	t.Helper()
	for _, region := range assets.Regions {
		if strings.EqualFold(region.Name, name) {
			return region
		}
	}
	t.Fatalf("region %s not found", name)
	return assets.Region{}
}
