package simulation

import (
	"testing"
	"time"

	"github.com/bas-x/basex/geometry"
	"github.com/stretchr/testify/require"
)

func TestSimulationServicingEntryAndExitTimingSummaryUsesExactCompletedVisitDurations(t *testing.T) {
	t.Parallel()

	ts, ctx, aircraft, events, base := newServicingSummaryTestFixture()

	first := startServicingVisitFromCommitted(t, ts, aircraft, ctx, events)
	require.Equal(t, ts.Resolution, first.servicingEnteredAt.Sub(first.transitionIntoServicingAt))

	first = finishServicingVisit(t, ts, aircraft, ctx, events, first)
	require.Equal(t, 5*time.Second, first.transitionOutOfServicingAt.Sub(first.servicingEnteredAt))
	require.Equal(t, 6*time.Second, first.transitionOutOfServicingAt.Sub(first.transitionIntoServicingAt))

	resetAircraftForCommittedLanding(aircraft, ts, base)

	second := startServicingVisitFromCommitted(t, ts, aircraft, ctx, events)
	second = finishServicingVisit(t, ts, aircraft, ctx, events, second)

	summary := summarizeCompletedServicingVisits([]servicingVisitTiming{first, second})
	require.Equal(t, 2, summary.completedVisitCount)
	require.Equal(t, int64(10000), summary.totalDurationMs)
	require.NotNil(t, summary.averageDurationMs)
	require.Equal(t, int64(5000), *summary.averageDurationMs)
}

func TestSimulationServicingOpenIntervalsSummaryExcludesInProgressVisits(t *testing.T) {
	t.Parallel()

	ts, ctx, aircraft, events, base := newServicingSummaryTestFixture()

	completed := startServicingVisitFromCommitted(t, ts, aircraft, ctx, events)
	completed = finishServicingVisit(t, ts, aircraft, ctx, events, completed)

	resetAircraftForCommittedLanding(aircraft, ts, base)

	open := startServicingVisitFromCommitted(t, ts, aircraft, ctx, events)
	stepServicingSummaryAircraft(ts, aircraft, ctx)
	require.Equal(t, "Servicing", aircraft.State.Name())
	stepServicingSummaryAircraft(ts, aircraft, ctx)
	require.Equal(t, "Servicing", aircraft.State.Name())
	require.Zero(t, open.transitionOutOfServicingAt)
	require.Equal(t, 2*time.Second, ts.Now().Sub(open.servicingEnteredAt))

	summary := summarizeCompletedServicingVisits([]servicingVisitTiming{completed, open})
	require.Equal(t, 1, summary.completedVisitCount)
	require.Equal(t, int64(5000), summary.totalDurationMs)
	require.NotNil(t, summary.averageDurationMs)
	require.Equal(t, int64(5000), *summary.averageDurationMs)
}

type servicingVisitTiming struct {
	transitionIntoServicingAt  time.Time
	servicingEnteredAt         time.Time
	transitionOutOfServicingAt time.Time
}

type servicingSummaryMetric struct {
	completedVisitCount int
	totalDurationMs     int64
	averageDurationMs   *int64
}

func newServicingSummaryTestFixture() (*TimeSim, FlightContext, *Aircraft, *[]AircraftStateChangeEvent, Airbase) {
	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	lifecycle := testLifecycleModel()
	lifecycle.Durations.Servicing = 5 * time.Second

	base := Airbase{
		ID:       BaseID{1},
		Location: geometry.Point{X: 120, Y: 80},
	}

	events := make([]AircraftStateChangeEvent, 0, 8)
	ctx := FlightContext{
		Clock:     ts,
		Lifecycle: lifecycle,
		Airbases:  []Airbase{base},
		OnAircraftStateChange: func(event AircraftStateChangeEvent) {
			events = append(events, event)
		},
	}

	aircraft := NewAircraft(TailNumber{0, 0, 0, 0, 0, 0, 0, 41}, &CommittedState{entered: true, enteredAt: ts.Now()}, []Need{{
		Type:               NeedFuel,
		Severity:           80,
		RequiredCapability: NeedFuel,
	}})
	aircraft.AssignedBase = base.ID
	aircraft.HasAssignment = true
	aircraft.Position = base.Location

	return ts, ctx, &aircraft, &events, base
}

func startServicingVisitFromCommitted(t *testing.T, ts *TimeSim, aircraft *Aircraft, ctx FlightContext, events *[]AircraftStateChangeEvent) servicingVisitTiming {
	t.Helper()

	before := len(*events)
	stepServicingSummaryAircraft(ts, aircraft, ctx)
	require.Equal(t, "Servicing", aircraft.State.Name())
	require.Len(t, *events, before+1)

	transition := (*events)[before]
	require.Equal(t, "Committed", transition.OldState)
	require.Equal(t, "Servicing", transition.NewState)
	require.Equal(t, ts.Now(), transition.Timestamp)

	servicingState, ok := aircraft.State.(*ServicingState)
	require.True(t, ok)
	require.False(t, servicingState.entered)
	require.True(t, servicingState.enteredAt.IsZero())

	stepServicingSummaryAircraft(ts, aircraft, ctx)
	require.Equal(t, "Servicing", aircraft.State.Name())
	require.Len(t, *events, before+1)

	servicingState, ok = aircraft.State.(*ServicingState)
	require.True(t, ok)
	require.True(t, servicingState.entered)
	require.Equal(t, ts.Now(), servicingState.enteredAt)

	return servicingVisitTiming{
		transitionIntoServicingAt: transition.Timestamp,
		servicingEnteredAt:        servicingState.enteredAt,
	}
}

func finishServicingVisit(t *testing.T, ts *TimeSim, aircraft *Aircraft, ctx FlightContext, events *[]AircraftStateChangeEvent, visit servicingVisitTiming) servicingVisitTiming {
	t.Helper()

	for range 4 {
		stepServicingSummaryAircraft(ts, aircraft, ctx)
		require.Equal(t, "Servicing", aircraft.State.Name())
	}

	before := len(*events)
	stepServicingSummaryAircraft(ts, aircraft, ctx)
	require.Equal(t, "Ready", aircraft.State.Name())
	require.Len(t, *events, before+1)

	exit := (*events)[before]
	require.Equal(t, "Servicing", exit.OldState)
	require.Equal(t, "Ready", exit.NewState)
	require.Equal(t, ts.Now(), exit.Timestamp)

	visit.transitionOutOfServicingAt = exit.Timestamp
	return visit
}

func resetAircraftForCommittedLanding(aircraft *Aircraft, ts *TimeSim, base Airbase) {
	aircraft.State = &CommittedState{entered: true, enteredAt: ts.Now()}
	aircraft.AssignedBase = base.ID
	aircraft.HasAssignment = true
	aircraft.Position = base.Location
}

func stepServicingSummaryAircraft(ts *TimeSim, aircraft *Aircraft, ctx FlightContext) {
	ts.Tick()
	aircraft.Step(ctx)
}

func summarizeCompletedServicingVisits(visits []servicingVisitTiming) servicingSummaryMetric {
	var total time.Duration
	count := 0

	for _, visit := range visits {
		if visit.servicingEnteredAt.IsZero() || visit.transitionOutOfServicingAt.IsZero() {
			continue
		}
		if !visit.transitionOutOfServicingAt.After(visit.servicingEnteredAt) {
			continue
		}
		total += visit.transitionOutOfServicingAt.Sub(visit.servicingEnteredAt)
		count++
	}

	summary := servicingSummaryMetric{
		completedVisitCount: count,
		totalDurationMs:     total.Milliseconds(),
	}
	if count == 0 {
		return summary
	}

	averageDurationMs := summary.totalDurationMs / int64(count)
	summary.averageDurationMs = &averageDurationMs
	return summary
}
