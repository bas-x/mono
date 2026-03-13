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
	ErrBaseAlreadyExists    = errors.New("simulation service: base already exists")
	ErrBaseNotFound         = errors.New("simulation service: base not found")
	ErrSimulationRunning    = errors.New("simulation service: simulation already running")
	ErrSimulationNotRunning = errors.New("simulation service: simulation not running")
	ErrSimulationPaused     = errors.New("simulation service: simulation already paused")
	ErrSimulationNotPaused  = errors.New("simulation service: simulation not paused")
	ErrSimulationNotFound   = errors.New("simulation service: simulation not found")
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
	paused  bool
}

type BaseSimulationConfig struct {
	Seed    [32]byte
	Options *simulation.SimulationOptions
}

type SimulationInfo struct {
	ID        string    `json:"id"`
	Running   bool      `json:"running"`
	Paused    bool      `json:"paused"`
	Tick      uint64    `json:"tick"`
	Timestamp time.Time `json:"timestamp"`
}

func NewSimulationService(cfg SimulationServiceConfig) *SimulationService {
	if cfg.Resolution <= 0 {
		cfg.Resolution = 5 * time.Second
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
	managed.paused = false
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
		managed.paused = false
	}()

	return nil
}

func (s *SimulationService) PauseSimulation(simulationID string) error {
	s.mu.Lock()
	managed, err := s.managedSimulationByIDLocked(simulationID)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	if !managed.running || managed.runner == nil {
		s.mu.Unlock()
		return ErrSimulationNotRunning
	}
	if managed.paused {
		s.mu.Unlock()
		return ErrSimulationPaused
	}
	runner := managed.runner
	managed.paused = true
	s.mu.Unlock()

	runner.Pause()
	return nil
}

func (s *SimulationService) ResumeSimulation(simulationID string) error {
	s.mu.Lock()
	managed, err := s.managedSimulationByIDLocked(simulationID)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	if !managed.running || managed.runner == nil {
		s.mu.Unlock()
		return ErrSimulationNotRunning
	}
	if !managed.paused {
		s.mu.Unlock()
		return ErrSimulationNotPaused
	}
	runner := managed.runner
	managed.paused = false
	s.mu.Unlock()

	runner.Unpause()
	return nil
}

func (s *SimulationService) ResetSimulation(simulationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	managed, err := s.managedSimulationByIDLocked(simulationID)
	if err != nil {
		return err
	}
	if managed.cancel != nil {
		managed.cancel()
	}
	if s.base == managed {
		s.base = nil
	}
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

func (s *SimulationService) Threats(simulationID string) ([]Threat, error) {
	sim, err := s.simulationByID(simulationID)
	if err != nil {
		return nil, err
	}

	raw := sim.Threats()
	threats := make([]Threat, len(raw))
	for i, threat := range raw {
		threats[i] = mapThreat(threat)
	}
	return threats, nil
}

func (s *SimulationService) Simulations() []SimulationInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]SimulationInfo, 0, 1)
	if s.base != nil {
		result = append(result, simulationInfoFromManaged(BaseSimulationID, s.base))
	}

	return result
}

func (s *SimulationService) Simulation(simulationID string) (SimulationInfo, error) {
	managed, err := s.managedSimulationByID(simulationID)
	if err != nil {
		return SimulationInfo{}, err
	}
	return simulationInfoFromManaged(simulationID, managed), nil
}

func simulationInfoFromManaged(id string, managed *managedSimulation) SimulationInfo {
	info := SimulationInfo{ID: id}
	if managed == nil {
		return info
	}
	info.Running = managed.running
	info.Paused = managed.paused
	if managed.sim != nil {
		info.Tick = managed.sim.Tick()
		info.Timestamp = managed.sim.Now()
	}
	return info
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
	_ = s.ResetSimulation(BaseSimulationID)
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

	sim.AddThreatSpawnedHook(func(event simulation.ThreatSpawnedEvent) {
		s.broadcaster.Emit(ThreatSpawnedEvent{
			Type:         EventTypeThreatSpawned,
			SimulationID: simulationID,
			Threat:       mapThreat(event.Threat),
			Timestamp:    event.Timestamp,
		})
	})

	sim.AddThreatTargetedHook(func(event simulation.ThreatTargetedEvent) {
		s.broadcaster.Emit(ThreatTargetedEvent{
			Type:         EventTypeThreatTargeted,
			SimulationID: simulationID,
			Threat:       mapThreat(event.Threat),
			TailNumber:   hex.EncodeToString(event.TailNumber[:]),
			Timestamp:    event.Timestamp,
		})
	})

	sim.AddThreatDespawnedHook(func(event simulation.ThreatDespawnedEvent) {
		s.broadcaster.Emit(ThreatDespawnedEvent{
			Type:         EventTypeThreatDespawned,
			SimulationID: simulationID,
			Threat:       mapThreat(event.Threat),
			Timestamp:    event.Timestamp,
		})
	})

	sim.AddAllAircraftPositionsHook(func(event simulation.AllAircraftPositionsEvent) {
		snapshots := make([]AircraftPositionSnapshot, len(event.Positions))
		for i, snap := range event.Positions {
			needs := make([]Need, len(snap.Needs))
			for j, need := range snap.Needs {
				needs[j] = Need{
					Type:               string(need.Type),
					Severity:           need.Severity,
					RequiredCapability: string(need.RequiredCapability),
					Blocking:           need.Blocking,
				}
			}
			snapshots[i] = AircraftPositionSnapshot{
				TailNumber: hex.EncodeToString(snap.TailNumber[:]),
				Position:   Point{X: snap.Position.X, Y: snap.Position.Y},
				State:      snap.State,
				Needs:      needs,
			}
		}
		s.broadcaster.Emit(AllAircraftPositionsEvent{
			Type:         EventTypeAllAircraftPositions,
			SimulationID: simulationID,
			Tick:         event.Tick,
			Timestamp:    event.Timestamp,
			Positions:    snapshots,
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
