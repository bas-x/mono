package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/prng"
	"github.com/bas-x/basex/services"
	"github.com/bas-x/basex/simulation"
)

func TestSimulationEventsWebSocketEndToEnd(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws/simulations/base/events"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	_, err = deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: websocketSafeOptions()})
	require.NoError(t, err)

	base, ok := deps.SimulationService.Base()
	require.True(t, ok)

	go func() {
		for range 12 {
			base.Step()
		}
	}()

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
	for {
		var payload map[string]any
		err := conn.ReadJSON(&payload)
		require.NoError(t, err)
		if payload["type"] != services.EventTypeAircraftStateChange {
			continue
		}
		require.Equal(t, services.BaseSimulationID, payload["simulationId"])
		require.Equal(t, "Outbound", payload["oldState"])
		require.Equal(t, "Engaged", payload["newState"])
		aircraft, ok := payload["aircraft"].(map[string]any)
		require.True(t, ok)
		require.NotEmpty(t, aircraft["tailNumber"])
		break
	}
}

func TestStartSimulationAndListSimulationsEndpoints(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	deps.SimulationService = services.NewSimulationService(services.SimulationServiceConfig{
		RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 128},
	})
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	resp, err := http.Get(httpServer.URL + "/simulations")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var initial struct {
		Simulations []services.SimulationInfo `json:"simulations"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&initial))
	require.Empty(t, initial.Simulations)

	_, err = deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: websocketSafeOptions()})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/start", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/start", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusConflict, resp.StatusCode)

	resp, err = http.Get(httpServer.URL + "/simulations")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listed struct {
		Simulations []services.SimulationInfo `json:"simulations"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&listed))
	require.Len(t, listed.Simulations, 1)
	require.Equal(t, services.BaseSimulationID, listed.Simulations[0].ID)
	require.True(t, listed.Simulations[0].Running)
}

func websocketSafeOptions() *simulation.SimulationOptions {
	return &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      1,
			MaxPerRegion:      1,
			MaxTotal:          1,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{
			AircraftMin: 1,
			AircraftMax: 1,
			NeedsMin:    1,
			NeedsMax:    2,
		},
	}
}
