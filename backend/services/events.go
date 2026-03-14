package services

import "time"

const (
	EventTypeAircraftStateChange  = "aircraft_state_change"
	EventTypeBranchCreated        = "branch_created"
	EventTypeLandingAssignment    = "landing_assignment"
	EventTypeSimulationStep       = "simulation_step"
	EventTypeSimulationEnded      = "simulation_ended"
	EventTypeSimulationClosed     = "simulation_closed"
	EventTypeThreatSpawned        = "threat_spawned"
	EventTypeThreatTargeted       = "threat_targeted"
	EventTypeThreatDespawned      = "threat_despawned"
	EventTypeAllAircraftPositions = "all_aircraft_positions"
)

const (
	SimulationClosedReasonReset  = "reset"
	SimulationClosedReasonCancel = "cancel"
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

type BranchCreatedEvent struct {
	Type           string       `json:"type"`
	SimulationID   string       `json:"simulationId"`
	BranchID       string       `json:"branchId"`
	ParentID       string       `json:"parentId"`
	SplitTick      uint64       `json:"splitTick"`
	SplitTimestamp time.Time    `json:"splitTimestamp"`
	SourceEvent    *SourceEvent `json:"sourceEvent,omitempty"`
}

func (e BranchCreatedEvent) EventType() string {
	if e.Type == "" {
		return EventTypeBranchCreated
	}

	return e.Type
}

func (e BranchCreatedEvent) EventSimulationID() string {
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

type SimulationEndedEvent struct {
	Type         string                 `json:"type"`
	SimulationID string                 `json:"simulationId"`
	Tick         uint64                 `json:"tick"`
	Timestamp    time.Time              `json:"timestamp"`
	Summary      ServicingSummaryObject `json:"summary"`
}

type ServicingSummaryObject struct {
	CompletedVisitCount int64  `json:"completedVisitCount"`
	TotalDurationMs     int64  `json:"totalDurationMs"`
	AverageDurationMs   *int64 `json:"averageDurationMs"`
}

func (e SimulationEndedEvent) EventType() string {
	return e.Type
}

func (e SimulationEndedEvent) EventSimulationID() string {
	return e.SimulationID
}

type SimulationClosedEvent struct {
	Type         string                 `json:"type"`
	SimulationID string                 `json:"simulationId"`
	Tick         uint64                 `json:"tick"`
	Timestamp    time.Time              `json:"timestamp"`
	Reason       string                 `json:"reason"`
	Summary      ServicingSummaryObject `json:"summary"`
}

func (e SimulationClosedEvent) EventType() string {
	return e.Type
}

func (e SimulationClosedEvent) EventSimulationID() string {
	return e.SimulationID
}

type ThreatSpawnedEvent struct {
	Type         string    `json:"type"`
	SimulationID string    `json:"simulationId"`
	Threat       Threat    `json:"threat"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e ThreatSpawnedEvent) EventType() string {
	return e.Type
}

func (e ThreatSpawnedEvent) EventSimulationID() string {
	return e.SimulationID
}

type ThreatTargetedEvent struct {
	Type         string    `json:"type"`
	SimulationID string    `json:"simulationId"`
	Threat       Threat    `json:"threat"`
	TailNumber   string    `json:"tailNumber"`
	Timestamp    time.Time `json:"timestamp"`
}

type ThreatDespawnedEvent struct {
	Type         string    `json:"type"`
	SimulationID string    `json:"simulationId"`
	Threat       Threat    `json:"threat"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e ThreatDespawnedEvent) EventType() string {
	return e.Type
}

func (e ThreatDespawnedEvent) EventSimulationID() string {
	return e.SimulationID
}

func (e ThreatTargetedEvent) EventType() string {
	return e.Type
}

func (e ThreatTargetedEvent) EventSimulationID() string {
	return e.SimulationID
}

type AllAircraftPositionsEvent struct {
	Type         string                     `json:"type"`
	SimulationID string                     `json:"simulationId"`
	Tick         uint64                     `json:"tick"`
	Timestamp    time.Time                  `json:"timestamp"`
	Positions    []AircraftPositionSnapshot `json:"positions"`
}

func (e AllAircraftPositionsEvent) EventType() string {
	return e.Type
}

func (e AllAircraftPositionsEvent) EventSimulationID() string {
	return e.SimulationID
}
