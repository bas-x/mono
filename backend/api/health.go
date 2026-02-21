package api

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
)

func GetHealth(logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		logger.Debug("health check")
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
}

func GetPing(logger *log.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		logger.Debug("ping")
		return c.String(http.StatusOK, "pong")
	}
}
