package services

import (
	"context"
	"encoding/hex"
	errors "errors"
	"sync"
	"time"

	"github.com/bas-x/basex/simulation"
)

var (
	ErrBaseAlreadyExists  = errors.New("simulation service: base already exists")
	ErrBaseNotFound       = errors.New("simulation service: base not found")
	ErrSimulationRunning  = errors.New("simulation service: simulation already running")
	ErrSimulationNotFound = errors.New("simulation service: simulation not found")
)

const BaseSimulationID = "base"

type SimulationServiceConfig struct {
	Resolution   time.Duration
	RunnerConfig simulation.ControlledRunnerConfig
}

type SimulationService struct {
	mu          sync.RWMutex
	base        *managedSimulation
	broadcaster *EventBroadcaster
	resolution  time.Duration
	runnerCfg   simulation.ControlledRunnerConfig
}

type managedSimulation struct {
	sim     *simulation.Simulation
	runner  *simulation.ControlledRunner
	cancel  context.CancelFunc
	running bool
}

type BaseSimulationConfig struct {
	Seed    [32]byte
	Options *simulation.SimulationOptions
}

type SimulationInfo struct {
	ID      string `json:"id"`
	Running bool   `json:"running"`
}

func NewSimulationService(cfg SimulationServiceConfig) *SimulationService {
	if cfg.Resolution <= 0 {
		cfg.Resolution = 500 * time.Millisecond
	}
	if cfg.RunnerConfig.TicksPerSecond == 0 {
		cfg.RunnerConfig.TicksPerSecond = 64
	}
	if cfg.RunnerConfig.MaxCatchUpTicks == 0 {
		cfg.RunnerConfig.MaxCatchUpTicks = 5
	}
	return &SimulationService{
		broadcaster: NewEventBroadcaster(0),
		resolution:  cfg.Resolution,
		runnerCfg:   cfg.RunnerConfig,
	}
}

// CreateBaseSimulation instantiates the base simulation using the configured defaults.
func (s *SimulationService) CreateBaseSimulation(cfg BaseSimulationConfig) (*simulation.Simulation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.base != nil {
		return nil, ErrBaseAlreadyExists
	}
	ts := simulation.New(s.resolution, simulation.WithEpoch(time.Now()))
	sim := simulation.NewSimulator(cfg.Seed, ts)
	s.registerHooks(BaseSimulationID, sim)
	options := cfg.Options
	if options == nil {
		options = &simulation.SimulationOptions{}
	}
	if err := sim.Init(options); err != nil {
		return nil, err
	}
	s.base = &managedSimulation{sim: sim}
	return sim, nil
}

func (s *SimulationService) StartSimulation(simulationID string) error {
	s.mu.Lock()
	managed, err := s.managedSimulationByIDLocked(simulationID)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	if managed.running {
		s.mu.Unlock()
		return ErrSimulationRunning
	}

	runner := simulation.NewControlledRunner(s.runnerCfg)
	ctx, cancel := context.WithCancel(context.Background())
	managed.runner = runner
	managed.cancel = cancel
	managed.running = true
	sim := managed.sim
	s.mu.Unlock()

	go func() {
		runner.Run(ctx, sim)
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.base != managed {
			return
		}
		managed.runner = nil
		managed.cancel = nil
		managed.running = false
	}()

	return nil
}

func (s *SimulationService) Broadcaster() *EventBroadcaster {
	return s.broadcaster
}

func (s *SimulationService) Base() (*simulation.Simulation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.base == nil {
		return nil, false
	}
	return s.base.sim, true
}

func (s *SimulationService) Airbases(simulationID string) ([]Airbase, error) {
	sim, err := s.simulationByID(simulationID)
	if err != nil {
		return nil, err
	}

	raw := sim.Airbases()
	airbases := make([]Airbase, len(raw))
	for i, airbase := range raw {
		airbases[i] = mapAirbase(airbase)
	}

	return airbases, nil
}

func (s *SimulationService) Aircrafts(simulationID string) ([]Aircraft, error) {
	sim, err := s.simulationByID(simulationID)
	if err != nil {
		return nil, err
	}

	raw := sim.Aircrafts()
	aircrafts := make([]Aircraft, len(raw))
	for i, aircraft := range raw {
		aircrafts[i] = mapAircraft(aircraft)
	}

	return aircrafts, nil
}

func (s *SimulationService) Simulations() []SimulationInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]SimulationInfo, 0, 1)
	if s.base != nil {
		result = append(result, SimulationInfo{ID: BaseSimulationID, Running: s.base.running})
	}

	return result
}

func (s *SimulationService) simulationByID(simulationID string) (*simulation.Simulation, error) {
	managed, err := s.managedSimulationByID(simulationID)
	if err != nil {
		return nil, err
	}
	return managed.sim, nil
}

func (s *SimulationService) managedSimulationByID(simulationID string) (*managedSimulation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.managedSimulationByIDLocked(simulationID)
}

func (s *SimulationService) managedSimulationByIDLocked(simulationID string) (*managedSimulation, error) {

	if simulationID != BaseSimulationID {
		return nil, ErrSimulationNotFound
	}
	if s.base == nil {
		return nil, ErrBaseNotFound
	}

	return s.base, nil
}

func (s *SimulationService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.base != nil && s.base.cancel != nil {
		s.base.cancel()
	}
	s.base = nil
}

func (s *SimulationService) registerHooks(simulationID string, sim *simulation.Simulation) {
	sim.AddAircraftStateChangeHook(func(event simulation.AircraftStateChangeEvent) {
		s.broadcaster.Emit(AircraftStateChangeEvent{
			Type:         EventTypeAircraftStateChange,
			SimulationID: simulationID,
			TailNumber:   hex.EncodeToString(event.TailNumber[:]),
			OldState:     event.OldState,
			NewState:     event.NewState,
			Aircraft:     mapAircraft(event.Aircraft),
			Timestamp:    event.Timestamp,
		})
	})

	sim.AddLandingAssignmentHook(func(event simulation.LandingAssignmentEvent) {
		s.broadcaster.Emit(LandingAssignmentEvent{
			Type:         EventTypeLandingAssignment,
			SimulationID: simulationID,
			TailNumber:   hex.EncodeToString(event.TailNumber[:]),
			BaseID:       hex.EncodeToString(event.Base[:]),
			Source:       mapAssignmentSource(event.Source),
			Timestamp:    event.Timestamp,
		})
	})

	sim.AddSimulationStepHook(func(event simulation.SimulationStepEvent) {
		s.broadcaster.Emit(SimulationStepEvent{
			Type:         EventTypeSimulationStep,
			SimulationID: simulationID,
			Tick:         event.Tick,
			Timestamp:    event.Timestamp,
		})
	})
}

func mapAssignmentSource(source simulation.LandingAssignmentSource) string {
	switch source {
	case simulation.AssignmentSourceAlgorithm:
		return "algorithm"
	case simulation.AssignmentSourceHuman:
		return "human"
	default:
		return "unknown"
	}
}
