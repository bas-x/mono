package simulation

import "time"

const (
	EngagementRange        = 25.0
	LandingRange           = 15.0
	OrbitRadius            = 8.0
	OrbitAngleDeltaPerTick = 0.314
)

type FlightContext struct {
	Clock                 *TimeSim
	Dispatcher            *Dispatcher
	Airbases              []Airbase
	Lifecycle             LifecycleModel
	Threats               *ThreatSet
	ActiveThreats         *ThreatSet
	OnAircraftStateChange func(AircraftStateChangeEvent)
	OnThreatTargeted      func(ThreatTargetedEvent)
}

type AircraftState interface {
	Name() string
	Step(aircraft *Aircraft, ctx FlightContext) AircraftState
	Clone() AircraftState
}

func consumeElapsed(lastUpdatedAt *time.Time, remainder *time.Duration, now time.Time) time.Duration {
	if lastUpdatedAt.IsZero() {
		*lastUpdatedAt = now
		return 0
	}
	elapsed := now.Sub(*lastUpdatedAt) + *remainder
	if elapsed <= 0 {
		return 0
	}
	*remainder = 0
	*lastUpdatedAt = now
	return elapsed
}

func findAssignedAirbase(airbases []Airbase, id BaseID) (Airbase, bool) {
	for _, airbase := range airbases {
		if airbase.ID == id {
			return airbase, true
		}
	}
	return Airbase{}, false
}
