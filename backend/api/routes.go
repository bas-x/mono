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
}

func bindAndValidate[T any](c echo.Context) (*T, error) {
	var req T
	err := c.Bind(&req)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = c.Validate(req)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err)
	}

	return &req, nil
}
