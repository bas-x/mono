package api

import (
	"errors"
	"math/rand/v2"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/bas-x/basex/prng"
	"github.com/bas-x/basex/services"
	"github.com/bas-x/basex/simulation"
)

func PostCreateBaseSimulation(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type request struct {
		Seed string `json:"seed"`
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

		cfg := services.BaseSimulationConfig{Seed: seed, Options: defaultBaseSimulationOptions()}
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
		SimulationID string `param:"simulationId"`
	}
	type response struct {
		ID string `json:"id"`
	}

	return func(c echo.Context) error {
		req, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}

		branchID, err := deps.SimulationService.BranchSimulation(req.SimulationID)
		if err != nil {
			if errors.Is(err, services.ErrBaseNotFound) || errors.Is(err, services.ErrSimulationNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "simulation not found")
			}
			logger.Error("branch simulation", "simulationId", req.SimulationID, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to branch simulation")
		}

		return c.JSON(http.StatusCreated, response{ID: branchID})
	}
}

func PostStartSimulation(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type request struct {
		SimulationID string `param:"simulationId"`
	}

	return func(c echo.Context) error {
		req, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}

		err = deps.SimulationService.StartSimulation(req.SimulationID)
		if err != nil {
			if errors.Is(err, services.ErrBaseNotFound) || errors.Is(err, services.ErrSimulationNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "simulation not found")
			}
			if errors.Is(err, services.ErrSimulationRunning) {
				return echo.NewHTTPError(http.StatusConflict, "simulation already running")
			}
			logger.Error("start simulation", "simulationId", req.SimulationID, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to start simulation")
		}

		return c.NoContent(http.StatusAccepted)
	}
}

func PostPauseSimulation(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type request struct {
		SimulationID string `param:"simulationId"`
	}

	return func(c echo.Context) error {
		req, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}

		err = deps.SimulationService.PauseSimulation(req.SimulationID)
		if err != nil {
			if errors.Is(err, services.ErrBaseNotFound) || errors.Is(err, services.ErrSimulationNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "simulation not found")
			}
			if errors.Is(err, services.ErrSimulationNotRunning) || errors.Is(err, services.ErrSimulationPaused) {
				return echo.NewHTTPError(http.StatusConflict, err.Error())
			}
			logger.Error("pause simulation", "simulationId", req.SimulationID, "err", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to pause simulation")
		}

		return c.NoContent(http.StatusAccepted)
	}
}

func PostResumeSimulation(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type request struct {
		SimulationID string `param:"simulationId"`
	}

	return func(c echo.Context) error {
		req, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}

		err = deps.SimulationService.ResumeSimulation(req.SimulationID)
		if err != nil {
			if errors.Is(err, services.ErrBaseNotFound) || errors.Is(err, services.ErrSimulationNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "simulation not found")
			}
			if errors.Is(err, services.ErrSimulationNotRunning) || errors.Is(err, services.ErrSimulationNotPaused) {
				return echo.NewHTTPError(http.StatusConflict, err.Error())
			}
			logger.Error("resume simulation", "simulationId", req.SimulationID, "err", err)
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
