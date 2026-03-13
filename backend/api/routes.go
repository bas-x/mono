package api

import (
	"net/http"

	"github.com/bas-x/basex/assert"
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"

	"github.com/labstack/echo/v4"
)

func registerRoutes(
	e *echo.Echo,
	logger *log.Logger,
	config *viper.Viper,
	deps *ServerDependencies,
) {
	assert.NotNil(e, "echo")
	assert.NotNil(logger, "logger")
	assert.NotNil(deps, "deps")
	assert.NotNil(config, "config")

	e.GET("", func(c echo.Context) error {
		logger.Debug("basex")
		return c.String(http.StatusOK, "basex")
	})

	e.GET("/health", GetHealth(logger))
	e.GET("/ping", GetPing(logger))

	e.GET("/simulations", GetSimulations(logger, deps))
	e.GET("/simulations/:simulationId", GetSimulation(logger, deps))
	e.POST("/simulations/base", PostCreateBaseSimulation(logger, deps))
	e.POST("/simulations/:simulationId/branch", PostBranchSimulation(logger, deps))
	e.POST("/simulations/start", PostStartSimulation(logger, deps))
	e.POST("/simulations/pause", PostPauseSimulation(logger, deps))
	e.POST("/simulations/resume", PostResumeSimulation(logger, deps))
	e.POST("/simulations/:simulationId/reset", PostResetSimulation(logger, deps))
	e.GET("/simulations/:simulationId/airbases", GetSimulationAirbases(logger, deps))
	e.GET("/simulations/:simulationId/aircrafts", GetSimulationAircrafts(logger, deps))
	e.POST("/simulations/:simulationId/aircraft/:tailNumber/assignment-override", PostOverrideAssignment(logger, deps))
	e.GET("/simulations/:simulationId/threats", GetSimulationThreats(logger, deps))
	e.GET("/ws/simulations/:simulationId/events", GetSimulationEventsWS(logger, deps))
}

func bindAndValidate[T any](c echo.Context) (*T, error) {
	var req T
	err := c.Bind(&req)
	if err != nil {
		if err != echo.ErrUnsupportedMediaType {
			return nil, echo.NewHTTPError(http.StatusBadRequest, err)
		}
	}

	err = c.Validate(req)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err)
	}

	return &req, nil
}
