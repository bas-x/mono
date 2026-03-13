package api

import (
	"bytes"
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
	deps.SimulationService = services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
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
		oldState, ok := payload["oldState"].(string)
		require.True(t, ok)
		newState, ok := payload["newState"].(string)
		require.True(t, ok)
		if !(oldState == "Ready" && newState == "Outbound") && !(oldState == "Outbound" && newState == "Engaged") {
			continue
		}
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

	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/start", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/start", nil)
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

	req, err = http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/reset", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/reset", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	resp, err = http.Get(httpServer.URL + "/simulations")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&listed))
	require.Empty(t, listed.Simulations)
}

func TestBranchSimulationEndpointAndBranchReads(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	resp, err := http.Post(httpServer.URL+"/simulations/base", "application/json", bytes.NewBufferString(`{"seed":"demo-seed"}`))
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/branch", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()
	require.NotEmpty(t, created.ID)
	require.NotEqual(t, services.BaseSimulationID, created.ID)

	resp, err = http.Get(httpServer.URL + "/simulations/" + created.ID)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var branchInfo services.SimulationInfo
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&branchInfo))
	require.Equal(t, created.ID, branchInfo.ID)

	resp, err = http.Get(httpServer.URL + "/simulations/" + created.ID + "/aircrafts")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var aircraftPayload struct {
		Aircrafts []services.Aircraft `json:"aircrafts"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&aircraftPayload))
	require.NotEmpty(t, aircraftPayload.Aircrafts)

	req, err = http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/"+created.ID+"/branch", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetSimulationEndpoint(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	resp, err := http.Get(httpServer.URL + "/simulations/base")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	_, err = deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: websocketSafeOptions()})
	require.NoError(t, err)

	resp, err = http.Get(httpServer.URL + "/simulations/base")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var payload services.SimulationInfo
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	require.Equal(t, services.BaseSimulationID, payload.ID)
	require.False(t, payload.Running)
	require.False(t, payload.Paused)
	require.Zero(t, payload.Tick)
	require.False(t, payload.Timestamp.IsZero())

	resp, err = http.Get(httpServer.URL + "/simulations/branch-a")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPauseResumeAndThreatEndpoints(t *testing.T) {
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

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      1,
			MaxPerRegion:      1,
			MaxTotal:          1,
			RegionProbability: prng.New(1, 1),
		},
		ThreatOpts: simulation.ThreatOptions{SpawnChance: prng.New(1, 1), MaxActive: 2},
	}})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/start", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/pause", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	resp, err = http.Get(httpServer.URL + "/simulations/base")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var info services.SimulationInfo
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&info))
	require.True(t, info.Running)
	require.True(t, info.Paused)

	req, err = http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/resume", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/resume", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusConflict, resp.StatusCode)

	base, ok := deps.SimulationService.Base()
	require.True(t, ok)
	base.Step()

	resp, err = http.Get(httpServer.URL + "/simulations/base/threats")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var threatsPayload struct {
		Threats []services.Threat `json:"threats"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&threatsPayload))
	require.NotEmpty(t, threatsPayload.Threats)
	require.NotZero(t, threatsPayload.Threats[0].Position.X+threatsPayload.Threats[0].Position.Y)
}

func TestGlobalLifecycleEndpointsApplyToBaseAndBranch(t *testing.T) {
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

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: websocketSafeOptions()})
	require.NoError(t, err)

	branchID, err := deps.SimulationService.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/start", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	require.Eventually(t, func() bool {
		baseInfo, err := deps.SimulationService.Simulation(services.BaseSimulationID)
		if err != nil {
			return false
		}
		branchInfo, err := deps.SimulationService.Simulation(branchID)
		if err != nil {
			return false
		}
		return baseInfo.Running && !baseInfo.Paused && branchInfo.Running && !branchInfo.Paused
	}, time.Second, 10*time.Millisecond)

	req, err = http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/pause", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	baseInfo, err := deps.SimulationService.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	branchInfo, err := deps.SimulationService.Simulation(branchID)
	require.NoError(t, err)
	require.True(t, baseInfo.Running)
	require.True(t, baseInfo.Paused)
	require.True(t, branchInfo.Running)
	require.True(t, branchInfo.Paused)

	req, err = http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/resume", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	require.Eventually(t, func() bool {
		baseInfo, err := deps.SimulationService.Simulation(services.BaseSimulationID)
		if err != nil {
			return false
		}
		branchInfo, err := deps.SimulationService.Simulation(branchID)
		if err != nil {
			return false
		}
		return baseInfo.Running && !baseInfo.Paused && branchInfo.Running && !branchInfo.Paused
	}, time.Second, 10*time.Millisecond)
}

func TestGlobalLifecycleEndpointsReturnNotFoundWithoutBaseSimulation(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	for _, path := range []string{"/simulations/start", "/simulations/pause", "/simulations/resume"} {
		req, err := http.NewRequest(http.MethodPost, httpServer.URL+path, nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusNotFound, resp.StatusCode, path)
	}
}

func TestBranchSimulationEventsWebSocketFiltersToBranch(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	deps.SimulationService = services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: websocketSafeOptions()})
	require.NoError(t, err)

	branchID, err := deps.SimulationService.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws/simulations/" + branchID + "/events"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, deps.SimulationService.StepSimulation(services.BaseSimulationID))
	require.NoError(t, deps.SimulationService.StepSimulation(branchID))

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
	for {
		var payload map[string]any
		err := conn.ReadJSON(&payload)
		require.NoError(t, err)
		require.Equal(t, branchID, payload["simulationId"])
		if payload["type"] == services.EventTypeSimulationStep {
			break
		}
	}
}

func TestCreateBaseSimulationProvidesGeneratedAircrafts(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	resp, err := http.Post(httpServer.URL+"/simulations/base", "application/json", bytes.NewBufferString(`{"seed":"demo-seed"}`))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	resp, err = http.Get(httpServer.URL + "/simulations/base/aircrafts")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var payload struct {
		Aircrafts []services.Aircraft `json:"aircrafts"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	require.NotEmpty(t, payload.Aircrafts)
	for _, aircraft := range payload.Aircrafts {
		require.NotEmpty(t, aircraft.TailNumber)
		require.NotEmpty(t, aircraft.State)
		require.NotEmpty(t, aircraft.Needs)
	}
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
			NeedsMin:    0,
			NeedsMax:    0,
		},
		ThreatOpts: simulation.ThreatOptions{
			SpawnChance: prng.New(1, 1),
			MaxActive:   1,
		},
	}
}
