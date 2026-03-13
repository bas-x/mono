package simulation

import (
	"time"

	"github.com/bas-x/basex/geometry"
)

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

type ThreatSpawnedEvent struct {
	Threat    Threat
	Timestamp time.Time
}

type ThreatTargetedEvent struct {
	Threat     Threat
	TailNumber TailNumber
	Timestamp  time.Time
}

type ThreatDespawnedEvent struct {
	Threat    Threat
	Timestamp time.Time
}

type AircraftStateChangeHook func(AircraftStateChangeEvent)

type LandingAssignmentHook func(LandingAssignmentEvent)

type SimulationStepHook func(SimulationStepEvent)

type ThreatSpawnedHook func(ThreatSpawnedEvent)

type ThreatTargetedHook func(ThreatTargetedEvent)

type ThreatDespawnedHook func(ThreatDespawnedEvent)

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

// AircraftPositionSnapshot captures the position of a single aircraft at a point in time.
type AircraftPositionSnapshot struct {
	TailNumber TailNumber
	Position   geometry.Point
	State      string
	Needs      []Need
}

// AllAircraftPositionsEvent is emitted every simulation tick with the current
// position of every aircraft in the fleet.
type AllAircraftPositionsEvent struct {
	Tick      uint64
	Timestamp time.Time
	Positions []AircraftPositionSnapshot
}

type AllAircraftPositionsHook func(AllAircraftPositionsEvent)
