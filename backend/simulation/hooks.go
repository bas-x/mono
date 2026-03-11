package simulation

import "time"

type AircraftStateChangeEvent struct {
	TailNumber TailNumber
	OldState   string
	NewState   string
	Aircraft   Aircraft
	Timestamp  time.Time
}

type LandingAssignmentEvent struct {
	TailNumber TailNumber
	Base       BaseID
	Source     LandingAssignmentSource
	Timestamp  time.Time
}

type SimulationStepEvent struct {
	Tick      uint64
	Timestamp time.Time
}

type AircraftStateChangeHook func(AircraftStateChangeEvent)

type LandingAssignmentHook func(LandingAssignmentEvent)

type SimulationStepHook func(SimulationStepEvent)

func safeInvoke[T any, H ~func(T)](hooks []H, event T) {
	for _, hook := range hooks {
		func(h H) {
			defer func() {
				_ = recover()
			}()
			h(event)
		}(hook)
	}
}
