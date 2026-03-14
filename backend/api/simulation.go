package api

import (
	"errors"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/bas-x/basex/prng"
	"github.com/bas-x/basex/services"
	"github.com/bas-x/basex/simulation"
)

type ratioRequest struct {
	Numerator   uint64 `json:"numerator"`
	Denominator uint64 `json:"denominator"`
}

type constellationOptionsRequest struct {
	IncludeRegions        []string      `json:"includeRegions"`
	ExcludeRegions        []string      `json:"excludeRegions"`
	MinPerRegion          uint          `json:"minPerRegion"`
	MaxPerRegion          uint          `json:"maxPerRegion"`
	MaxTotal              uint          `json:"maxTotal"`
	RegionProbability     *ratioRequest `json:"regionProbability"`
	MaxAttemptsPerAirbase uint          `json:"maxAttemptsPerAirbase"`
}

type fleetOptionsRequest struct {
	AircraftMin    uint          `json:"aircraftMin"`
	AircraftMax    uint          `json:"aircraftMax"`
	NeedsMin       uint          `json:"needsMin"`
	NeedsMax       uint          `json:"needsMax"`
	NeedsPool      []string      `json:"needsPool"`
	SeverityMin    uint          `json:"severityMin"`
	SeverityMax    uint          `json:"severityMax"`
	BlockingChance *ratioRequest `json:"blockingChance"`
}

type threatOptionsRequest struct {
	SpawnChance    *ratioRequest `json:"spawnChance"`
	MaxActive      uint          `json:"maxActive"`
	MaxActiveTicks uint64        `json:"maxActiveTicks"`
}

type phaseDurationsRequest struct {
	Outbound        time.Duration `json:"outbound"`
	Engaged         time.Duration `json:"engaged"`
	InboundDecision time.Duration `json:"inboundDecision"`
	CommitApproach  time.Duration `json:"commitApproach"`
	Servicing       time.Duration `json:"servicing"`
	Ready           time.Duration `json:"ready"`
}

type needRateModelRequest struct {
	OutboundMilliPerHour  int64 `json:"outboundMilliPerHour"`
	EngagedMilliPerHour   int64 `json:"engagedMilliPerHour"`
	ServicingMilliPerHour int64 `json:"servicingMilliPerHour"`
	VariancePermille      int64 `json:"variancePermille"`
}

type lifecycleOptionsRequest struct {
	Durations       *phaseDurationsRequest          `json:"durations"`
	ReturnThreshold int                             `json:"returnThreshold"`
	NeedRates       map[string]needRateModelRequest `json:"needRates"`
}

type simulationOptionsRequest struct {
	ConstellationOpts *constellationOptionsRequest `json:"constellationOpts"`
	FleetOpts         *fleetOptionsRequest         `json:"fleetOpts"`
	ThreatOpts        *threatOptionsRequest        `json:"threatOpts"`
	LifecycleOpts     *lifecycleOptionsRequest     `json:"lifecycleOpts"`
}

type branchSourceEventRequest struct {
	ID   string  `json:"id"`
	Type string  `json:"type"`
	Tick *uint64 `json:"tick"`
}

func PostCreateBaseSimulation(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type request struct {
		Seed              string                    `json:"seed"`
		UntilTick         int64                     `json:"untilTick"`
		SimulationOptions *simulationOptionsRequest `json:"simulationOptions"`
	}
	type response struct {
		ID string `json:"id"`
	}
	return func(c echo.Context) error {
		req, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}

		seed, err := parseSeed(req.Seed)
		if err != nil {
			logger.Warn("invalid seed", "err", err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		options, err := mapSimulationOptionsRequest(req.SimulationOptions)
		if err != nil {
			logger.Warn("invalid simulation options", "err", err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if options == nil {
			options = defaultBaseSimulationOptions()
		}
		cfg := services.BaseSimulationConfig{
			Seed:      seed,
			UntilTick: req.UntilTick,
			Options:   options,
		}
		_, err = deps.SimulationService.CreateBaseSimulation(cfg)
		if err != nil {
			if errors.Is(err, services.ErrBaseAlreadyExists) {
				return echo.NewHTTPError(http.StatusConflict, "base simulation already exists")
			}
			logger.Error("create simulation core", "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create simulation")
		}

		resp := response{ID: services.BaseSimulationID}
		return c.JSON(http.StatusCreated, resp)
	}
}

func GetSimulations(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type response struct {
		Simulations []services.SimulationInfo `json:"simulations"`
	}

	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, response{Simulations: deps.SimulationService.Simulations()})
	}
}

func GetSimulation(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type request struct {
		SimulationID string `param:"simulationId"`
	}

	return func(c echo.Context) error {
		req, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}

		simulationInfo, err := deps.SimulationService.Simulation(req.SimulationID)
		if err != nil {
			if errors.Is(err, services.ErrBaseNotFound) || errors.Is(err, services.ErrSimulationNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "simulation not found")
			}
			logger.Error("get simulation", "simulationId", req.SimulationID, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get simulation")
		}

		return c.JSON(http.StatusOK, simulationInfo)
	}
}

func PostBranchSimulation(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type request struct {
		SimulationID string                    `param:"simulationId"`
		SourceEvent  *branchSourceEventRequest `json:"sourceEvent"`
	}

	return func(c echo.Context) error {
		req, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}

		sourceEvent, err := mapBranchSourceEventRequest(req.SourceEvent)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		branchID, err := branchSimulation(req.SimulationID, sourceEvent, deps)
		if err != nil {
			if errors.Is(err, services.ErrBaseNotFound) || errors.Is(err, services.ErrSimulationNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "simulation not found")
			}
			logger.Error("branch simulation", "simulationId", req.SimulationID, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to branch simulation")
		}

		simulationInfo, err := deps.SimulationService.Simulation(branchID)
		if err != nil {
			logger.Error("load branched simulation", "branchId", branchID, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to load branched simulation")
		}
		if sourceEvent != nil && simulationInfo.SourceEvent == nil {
			simulationInfo.SourceEvent = sourceEvent
		}

		return c.JSON(http.StatusCreated, simulationInfo)
	}
}

type sourceEventBrancher interface {
	BranchSimulationWithSourceEvent(simulationID string, sourceEvent *services.SourceEvent) (string, error)
}

func mapBranchSourceEventRequest(req *branchSourceEventRequest) (*services.SourceEvent, error) {
	if req == nil {
		return nil, nil
	}
	if strings.TrimSpace(req.ID) == "" {
		return nil, errors.New("sourceEvent.id is required")
	}
	if strings.TrimSpace(req.Type) == "" {
		return nil, errors.New("sourceEvent.type is required")
	}
	if req.Tick == nil {
		return nil, errors.New("sourceEvent.tick is required")
	}

	return &services.SourceEvent{
		ID:   req.ID,
		Type: req.Type,
		Tick: *req.Tick,
	}, nil
}

func branchSimulation(simulationID string, sourceEvent *services.SourceEvent, deps *ServerDependencies) (string, error) {
	if sourceEvent == nil {
		return deps.SimulationService.BranchSimulation(simulationID)
	}
	if brancher, ok := any(deps.SimulationService).(sourceEventBrancher); ok {
		return brancher.BranchSimulationWithSourceEvent(simulationID, sourceEvent)
	}
	return deps.SimulationService.BranchSimulation(simulationID)
}

func PostStartSimulation(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := deps.SimulationService.StartSimulation(services.BaseSimulationID)
		if err != nil {
			if errors.Is(err, services.ErrBaseNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "simulation not found")
			}
			if errors.Is(err, services.ErrSimulationRunning) {
				return echo.NewHTTPError(http.StatusConflict, "simulation already running")
			}
			logger.Error("start simulations", "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to start simulation")
		}

		return c.NoContent(http.StatusAccepted)
	}
}

func PostPauseSimulation(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := deps.SimulationService.PauseSimulation(services.BaseSimulationID)
		if err != nil {
			if errors.Is(err, services.ErrBaseNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "simulation not found")
			}
			if errors.Is(err, services.ErrSimulationNotRunning) || errors.Is(err, services.ErrSimulationPaused) {
				return echo.NewHTTPError(http.StatusConflict, err.Error())
			}
			logger.Error("pause simulations", "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to pause simulation")
		}

		return c.NoContent(http.StatusAccepted)
	}
}

func PostResumeSimulation(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := deps.SimulationService.ResumeSimulation(services.BaseSimulationID)
		if err != nil {
			if errors.Is(err, services.ErrBaseNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "simulation not found")
			}
			if errors.Is(err, services.ErrSimulationNotRunning) || errors.Is(err, services.ErrSimulationNotPaused) {
				return echo.NewHTTPError(http.StatusConflict, err.Error())
			}
			logger.Error("resume simulations", "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to resume simulation")
		}

		return c.NoContent(http.StatusAccepted)
	}
}

func PostResetSimulation(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type request struct {
		SimulationID string `param:"simulationId"`
	}

	return func(c echo.Context) error {
		req, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}

		err = deps.SimulationService.ResetSimulation(req.SimulationID)
		if err != nil {
			if errors.Is(err, services.ErrBaseNotFound) || errors.Is(err, services.ErrSimulationNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "simulation not found")
			}
			logger.Error("reset simulation", "simulationId", req.SimulationID, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to reset simulation")
		}

		return c.NoContent(http.StatusAccepted)
	}
}

func GetSimulationAirbases(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type request struct {
		SimulationID string `param:"simulationId"`
	}
	type response struct {
		Airbases []services.Airbase `json:"airbases"`
	}
	return func(c echo.Context) error {
		req, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}

		airbases, err := deps.SimulationService.Airbases(req.SimulationID)
		if err != nil {
			if errors.Is(err, services.ErrBaseNotFound) || errors.Is(err, services.ErrSimulationNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "simulation not found")
			}
			logger.Error("list simulation airbases", "simulationId", req.SimulationID, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list airbases")
		}

		return c.JSON(http.StatusOK, response{Airbases: airbases})
	}
}

func GetSimulationAircrafts(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type request struct {
		SimulationID string `param:"simulationId"`
	}
	type response struct {
		Aircrafts []services.Aircraft `json:"aircrafts"`
	}
	return func(c echo.Context) error {
		req, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}

		aircrafts, err := deps.SimulationService.Aircrafts(req.SimulationID)
		if err != nil {
			if errors.Is(err, services.ErrBaseNotFound) || errors.Is(err, services.ErrSimulationNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "simulation not found")
			}
			logger.Error("list simulation aircrafts", "simulationId", req.SimulationID, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list aircrafts")
		}

		return c.JSON(http.StatusOK, response{Aircrafts: aircrafts})
	}
}

func GetSimulationThreats(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type request struct {
		SimulationID string `param:"simulationId"`
	}
	type response struct {
		Threats []services.Threat `json:"threats"`
	}

	return func(c echo.Context) error {
		req, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}

		threats, err := deps.SimulationService.Threats(req.SimulationID)
		if err != nil {
			if errors.Is(err, services.ErrBaseNotFound) || errors.Is(err, services.ErrSimulationNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "simulation not found")
			}
			logger.Error("list simulation threats", "simulationId", req.SimulationID, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list threats")
		}

		return c.JSON(http.StatusOK, response{Threats: threats})
	}
}

func PostOverrideAssignment(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type request struct {
		SimulationID string `param:"simulationId"`
		TailNumber   string `param:"tailNumber"`
		BaseID       string `json:"baseId"`
	}
	type response struct {
		Aircraft   services.Aircraft   `json:"aircraft"`
		Assignment services.Assignment `json:"assignment"`
	}

	return func(c echo.Context) error {
		req, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}

		aircraft, assignment, err := deps.SimulationService.OverrideAssignment(req.SimulationID, req.TailNumber, req.BaseID)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrBaseNotFound), errors.Is(err, services.ErrSimulationNotFound), errors.Is(err, services.ErrAircraftNotFound):
				return echo.NewHTTPError(http.StatusNotFound, "simulation or aircraft not found")
			case errors.Is(err, services.ErrInvalidTailNumber), errors.Is(err, services.ErrInvalidBaseID), errors.Is(err, simulation.ErrAirbaseNotFound):
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			case errors.Is(err, services.ErrAssignmentTooLate):
				return echo.NewHTTPError(http.StatusConflict, "assignment override too late")
			default:
				logger.Error("override assignment", "simulationId", req.SimulationID, "tailNumber", req.TailNumber, "baseId", req.BaseID, "err", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to override assignment")
			}
		}

		return c.JSON(http.StatusOK, response{Aircraft: aircraft, Assignment: assignment})
	}
}

func parseSeed(raw string) ([32]byte, error) {
	var out [32]byte
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return out, nil
	}
	if len(trimmed) == 0 {
		return out, errors.New("empty seed")
	}
	copy(out[:], []byte(trimmed))
	return out, nil
}

func mapSimulationOptionsRequest(req *simulationOptionsRequest) (*simulation.SimulationOptions, error) {
	if req == nil {
		return nil, nil
	}
	defaults := defaultBaseSimulationOptions()
	options := &simulation.SimulationOptions{}
	if defaults != nil {
		copied := *defaults
		options = &copied
	}
	if req.ConstellationOpts != nil {
		options.ConstellationOpts = simulation.ConstellationOptions{
			IncludeRegions:        req.ConstellationOpts.IncludeRegions,
			ExcludeRegions:        req.ConstellationOpts.ExcludeRegions,
			MinPerRegion:          req.ConstellationOpts.MinPerRegion,
			MaxPerRegion:          req.ConstellationOpts.MaxPerRegion,
			MaxTotal:              req.ConstellationOpts.MaxTotal,
			MaxAttemptsPerAirbase: req.ConstellationOpts.MaxAttemptsPerAirbase,
		}
		if req.ConstellationOpts.RegionProbability != nil {
			options.ConstellationOpts.RegionProbability = prng.New(req.ConstellationOpts.RegionProbability.Numerator, req.ConstellationOpts.RegionProbability.Denominator)
		}
	}
	if req.FleetOpts != nil {
		needsPool, err := mapNeedTypes(req.FleetOpts.NeedsPool)
		if err != nil {
			return nil, err
		}
		options.FleetOpts = simulation.FleetOptions{
			AircraftMin: req.FleetOpts.AircraftMin,
			AircraftMax: req.FleetOpts.AircraftMax,
			NeedsMin:    req.FleetOpts.NeedsMin,
			NeedsMax:    req.FleetOpts.NeedsMax,
			NeedsPool:   needsPool,
			SeverityMin: req.FleetOpts.SeverityMin,
			SeverityMax: req.FleetOpts.SeverityMax,
		}
		if req.FleetOpts.BlockingChance != nil {
			options.FleetOpts.BlockingChance = prng.New(req.FleetOpts.BlockingChance.Numerator, req.FleetOpts.BlockingChance.Denominator)
		}
	}
	if req.ThreatOpts != nil {
		options.ThreatOpts = simulation.ThreatOptions{
			MaxActive:      req.ThreatOpts.MaxActive,
			MaxActiveTicks: req.ThreatOpts.MaxActiveTicks,
		}
		if req.ThreatOpts.SpawnChance != nil {
			options.ThreatOpts.SpawnChance = prng.New(req.ThreatOpts.SpawnChance.Numerator, req.ThreatOpts.SpawnChance.Denominator)
		}
	}
	if req.LifecycleOpts != nil {
		lifecycle := simulation.DefaultLifecycleModel()
		if req.LifecycleOpts.Durations != nil {
			lifecycle.Durations = simulation.PhaseDurations{
				Outbound:        req.LifecycleOpts.Durations.Outbound,
				Engaged:         req.LifecycleOpts.Durations.Engaged,
				InboundDecision: req.LifecycleOpts.Durations.InboundDecision,
				CommitApproach:  req.LifecycleOpts.Durations.CommitApproach,
				Servicing:       req.LifecycleOpts.Durations.Servicing,
				Ready:           req.LifecycleOpts.Durations.Ready,
			}
		}
		if req.LifecycleOpts.ReturnThreshold != 0 {
			lifecycle.ReturnThreshold = req.LifecycleOpts.ReturnThreshold
		}
		if len(req.LifecycleOpts.NeedRates) > 0 {
			needRates := make(map[simulation.NeedType]simulation.NeedRateModel, len(req.LifecycleOpts.NeedRates))
			for key, value := range req.LifecycleOpts.NeedRates {
				needType, err := mapNeedType(key)
				if err != nil {
					return nil, err
				}
				needRates[needType] = simulation.NeedRateModel{
					OutboundMilliPerHour:  value.OutboundMilliPerHour,
					EngagedMilliPerHour:   value.EngagedMilliPerHour,
					ServicingMilliPerHour: value.ServicingMilliPerHour,
					VariancePermille:      value.VariancePermille,
				}
			}
			lifecycle.NeedRates = needRates
		}
		options.LifecycleOpts = &lifecycle
	}
	return options, nil
}

func mapNeedTypes(values []string) ([]simulation.NeedType, error) {
	if len(values) == 0 {
		return nil, nil
	}
	result := make([]simulation.NeedType, 0, len(values))
	for _, value := range values {
		needType, err := mapNeedType(value)
		if err != nil {
			if strings.TrimSpace(value) == "" {
				continue
			}
			return nil, err
		}
		result = append(result, needType)
	}
	return result, nil
}

func mapNeedType(value string) (simulation.NeedType, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "fuel":
		return simulation.NeedFuel, nil
	case "charge":
		return simulation.NeedCharge, nil
	case "munitions":
		return simulation.NeedMunitions, nil
	case "repairs":
		return simulation.NeedRepairs, nil
	case "maintenance":
		return simulation.NeedMaintenance, nil
	case "missionconfiguration", "mission_configuration", "mission-configuration":
		return simulation.NeedMissionConfiguration, nil
	case "crewsupport", "crew_support", "crew-support":
		return simulation.NeedCrewSupport, nil
	case "emergency":
		return simulation.NeedEmergency, nil
	case "weatherconstraint", "weather_constraint", "weather-constraint":
		return simulation.NeedWeatherConstraint, nil
	case "groundsupport", "ground_support", "ground-support":
		return simulation.NeedGroundSupport, nil
	case "protection":
		return simulation.NeedProtection, nil
	case "":
		return "", errors.New("empty need type")
	default:
		return "", errors.New("invalid need type: " + value)
	}
}

func defaultBaseSimulationOptions() *simulation.SimulationOptions {
	return &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge", "Gotland"},
			MinPerRegion:      1,
			MaxPerRegion:      1,
			MaxTotal:          2,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{
			AircraftMin:  4,
			AircraftMax:  4,
			NeedsMin:     1,
			NeedsMax:     3,
			SeverityMin:  60,
			SeverityMax:  90,
			StateFactory: func(_ *rand.Rand) simulation.AircraftState { return &simulation.ReadyState{} },
		},
		ThreatOpts: simulation.ThreatOptions{
			SpawnChance: prng.New(1, 4),
			MaxActive:   3,
		},
	}
}
