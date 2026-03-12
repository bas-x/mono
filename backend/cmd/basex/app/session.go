package app

import (
	"errors"
	"slices"

	"github.com/bas-x/basex/assets"
	"github.com/bas-x/basex/prng"
	"github.com/bas-x/basex/services"
	"github.com/bas-x/basex/simulation"
)

type Session struct {
	service *services.SimulationService
	subID   uint64
	events  <-chan services.Event
}

func NewSession(service *services.SimulationService) *Session {
	s := &Session{service: service}
	s.resubscribe()
	return s
}

func (s *Session) Close() {
	if s == nil || s.service == nil {
		return
	}
	s.service.Broadcaster().Unsubscribe(s.subID)
}

func (s *Session) Create(seed string) (services.SimulationInfo, []services.Airbase, []services.Aircraft, []services.Threat, error) {
	parsedSeed, err := parseSeed(seed)
	if err != nil {
		return services.SimulationInfo{}, nil, nil, nil, err
	}
	_, err = s.service.CreateBaseSimulation(services.BaseSimulationConfig{
		Seed:    parsedSeed,
		Options: defaultSimulationOptions(),
	})
	if err != nil {
		return services.SimulationInfo{}, nil, nil, nil, err
	}
	return s.Snapshot(services.BaseSimulationID)
}

func (s *Session) Snapshot(simulationID string) (services.SimulationInfo, []services.Airbase, []services.Aircraft, []services.Threat, error) {
	info, err := s.service.Simulation(simulationID)
	if err != nil {
		return services.SimulationInfo{}, nil, nil, nil, err
	}
	airbases, err := s.service.Airbases(simulationID)
	if err != nil {
		return services.SimulationInfo{}, nil, nil, nil, err
	}
	aircrafts, err := s.service.Aircrafts(simulationID)
	if err != nil {
		return services.SimulationInfo{}, nil, nil, nil, err
	}
	threats, err := s.service.Threats(simulationID)
	if err != nil {
		return services.SimulationInfo{}, nil, nil, nil, err
	}
	return info, airbases, aircrafts, threats, nil
}

func (s *Session) Start(simulationID string) error {
	return s.service.StartSimulation(simulationID)
}

func (s *Session) Pause(simulationID string) error {
	return s.service.PauseSimulation(simulationID)
}

func (s *Session) Resume(simulationID string) error {
	return s.service.ResumeSimulation(simulationID)
}

func (s *Session) Reset(simulationID string) error {
	return s.service.ResetSimulation(simulationID)
}

func (s *Session) DrainEvents() []services.Event {
	if s == nil || s.events == nil {
		return nil
	}
	var out []services.Event
	for {
		select {
		case event, ok := <-s.events:
			if !ok {
				s.resubscribe()
				return out
			}
			out = append(out, event)
		default:
			return out
		}
	}
}

func (s *Session) resubscribe() {
	if s == nil || s.service == nil {
		return
	}
	subID, events := s.service.Broadcaster().Subscribe()
	s.subID = subID
	s.events = events
}

func defaultSimulationOptions() *simulation.SimulationOptions {
	return &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    slices.Clone(assets.RegionNames),
			MinPerRegion:      1,
			MaxPerRegion:      3,
			MaxTotal:          15,
			RegionProbability: prng.New(1, 2),
		},
		FleetOpts: simulation.FleetOptions{
			AircraftMin: 6,
			AircraftMax: 40,
			NeedsMin:    1,
			NeedsMax:    uint(len(simulation.AllNeedTypes)),
			SeverityMin: 0,
			SeverityMax: 40,
		},
		ThreatOpts: simulation.ThreatOptions{
			SpawnChance: prng.New(1, 4),
			MaxActive:   3,
		},
	}
}

func mustSimulationID(state *State) (string, error) {
	if state == nil || state.SimulationID == "" {
		return "", errors.New("no active simulation")
	}
	return state.SimulationID, nil
}
