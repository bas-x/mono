package simulation

import "time"

type ServicingState struct {
	entered          bool
	enteredAt        time.Time
	lastNeedUpdateAt time.Time
}

func (s *ServicingState) Name() string { return "Servicing" }

func (s *ServicingState) Step(a *Aircraft, ctx FlightContext) AircraftState {
	now := ctx.Clock.Now()
	if !s.entered {
		s.entered = true
		s.enteredAt = now
		s.lastNeedUpdateAt = now
		if base, ok := findAssignedAirbase(ctx.Airbases, a.AssignedBase); ok {
			a.Position = base.Location
		}
		return s
	}
	if base, ok := findAssignedAirbase(ctx.Airbases, a.AssignedBase); ok {
		a.ApplyNeedPhase(s.consumeElapsed(now), NeedPhaseServicing, ctx.Lifecycle, &base)
	} else {
		a.ApplyNeedPhase(s.consumeElapsed(now), NeedPhaseServicing, ctx.Lifecycle, nil)
	}
	if now.Sub(s.enteredAt) >= ctx.Lifecycle.Durations.Servicing {
		a.ResetNeeds()
		a.ResetNeedRemainders()
		return &ReadyState{}
	}
	return s
}

func (s *ServicingState) Clone() AircraftState {
	if s == nil {
		return &ServicingState{}
	}
	clone := *s
	return &clone
}

func (s *ServicingState) consumeElapsed(now time.Time) time.Duration {
	var remainder time.Duration
	return consumeElapsed(&s.lastNeedUpdateAt, &remainder, now)
}
