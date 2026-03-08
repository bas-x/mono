package simulation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDispatcherAssignsRoundRobin(t *testing.T) {
	t.Parallel()
	constellation := &Constellation{
		airbases: []Airbase{
			{ID: BaseID{0, 0, 0, 0, 0, 0, 0, 1}},
			{ID: BaseID{0, 0, 0, 0, 0, 0, 0, 2}},
		},
	}
	dispatcher := NewDispatcher(constellation, &RoundRobinAssigner{})

	assignment, err := dispatcher.RegisterInbound(TailNumber{0, 0, 0, 0, 0, 0, 0, 1})
	require.NoError(t, err)
	require.Equal(t, BaseID{0, 0, 0, 0, 0, 0, 0, 1}, assignment.Base)
	require.Equal(t, AssignmentSourceAlgorithm, assignment.Source)

	assignment, err = dispatcher.RegisterInbound(TailNumber{0, 0, 0, 0, 0, 0, 0, 2})
	require.NoError(t, err)
	require.Equal(t, BaseID{0, 0, 0, 0, 0, 0, 0, 2}, assignment.Base)

	assignment, err = dispatcher.RegisterInbound(TailNumber{0, 0, 0, 0, 0, 0, 0, 3})
	require.NoError(t, err)
	require.Equal(t, BaseID{0, 0, 0, 0, 0, 0, 0, 1}, assignment.Base)
}

func TestDispatcherOverrides(t *testing.T) {
	t.Parallel()
	constellation := &Constellation{
		airbases: []Airbase{
			{ID: BaseID{1}},
			{ID: BaseID{2}},
		},
	}
	dispatcher := NewDispatcher(constellation, &RoundRobinAssigner{})
	tail := TailNumber{9}

	// Initial registration uses algorithmic assignment.
	assignment, err := dispatcher.RegisterInbound(tail)
	require.NoError(t, err)
	require.Equal(t, BaseID{1}, assignment.Base)
	require.Equal(t, AssignmentSourceAlgorithm, assignment.Source)

	// Human override switches to the other base.
	override, err := dispatcher.OverrideAssignment(tail, BaseID{2})
	require.NoError(t, err)
	require.Equal(t, BaseID{2}, override.Base)
	require.Equal(t, AssignmentSourceHuman, override.Source)

	// Clearing the override reverts back to the nearest base.
	reassigned, err := dispatcher.ClearOverride(tail)
	require.NoError(t, err)
	require.Equal(t, BaseID{2}, reassigned.Base)
	require.Equal(t, AssignmentSourceAlgorithm, reassigned.Source)
}
