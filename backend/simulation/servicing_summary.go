package simulation

import "time"

type ServicingSummary struct {
	CompletedVisitCount int64
	TotalDurationMs     int64
	AverageDurationMs   *int64
}

type servicingSummaryAccumulator struct {
	completedVisitCount int64
	totalDurationMs     int64
	activeVisits        map[TailNumber]time.Time
}

func newServicingSummaryAccumulator() *servicingSummaryAccumulator {
	return &servicingSummaryAccumulator{
		activeVisits: make(map[TailNumber]time.Time),
	}
}

func (a *servicingSummaryAccumulator) Clone() *servicingSummaryAccumulator {
	if a == nil {
		return newServicingSummaryAccumulator()
	}
	cloned := &servicingSummaryAccumulator{
		completedVisitCount: a.completedVisitCount,
		totalDurationMs:     a.totalDurationMs,
		activeVisits:        make(map[TailNumber]time.Time, len(a.activeVisits)),
	}
	for tail, enteredAt := range a.activeVisits {
		cloned.activeVisits[tail] = enteredAt
	}
	return cloned
}

func (a *servicingSummaryAccumulator) Record(event AircraftStateChangeEvent, resolution time.Duration) {
	if a == nil || resolution <= 0 {
		return
	}
	switch {
	case event.NewState == "Servicing":
		a.activeVisits[event.TailNumber] = event.Timestamp.Add(resolution)
	case event.OldState == "Servicing":
		enteredAt, ok := a.activeVisits[event.TailNumber]
		if !ok {
			return
		}
		delete(a.activeVisits, event.TailNumber)
		if !event.Timestamp.After(enteredAt) {
			return
		}
		a.completedVisitCount++
		a.totalDurationMs += event.Timestamp.Sub(enteredAt).Milliseconds()
	}
}

func (a *servicingSummaryAccumulator) Summary() ServicingSummary {
	if a == nil {
		return ServicingSummary{}
	}
	summary := ServicingSummary{
		CompletedVisitCount: a.completedVisitCount,
		TotalDurationMs:     a.totalDurationMs,
	}
	if a.completedVisitCount == 0 {
		return summary
	}
	averageDurationMs := a.totalDurationMs / a.completedVisitCount
	summary.AverageDurationMs = &averageDurationMs
	return summary
}
