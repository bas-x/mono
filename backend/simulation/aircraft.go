package simulation

import "github.com/bas-x/basex/assert"

type TailNumber [8]byte

type Aircraft struct {
	TailNumber    TailNumber
	Needs         []Need
	State         AircraftState
	AssignedBase  BaseID
	HasAssignment bool
}

func NewAircraft(
	tn TailNumber,
	state AircraftState,
	needs []Need,
) Aircraft {
	assert.NotNil(state, "state")
	if needs == nil {
		needs = make([]Need, 0)
	}
	clonedNeeds := make([]Need, len(needs))
	for i, need := range needs {
		clonedNeeds[i] = need.Clone()
	}
	aircraft := Aircraft{
		TailNumber: tn,
		State:      state.Clone(),
		Needs:      clonedNeeds,
	}
	aircraft.AssertInvariants()
	return aircraft
}

func (a *Aircraft) Step(ctx FlightContext) {
	a.AssertInvariants()
	oldState := a.State.Name()
	nextState := a.State.Step(a, ctx)
	assert.NotNil(nextState, "next state")
	a.State = nextState
	if oldState != a.State.Name() && ctx.OnAircraftStateChange != nil {
		ctx.OnAircraftStateChange(AircraftStateChangeEvent{
			TailNumber: a.TailNumber,
			OldState:   oldState,
			NewState:   a.State.Name(),
			Aircraft:   *a.Clone(),
			Timestamp:  ctx.Clock.Now(),
		})
	}
}

func (a *Aircraft) Clone() *Aircraft {
	clonedNeeds := make([]Need, len(a.Needs))
	for i, need := range a.Needs {
		clonedNeeds[i] = need.Clone()
	}
	return &Aircraft{
		TailNumber:    a.TailNumber,
		State:         a.State.Clone(),
		Needs:         clonedNeeds,
		AssignedBase:  a.AssignedBase,
		HasAssignment: a.HasAssignment,
	}
}

func (a *Aircraft) AssertInvariants() {
	assert.NotNil(a, "aircraft")
	assert.NotNil(a.State, "aircraft state")
	var seen uint64
	for _, need := range a.Needs {
		need.AssertInvariants()

		idx, ok := NeedTypeIndex(need.Type)
		assert.True(ok, "aircraft need type registered", need.Type)

		mask := uint64(1) << idx
		assert.True(seen&mask == 0, "aircraft needs unique", need.Type)

		seen |= mask
	}
	assert.NotNil(a.State, "aircraft state")
}
