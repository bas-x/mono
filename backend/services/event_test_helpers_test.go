package services_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/bas-x/basex/services"
)

type eventExtractor[T services.Event] func(services.Event) (T, bool)

type serviceEventWatcher struct {
	events  <-chan services.Event
	pending []services.Event
}

func subscribeToServiceEvents(t *testing.T, svc *services.SimulationService) *serviceEventWatcher {
	t.Helper()

	clientID, events := svc.Broadcaster().Subscribe()
	t.Cleanup(func() {
		svc.Broadcaster().Unsubscribe(clientID)
	})

	return &serviceEventWatcher{events: events}
}

func takePendingMatching[T services.Event](w *serviceEventWatcher, eventType string, simulationID string, extract eventExtractor[T]) (T, bool) {
	var zero T

	for i, raw := range w.pending {
		event, ok := extract(raw)
		if !ok || event.EventType() != eventType || event.EventSimulationID() != simulationID {
			continue
		}

		w.pending = append(w.pending[:i], w.pending[i+1:]...)
		return event, true
	}

	return zero, false
}

func requireNextScopedEvent[T services.Event](
	t *testing.T,
	watcher *serviceEventWatcher,
	timeout time.Duration,
	eventType string,
	simulationID string,
	extract eventExtractor[T],
) T {
	t.Helper()

	if event, ok := takePendingMatching(watcher, eventType, simulationID, extract); ok {
		return event
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	mismatchCount := 0
	latestMismatch := ""

	for {
		select {
		case raw, ok := <-watcher.events:
			if !ok {
				t.Fatalf("event stream closed while waiting for %s event for simulation %q", eventType, simulationID)
			}

			event, typed := extract(raw)
			if typed && event.EventType() == eventType && event.EventSimulationID() == simulationID {
				return event
			}

			if typed && event.EventType() == eventType {
				mismatchCount++
				latestMismatch = fmt.Sprintf("%#v", event)
			}

			watcher.pending = append(watcher.pending, raw)
		case <-timer.C:
			if mismatchCount > 0 {
				t.Fatalf("timed out waiting for %s event for simulation %q after seeing %d same-type event(s) for other simulations; latest=%s", eventType, simulationID, mismatchCount, latestMismatch)
			}

			t.Fatalf("timed out waiting for %s event for simulation %q", eventType, simulationID)
		}
	}
}

func requireNoScopedEvent[T services.Event](
	t *testing.T,
	watcher *serviceEventWatcher,
	timeout time.Duration,
	eventType string,
	simulationID string,
	extract eventExtractor[T],
) {
	t.Helper()

	if event, ok := takePendingMatching(watcher, eventType, simulationID, extract); ok {
		t.Fatalf("unexpected buffered %s event for simulation %q: %#v", eventType, simulationID, event)
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case raw, ok := <-watcher.events:
			if !ok {
				t.Fatalf("event stream closed while asserting no %s event for simulation %q", eventType, simulationID)
			}

			event, typed := extract(raw)
			if typed && event.EventType() == eventType && event.EventSimulationID() == simulationID {
				t.Fatalf("unexpected %s event for simulation %q: %#v", eventType, simulationID, event)
			}

			watcher.pending = append(watcher.pending, raw)
		case <-timer.C:
			return
		}
	}
}

func asBranchCreatedEvent(event services.Event) (services.BranchCreatedEvent, bool) {
	typed, ok := event.(services.BranchCreatedEvent)
	return typed, ok
}

func asSimulationStepEvent(event services.Event) (services.SimulationStepEvent, bool) {
	typed, ok := event.(services.SimulationStepEvent)
	return typed, ok
}

func asSimulationEndedEvent(event services.Event) (services.SimulationEndedEvent, bool) {
	typed, ok := event.(services.SimulationEndedEvent)
	return typed, ok
}

func asAllAircraftPositionsEvent(event services.Event) (services.AllAircraftPositionsEvent, bool) {
	typed, ok := event.(services.AllAircraftPositionsEvent)
	return typed, ok
}

func requireNextBranchCreatedEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) services.BranchCreatedEvent {
	t.Helper()
	return requireNextScopedEvent(t, watcher, timeout, services.EventTypeBranchCreated, simulationID, asBranchCreatedEvent)
}

func requireNoBranchCreatedEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) {
	t.Helper()
	requireNoScopedEvent(t, watcher, timeout, services.EventTypeBranchCreated, simulationID, asBranchCreatedEvent)
}

func requireNextSimulationStepEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) services.SimulationStepEvent {
	t.Helper()
	return requireNextScopedEvent(t, watcher, timeout, services.EventTypeSimulationStep, simulationID, asSimulationStepEvent)
}

func requireNoSimulationStepEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) {
	t.Helper()
	requireNoScopedEvent(t, watcher, timeout, services.EventTypeSimulationStep, simulationID, asSimulationStepEvent)
}

func requireNextSimulationEndedEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) services.SimulationEndedEvent {
	t.Helper()
	return requireNextScopedEvent(t, watcher, timeout, services.EventTypeSimulationEnded, simulationID, asSimulationEndedEvent)
}

func requireNoSimulationEndedEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) {
	t.Helper()
	requireNoScopedEvent(t, watcher, timeout, services.EventTypeSimulationEnded, simulationID, asSimulationEndedEvent)
}

func requireNextAllAircraftPositionsEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) services.AllAircraftPositionsEvent {
	t.Helper()
	return requireNextScopedEvent(t, watcher, timeout, services.EventTypeAllAircraftPositions, simulationID, asAllAircraftPositionsEvent)
}

func requireNoAllAircraftPositionsEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) {
	t.Helper()
	requireNoScopedEvent(t, watcher, timeout, services.EventTypeAllAircraftPositions, simulationID, asAllAircraftPositionsEvent)
}
