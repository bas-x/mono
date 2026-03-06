package simulation

import (
	"math/rand/v2"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"
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
