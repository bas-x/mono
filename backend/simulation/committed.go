package simulation

import "time"

import "github.com/bas-x/basex/geometry"

type CommittedState struct {
	entered   bool
	enteredAt time.Time
}

func (c *CommittedState) Name() string { return "Committed" }

func (c *CommittedState) Step(a *Aircraft, ctx FlightContext) AircraftState {
	now := ctx.Clock.Now()
	if !c.entered {
		c.entered = true
		c.enteredAt = now
	}
	if base, ok := findAssignedAirbase(ctx.Airbases, a.AssignedBase); ok {
		dist := geometry.Distance(a.Position, base.Location)
		if dist > 0 && a.Speed > 0 {
			step := a.Speed * ctx.Clock.Resolution.Seconds()
			if step >= dist {
				a.Position = base.Location
			} else {
				dir := base.Location.Sub(a.Position).Scale(1.0 / dist)
				a.Position = a.Position.Add(dir.Scale(step))
			}
		}
		if geometry.Distance(a.Position, base.Location) <= LandingRange {
			return &ServicingState{}
		}
	}
	if now.Sub(c.enteredAt) >= ctx.Lifecycle.Durations.CommitApproach {
		return &ServicingState{}
	}
	return c
}

func (c *CommittedState) Clone() AircraftState {
	if c == nil {
		return &CommittedState{}
	}
	clone := *c
	return &clone
}
