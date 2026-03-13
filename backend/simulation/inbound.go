package simulation

import "time"

import "github.com/bas-x/basex/geometry"

type InboundState struct {
	entered    bool
	enteredAt  time.Time
	assignment LandingAssignment
	assigned   bool
}

func (i *InboundState) Name() string { return "Inbound" }

func (i *InboundState) Step(a *Aircraft, ctx FlightContext) AircraftState {
	now := ctx.Clock.Now()
	if !i.entered {
		i.entered = true
		i.enteredAt = now
	}
	assignment, ok := ctx.Dispatcher.AssignmentFor(a.TailNumber)
	if !ok {
		if got, err := ctx.Dispatcher.RegisterInbound(a.TailNumber); err == nil {
			assignment = got
			ok = true
		}
	}
	if ok {
		i.assigned = true
		i.assignment = assignment
	}
	if i.assigned {
		if base, ok := findAssignedAirbase(ctx.Airbases, i.assignment.Base); ok {
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
		}
	}
	if i.assigned && now.Sub(i.enteredAt) >= ctx.Lifecycle.Durations.InboundDecision {
		ctx.Dispatcher.RemoveInbound(a.TailNumber)
		a.AssignedBase = i.assignment.Base
		a.HasAssignment = true
		return &CommittedState{}
	}
	return i
}

func (i *InboundState) Clone() AircraftState {
	if i == nil {
		return &InboundState{}
	}
	clone := *i
	return &clone
}
