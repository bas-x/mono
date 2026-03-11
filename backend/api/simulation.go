package api

import (
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

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

		cfg := services.BaseSimulationConfig{Seed: seed, Options: &simulation.SimulationOptions{}}
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

func parseSeed(raw string) ([32]byte, error) {
	var out [32]byte
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return out, nil
	}
	decoded, err := hex.DecodeString(trimmed)
	if err != nil {
		return out, err
	}
	if len(decoded) != len(out) {
		return out, errors.New("seed must be 64 hex characters")
	}
	copy(out[:], decoded)
	return out, nil
}
