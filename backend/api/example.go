package api

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
)

func PostExample(logger *log.Logger, e *struct{}) echo.HandlerFunc {
	type request struct {
		Param string `param:"param_a"`
		Query string `query:"query_a"`
		BodyA string `json:"body_a"`
		BodyB string `json:"body_b"`
	}
	return func(c echo.Context) error {
		_, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	}
}
