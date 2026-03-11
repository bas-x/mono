package api

import (
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var simulationEventsUpgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

func GetSimulationEventsWS(logger *log.Logger, deps *ServerDependencies) echo.HandlerFunc {
	type request struct {
		SimulationID string `param:"simulationId"`
	}

	return func(c echo.Context) error {
		req, err := bindAndValidate[request](c)
		if err != nil {
			return err
		}

		conn, err := simulationEventsUpgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			logger.Error("upgrade simulation websocket", "simulationId", req.SimulationID, "err", err)
			return err
		}

		clientID, events := deps.SimulationService.Broadcaster().Subscribe()
		defer deps.SimulationService.Broadcaster().Unsubscribe(clientID)
		defer conn.Close()

		readDone := make(chan struct{})
		go func() {
			defer close(readDone)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		for {
			select {
			case <-c.Request().Context().Done():
				return nil
			case <-readDone:
				return nil
			case event, ok := <-events:
				if !ok {
					return nil
				}
				if event.EventSimulationID() != req.SimulationID {
					continue
				}
				if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
					return nil
				}
				if err := conn.WriteJSON(event); err != nil {
					return nil
				}
			}
		}
	}
}
