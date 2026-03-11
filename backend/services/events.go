package services

import "time"

const (
	EventTypeAircraftStateChange = "aircraft_state_change"
	EventTypeLandingAssignment   = "landing_assignment"
	EventTypeSimulationStep      = "simulation_step"
)

type Event interface {
	EventType() string
	EventSimulationID() string
}

type AircraftStateChangeEvent struct {
	Type         string    `json:"type"`
	SimulationID string    `json:"simulationId"`
	TailNumber   string    `json:"tailNumber"`
	OldState     string    `json:"oldState"`
	NewState     string    `json:"newState"`
	Aircraft     Aircraft  `json:"aircraft"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e AircraftStateChangeEvent) EventType() string {
	return e.Type
}

func (e AircraftStateChangeEvent) EventSimulationID() string {
	return e.SimulationID
}

type LandingAssignmentEvent struct {
	Type         string    `json:"type"`
	SimulationID string    `json:"simulationId"`
	TailNumber   string    `json:"tailNumber"`
	BaseID       string    `json:"baseId"`
	Source       string    `json:"source"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e LandingAssignmentEvent) EventType() string {
	return e.Type
}

func (e LandingAssignmentEvent) EventSimulationID() string {
	return e.SimulationID
}

type SimulationStepEvent struct {
	Type         string    `json:"type"`
	SimulationID string    `json:"simulationId"`
	Tick         uint64    `json:"tick"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e SimulationStepEvent) EventType() string {
	return e.Type
}

func (e SimulationStepEvent) EventSimulationID() string {
	return e.SimulationID
}
