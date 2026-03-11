package services

import (
	errors "errors"
	"sync"
	"time"

	"github.com/bas-x/basex/simulation"
)

var (
	ErrBaseAlreadyExists  = errors.New("simulation service: base already exists")
	ErrBaseNotFound       = errors.New("simulation service: base not found")
	ErrSimulationNotFound = errors.New("simulation service: simulation not found")
)

const BaseSimulationID = "base"

type SimulationServiceConfig struct {
	Resolution   time.Duration
	RunnerConfig simulation.ControlledRunnerConfig
}

type SimulationService struct {
	mu         sync.RWMutex
	base       *simulation.Simulation
	resolution time.Duration
	runnerCfg  simulation.ControlledRunnerConfig
}

type BaseSimulationConfig struct {
	Seed    [32]byte
	Options *simulation.SimulationOptions
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
		resolution: cfg.Resolution,
		runnerCfg:  cfg.RunnerConfig,
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
	options := cfg.Options
	if options == nil {
		options = &simulation.SimulationOptions{}
	}
	if err := sim.Init(options); err != nil {
		return nil, err
	}
	s.base = sim
	return sim, nil
}

func (s *SimulationService) Base() (*simulation.Simulation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.base == nil {
		return nil, false
	}
	return s.base, true
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

func (s *SimulationService) simulationByID(simulationID string) (*simulation.Simulation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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
	s.base = nil
}
