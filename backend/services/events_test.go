package services

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventBroadcaster_EmitAndUnsubscribe(t *testing.T) {
	t.Parallel()

	broadcaster := NewEventBroadcaster(1)
	clientID, events := broadcaster.Subscribe()
	expected := SimulationStepEvent{Type: EventTypeSimulationStep, SimulationID: BaseSimulationID, Tick: 1}

	broadcaster.Emit(expected)

	got := <-events
	require.Equal(t, expected, got)

	broadcaster.Unsubscribe(clientID)
	_, ok := <-events
	require.False(t, ok)
}

func TestEventBroadcaster_DropsSlowSubscribers(t *testing.T) {
	t.Parallel()

	broadcaster := NewEventBroadcaster(1)
	_, slowEvents := broadcaster.Subscribe()
	_, fastEvents := broadcaster.Subscribe()

	first := SimulationStepEvent{Type: EventTypeSimulationStep, SimulationID: BaseSimulationID, Tick: 1}
	second := SimulationStepEvent{Type: EventTypeSimulationStep, SimulationID: BaseSimulationID, Tick: 2}

	broadcaster.Emit(first)
	require.Equal(t, first, <-fastEvents)
	broadcaster.Emit(second)

	require.Equal(t, second, <-fastEvents)

	require.Equal(t, first, <-slowEvents)
	_, ok := <-slowEvents
	require.False(t, ok)
}
