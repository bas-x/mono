package simulation

import "github.com/bas-x/basex/assert"

type TailNumber [8]byte

type Aircraft struct {
	TailNumber TailNumber
	State      AircraftState
}

func NewAircraft(tn TailNumber, state AircraftState) Aircraft {
	assert.NotNil(state, "state")
	return Aircraft{
		TailNumber: tn,
		State:      state,
	}
}

func (a *Aircraft) Step() {
	nextState := a.State.Step(a)
	assert.NotNil(nextState, "next state")
	a.State = nextState
}

func (a *Aircraft) Clone() *Aircraft {
	return &Aircraft{
		TailNumber: a.TailNumber,
		State:      a.State.Clone(),
	}
}
