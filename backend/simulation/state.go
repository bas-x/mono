package simulation

import "time"

type FlightContext struct {
	Clock      *TimeSim
	Dispatcher *Dispatcher
}

type AircraftState interface {
	Name() string
	Step(aircraft *Aircraft, ctx FlightContext) AircraftState
	Clone() AircraftState
}

const (
	outboundDuration       = 5 * time.Second
	engagedDuration        = 5 * time.Second
	inboundDecisionDelay   = 3 * time.Second
	commitApproachDuration = 4 * time.Second
	servicingDuration      = 6 * time.Second
)

type ReadyState struct {
	entered   bool
	enteredAt time.Time
}

func (r *ReadyState) Name() string { return "Ready" }

func (r *ReadyState) Step(_ *Aircraft, ctx FlightContext) AircraftState {
	if !r.entered {
		r.entered = true
		r.enteredAt = ctx.Clock.Now()
	}
	return r
}

func (r *ReadyState) Clone() AircraftState {
	if r == nil {
		return &ReadyState{}
	}
	clone := *r
	return &clone
}

type OutboundState struct {
	entered   bool
	enteredAt time.Time
}

func (o *OutboundState) Name() string { return "Outbound" }

func (o *OutboundState) Step(_ *Aircraft, ctx FlightContext) AircraftState {
	now := ctx.Clock.Now()
	if !o.entered {
		o.entered = true
		o.enteredAt = now
		return o
	}
	if now.Sub(o.enteredAt) >= outboundDuration {
		return &EngagedState{}
	}
	return o
}

func (o *OutboundState) Clone() AircraftState {
	if o == nil {
		return &OutboundState{}
	}
	clone := *o
	return &clone
}

type EngagedState struct {
	entered   bool
	enteredAt time.Time
}

func (e *EngagedState) Name() string { return "Engaged" }

func (e *EngagedState) Step(_ *Aircraft, ctx FlightContext) AircraftState {
	now := ctx.Clock.Now()
	if !e.entered {
		e.entered = true
		e.enteredAt = now
		return e
	}
	if now.Sub(e.enteredAt) >= engagedDuration {
		return &InboundState{}
	}
	return e
}

func (e *EngagedState) Clone() AircraftState {
	if e == nil {
		return &EngagedState{}
	}
	clone := *e
	return &clone
}

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
	if i.assigned && now.Sub(i.enteredAt) >= inboundDecisionDelay {
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
	if now.Sub(c.enteredAt) >= commitApproachDuration {
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

type ServicingState struct {
	entered   bool
	enteredAt time.Time
}

func (s *ServicingState) Name() string { return "Servicing" }

func (s *ServicingState) Step(_ *Aircraft, ctx FlightContext) AircraftState {
	now := ctx.Clock.Now()
	if !s.entered {
		s.entered = true
		s.enteredAt = now
		return s
	}
	if now.Sub(s.enteredAt) >= servicingDuration {
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
