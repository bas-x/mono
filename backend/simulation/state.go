package simulation

type AircraftState interface {
	Name() string
	Step(aircraft *Aircraft) AircraftState
	Clone() AircraftState
}

type ReadyState struct {
}

func (r ReadyState) Name() string {
	return "Ready"
}

func (r ReadyState) Step(_ *Aircraft) AircraftState {
	return r
}

func (r ReadyState) Clone() AircraftState {
	return r
}
