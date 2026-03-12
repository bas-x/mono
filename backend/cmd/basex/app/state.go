package app

import (
	"fmt"
	"time"

	"github.com/bas-x/basex/services"
)

type Status string

const (
	StatusIdle    Status = "idle"
	StatusReady   Status = "ready"
	StatusRunning Status = "running"
	StatusPaused  Status = "paused"
	StatusError   Status = "error"
)

type State struct {
	Status            Status              `json:"status"`
	SimulationID      string              `json:"simulationId,omitempty"`
	Running           bool                `json:"running"`
	Paused            bool                `json:"paused"`
	Tick              uint64              `json:"tick"`
	CurrentTime       time.Time           `json:"currentTime"`
	Airbases          []services.Airbase  `json:"airbases"`
	Aircraft          []services.Aircraft `json:"aircraft"`
	Threats           []services.Threat   `json:"threats"`
	ThreatActivity    []string            `json:"threatActivity,omitempty"`
	SelectedKind      string              `json:"selectedKind,omitempty"`
	SelectedAirbaseID string              `json:"selectedAirbaseId,omitempty"`
	SelectedAircraft  string              `json:"selectedAircraftTail,omitempty"`
	LastEventType     string              `json:"lastEventType,omitempty"`
	EventLog          []string            `json:"eventLog,omitempty"`
	RecentEvents      []string            `json:"recentEvents,omitempty"`
	Error             string              `json:"error,omitempty"`
}

func NewState() *State {
	return &State{Status: StatusIdle, Airbases: []services.Airbase{}, Aircraft: []services.Aircraft{}, Threats: []services.Threat{}, ThreatActivity: []string{}, RecentEvents: []string{}, EventLog: []string{}}
}

func (s *State) SetSnapshot(info services.SimulationInfo, airbases []services.Airbase, aircraft []services.Aircraft, threats []services.Threat) {
	s.SimulationID = info.ID
	s.Running = info.Running
	s.Paused = info.Paused
	s.Airbases = append([]services.Airbase(nil), airbases...)
	s.Aircraft = append([]services.Aircraft(nil), aircraft...)
	s.Threats = append([]services.Threat(nil), threats...)
	s.Tick = info.Tick
	s.CurrentTime = info.Timestamp
	s.Error = ""
	if info.Paused {
		s.Status = StatusPaused
	} else if info.Running {
		s.Status = StatusRunning
	} else {
		s.Status = StatusReady
	}
	if s.SelectedAirbaseID == "" && len(s.Airbases) > 0 {
		s.SelectedAirbaseID = s.Airbases[0].ID
		if s.SelectedKind == "" {
			s.SelectedKind = "airbase"
		}
	}
	if s.SelectedAircraft == "" && len(s.Aircraft) > 0 {
		s.SelectedAircraft = s.Aircraft[0].TailNumber
		if s.SelectedKind == "" {
			s.SelectedKind = "aircraft"
		}
	}
	s.addEvent("snapshot loaded", true)
}

func (s *State) SetError(err error) {
	if err == nil {
		return
	}
	s.Status = StatusError
	s.Error = err.Error()
	s.addEvent("error: "+err.Error(), true)
}

func (s *State) Clear() {
	*s = *NewState()
}

func (s *State) ApplyEvent(event services.Event) {
	if event == nil {
		return
	}
	s.LastEventType = event.EventType()
	if simID := event.EventSimulationID(); simID != "" {
		s.SimulationID = simID
	}
	s.Running = true
	s.Paused = false
	s.Status = StatusRunning
	s.Error = ""
	switch e := event.(type) {
	case services.SimulationStepEvent:
		s.addEvent(fmt.Sprintf("tick %d", e.Tick), false)
		s.Tick = e.Tick
		s.CurrentTime = e.Timestamp
	case services.AircraftStateChangeEvent:
		s.addEvent(eventSummary(event), true)
		for i := range s.Aircraft {
			if s.Aircraft[i].TailNumber == e.TailNumber {
				s.Aircraft[i] = e.Aircraft
				return
			}
		}
		s.Aircraft = append(s.Aircraft, e.Aircraft)
	case services.LandingAssignmentEvent:
		s.addEvent(eventSummary(event), true)
		for i := range s.Aircraft {
			if s.Aircraft[i].TailNumber == e.TailNumber {
				s.Aircraft[i].AssignedTo = &e.BaseID
				return
			}
		}
	case services.ThreatSpawnedEvent:
		s.addEvent("threat spawned in "+e.Threat.Region, true)
		s.addThreatActivity("spawned " + e.Threat.Region)
		s.Threats = append(s.Threats, e.Threat)
	case services.ThreatClaimedEvent:
		s.addEvent("threat claimed in "+e.Threat.Region, true)
		s.addThreatActivity("claimed " + e.Threat.Region + " by " + shortTail(e.TailNumber))
		filtered := s.Threats[:0]
		for _, threat := range s.Threats {
			if threat.ID != e.Threat.ID {
				filtered = append(filtered, threat)
			}
		}
		s.Threats = append([]services.Threat(nil), filtered...)
	default:
		s.addEvent(eventSummary(event), true)
	}
}

func (s *State) addThreatActivity(message string) {
	if message == "" {
		return
	}
	s.ThreatActivity = append([]string{message}, s.ThreatActivity...)
	if len(s.ThreatActivity) > 8 {
		s.ThreatActivity = s.ThreatActivity[:8]
	}
}

func (s *State) SelectAirbaseIndex(index int) {
	if index < 0 || index >= len(s.Airbases) {
		return
	}
	s.SelectedKind = "airbase"
	s.SelectedAirbaseID = s.Airbases[index].ID
}

func shortTail(tail string) string {
	if len(tail) <= 8 {
		return tail
	}
	return tail[:8]
}

func (s *State) SelectAircraftIndex(index int) {
	if index < 0 || index >= len(s.Aircraft) {
		return
	}
	s.SelectedKind = "aircraft"
	s.SelectedAircraft = s.Aircraft[index].TailNumber
}

func (s *State) addEvent(message string, recent bool) {
	if message == "" {
		return
	}
	s.EventLog = append(s.EventLog, message)
	if len(s.EventLog) > 512 {
		s.EventLog = s.EventLog[len(s.EventLog)-512:]
	}
	if !recent {
		return
	}
	s.RecentEvents = append([]string{message}, s.RecentEvents...)
	if len(s.RecentEvents) > 12 {
		s.RecentEvents = s.RecentEvents[:12]
	}
}

func eventSummary(event services.Event) string {
	switch e := event.(type) {
	case services.SimulationStepEvent:
		return "tick advanced"
	case services.AircraftStateChangeEvent:
		return e.TailNumber + " " + e.OldState + "→" + e.NewState
	case services.LandingAssignmentEvent:
		return e.TailNumber + " assigned " + e.BaseID
	default:
		return event.EventType()
	}
}
