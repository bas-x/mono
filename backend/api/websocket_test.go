package api

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand/v2"
	"net"
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

	var created services.SimulationInfo
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()
	require.NotEmpty(t, created.ID)
	require.NotEqual(t, services.BaseSimulationID, created.ID)
	require.NotNil(t, created.ParentID)
	require.Equal(t, services.BaseSimulationID, *created.ParentID)
	require.NotNil(t, created.SplitTick)
	require.NotNil(t, created.SplitTimestamp)

	resp, err = http.Get(httpServer.URL + "/simulations/" + created.ID)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var branchInfo services.SimulationInfo
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&branchInfo))
	require.Equal(t, created.ID, branchInfo.ID)
	require.NotNil(t, branchInfo.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchInfo.ParentID)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, *created.SplitTick, *branchInfo.SplitTick)
	require.Equal(t, *created.SplitTimestamp, *branchInfo.SplitTimestamp)

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

func TestListSimulationsSourceEvent(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: websocketSafeOptions()})
	require.NoError(t, err)

	legacyBranchID, err := deps.SimulationService.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	branchBody := bytes.NewBufferString(`{"sourceEvent":{"id":"evt-list","type":"landing_assignment","tick":7}}`)
	branchReq, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/branch", branchBody)
	require.NoError(t, err)
	branchReq.Header.Set("Content-Type", "application/json")

	branchResp, err := http.DefaultClient.Do(branchReq)
	require.NoError(t, err)
	defer branchResp.Body.Close()
	require.Equal(t, http.StatusCreated, branchResp.StatusCode)

	var created services.SimulationInfo
	require.NoError(t, json.NewDecoder(branchResp.Body).Decode(&created))
	require.NotEmpty(t, created.ID)
	require.NotNil(t, created.SourceEvent)
	require.Equal(t, "evt-list", created.SourceEvent.ID)
	require.Equal(t, "landing_assignment", created.SourceEvent.Type)
	require.Equal(t, uint64(7), created.SourceEvent.Tick)

	resp, err := http.Get(httpServer.URL + "/simulations")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var payload struct {
		Simulations []map[string]any `json:"simulations"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))

	byID := make(map[string]map[string]any, len(payload.Simulations))
	for _, item := range payload.Simulations {
		id, ok := item["id"].(string)
		require.True(t, ok)
		byID[id] = item
	}

	legacyPayload, ok := byID[legacyBranchID]
	require.True(t, ok)
	_, hasLegacySourceEvent := legacyPayload["sourceEvent"]
	require.False(t, hasLegacySourceEvent)

	newPayload, ok := byID[created.ID]
	require.True(t, ok)
	sourceEventRaw, hasNewSourceEvent := newPayload["sourceEvent"]
	require.True(t, hasNewSourceEvent)
	sourceEvent, ok := sourceEventRaw.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "evt-list", sourceEvent["id"])
	require.Equal(t, "landing_assignment", sourceEvent["type"])
	require.Equal(t, float64(7), sourceEvent["tick"])
}

func TestBranchSimulationEndpointAcceptsSourceEventRequestBody(t *testing.T) {
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

	body := bytes.NewBufferString(`{"sourceEvent":{"id":"evt-123","type":"landing_assignment","tick":7}}`)
	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/branch", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var payload map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	require.NotEmpty(t, payload["id"])
	require.Equal(t, services.BaseSimulationID, payload["parentId"])
	require.NotNil(t, payload["splitTick"])
	require.NotNil(t, payload["splitTimestamp"])

	sourceEventRaw, ok := payload["sourceEvent"]
	require.True(t, ok)
	sourceEvent, ok := sourceEventRaw.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "evt-123", sourceEvent["id"])
	require.Equal(t, "landing_assignment", sourceEvent["type"])
	require.Equal(t, float64(7), sourceEvent["tick"])
}

func TestBranchSimulationEndpointSourceEventAbsentRemainsBackwardCompatible(t *testing.T) {
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
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var payload map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	require.NotEmpty(t, payload["id"])
	require.Equal(t, services.BaseSimulationID, payload["parentId"])
	_, hasSourceEvent := payload["sourceEvent"]
	require.False(t, hasSourceEvent)
}

func TestBranchSimulationEndpointRejectsMalformedSourceEventRequests(t *testing.T) {
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

	invalidBodies := []string{
		`{"sourceEvent":{}}`,
		`{"sourceEvent":{"id":"evt-only"}}`,
		`{"sourceEvent":{"type":"landing_assignment"}}`,
		`{"sourceEvent":{"tick":9}}`,
		`{"sourceEvent":{"id":"evt-123","type":"landing_assignment"}}`,
		`{"sourceEvent":{"id":"evt-123","tick":9}}`,
		`{"sourceEvent":{"type":"landing_assignment","tick":9}}`,
		`{"sourceEvent":{"id":"","type":"landing_assignment","tick":9}}`,
		`{"sourceEvent":{"id":"evt-123","type":"","tick":9}}`,
	}

	for _, body := range invalidBodies {
		body := body
		t.Run(body, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/branch", bytes.NewBufferString(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			resp.Body.Close()
			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
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
	require.Nil(t, payload.ParentID)
	require.Nil(t, payload.SplitTick)
	require.Nil(t, payload.SplitTimestamp)

	branchID, err := deps.SimulationService.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	resp, err = http.Get(httpServer.URL + "/simulations/" + branchID)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var branchPayload services.SimulationInfo
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&branchPayload))
	require.Equal(t, branchID, branchPayload.ID)
	require.NotNil(t, branchPayload.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchPayload.ParentID)
	require.NotNil(t, branchPayload.SplitTick)
	require.NotNil(t, branchPayload.SplitTimestamp)

	resp, err = http.Get(httpServer.URL + "/simulations/branch-a")
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetSimulationSourceEvent(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: websocketSafeOptions()})
	require.NoError(t, err)

	legacyBranchID, err := deps.SimulationService.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	branchBody := bytes.NewBufferString(`{"sourceEvent":{"id":"evt-detail","type":"landing_assignment","tick":11}}`)
	branchReq, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/branch", branchBody)
	require.NoError(t, err)
	branchReq.Header.Set("Content-Type", "application/json")

	branchResp, err := http.DefaultClient.Do(branchReq)
	require.NoError(t, err)
	defer branchResp.Body.Close()
	require.Equal(t, http.StatusCreated, branchResp.StatusCode)

	var created services.SimulationInfo
	require.NoError(t, json.NewDecoder(branchResp.Body).Decode(&created))
	require.NotEmpty(t, created.ID)

	resp, err := http.Get(httpServer.URL + "/simulations/" + created.ID)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var detailPayload map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&detailPayload))

	sourceEventRaw, hasSourceEvent := detailPayload["sourceEvent"]
	require.True(t, hasSourceEvent)
	sourceEvent, ok := sourceEventRaw.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "evt-detail", sourceEvent["id"])
	require.Equal(t, "landing_assignment", sourceEvent["type"])
	require.Equal(t, float64(11), sourceEvent["tick"])

	legacyResp, err := http.Get(httpServer.URL + "/simulations/" + legacyBranchID)
	require.NoError(t, err)
	defer legacyResp.Body.Close()
	require.Equal(t, http.StatusOK, legacyResp.StatusCode)

	var legacyPayload map[string]any
	require.NoError(t, json.NewDecoder(legacyResp.Body).Decode(&legacyPayload))
	_, hasLegacySourceEvent := legacyPayload["sourceEvent"]
	require.False(t, hasLegacySourceEvent)
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

func TestBranchCreatedEventOverWebSocket(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: websocketSafeOptions()})
	require.NoError(t, err)

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws/simulations/base/events"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	branchID, err := deps.SimulationService.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	branchInfo, err := deps.SimulationService.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfo.ParentID)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
	for {
		var payload map[string]any
		err := conn.ReadJSON(&payload)
		require.NoError(t, err)
		if payload["type"] != services.EventTypeBranchCreated {
			continue
		}

		require.Equal(t, services.EventTypeBranchCreated, payload["type"])
		require.Equal(t, services.BaseSimulationID, payload["simulationId"])
		require.Equal(t, branchID, payload["branchId"])
		require.Equal(t, *branchInfo.ParentID, payload["parentId"])
		require.Equal(t, float64(*branchInfo.SplitTick), payload["splitTick"])

		splitTimestampRaw, ok := payload["splitTimestamp"].(string)
		require.True(t, ok)
		parsedSplitTimestamp, err := time.Parse(time.RFC3339Nano, splitTimestampRaw)
		require.NoError(t, err)
		require.Equal(t, *branchInfo.SplitTimestamp, parsedSplitTimestamp)
		break
	}
}

func TestBranchCreatedEventSourceEventOverWebSocket(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: websocketSafeOptions()})
	require.NoError(t, err)

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws/simulations/base/events"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	legacyBranchID, err := deps.SimulationService.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	branchBody := bytes.NewBufferString(`{"sourceEvent":{"id":"evt-ws","type":"landing_assignment","tick":13}}`)
	branchReq, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/branch", branchBody)
	require.NoError(t, err)
	branchReq.Header.Set("Content-Type", "application/json")

	branchResp, err := http.DefaultClient.Do(branchReq)
	require.NoError(t, err)
	defer branchResp.Body.Close()
	require.Equal(t, http.StatusCreated, branchResp.StatusCode)

	var created services.SimulationInfo
	require.NoError(t, json.NewDecoder(branchResp.Body).Decode(&created))
	require.NotEmpty(t, created.ID)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
	branchCreatedEvents := make(map[string]map[string]any, 2)
	for len(branchCreatedEvents) < 2 {
		var payload map[string]any
		err := conn.ReadJSON(&payload)
		require.NoError(t, err)
		if payload["type"] != services.EventTypeBranchCreated {
			continue
		}
		branchID, ok := payload["branchId"].(string)
		require.True(t, ok)
		branchCreatedEvents[branchID] = payload
	}

	legacyPayload, ok := branchCreatedEvents[legacyBranchID]
	require.True(t, ok)
	_, hasLegacySourceEvent := legacyPayload["sourceEvent"]
	require.False(t, hasLegacySourceEvent)

	newPayload, ok := branchCreatedEvents[created.ID]
	require.True(t, ok)
	sourceEventRaw, hasNewSourceEvent := newPayload["sourceEvent"]
	require.True(t, hasNewSourceEvent)
	sourceEvent, ok := sourceEventRaw.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "evt-ws", sourceEvent["id"])
	require.Equal(t, "landing_assignment", sourceEvent["type"])
	require.Equal(t, float64(13), sourceEvent["tick"])
}

func TestBranchCreatedEventIsDeliveredOnlyToBaseStream(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: websocketSafeOptions()})
	require.NoError(t, err)

	branchStreamID, err := deps.SimulationService.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	baseWSURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws/simulations/base/events"
	baseConn, _, err := websocket.DefaultDialer.Dial(baseWSURL, nil)
	require.NoError(t, err)
	defer baseConn.Close()

	branchWSURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws/simulations/" + branchStreamID + "/events"
	branchConn, _, err := websocket.DefaultDialer.Dial(branchWSURL, nil)
	require.NoError(t, err)
	defer branchConn.Close()

	newBranchID, err := deps.SimulationService.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	newBranchInfo, err := deps.SimulationService.Simulation(newBranchID)
	require.NoError(t, err)
	require.NotNil(t, newBranchInfo.ParentID)
	require.NotNil(t, newBranchInfo.SplitTick)
	require.NotNil(t, newBranchInfo.SplitTimestamp)

	require.NoError(t, baseConn.SetReadDeadline(time.Now().Add(2*time.Second)))
	for {
		var payload map[string]any
		err := baseConn.ReadJSON(&payload)
		require.NoError(t, err)
		if payload["type"] != services.EventTypeBranchCreated {
			continue
		}

		require.Equal(t, services.EventTypeBranchCreated, payload["type"])
		require.Equal(t, services.BaseSimulationID, payload["simulationId"])
		require.Equal(t, newBranchID, payload["branchId"])
		require.Equal(t, *newBranchInfo.ParentID, payload["parentId"])
		require.Equal(t, float64(*newBranchInfo.SplitTick), payload["splitTick"])

		splitTimestampRaw, ok := payload["splitTimestamp"].(string)
		require.True(t, ok)
		parsedSplitTimestamp, err := time.Parse(time.RFC3339Nano, splitTimestampRaw)
		require.NoError(t, err)
		require.Equal(t, *newBranchInfo.SplitTimestamp, parsedSplitTimestamp)
		break
	}

	require.NoError(t, branchConn.SetReadDeadline(time.Now().Add(300*time.Millisecond)))
	var branchPayload map[string]any
	err = branchConn.ReadJSON(&branchPayload)
	require.Error(t, err)
	netErr, ok := err.(net.Error)
	require.True(t, ok)
	require.True(t, netErr.Timeout())
}

func TestSimulationEndedEventOverWebSocket(t *testing.T) {
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

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: websocketSafeOptions(), UntilTick: 3})
	require.NoError(t, err)

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws/simulations/base/events"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/start", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
	for {
		var payload map[string]any
		err := conn.ReadJSON(&payload)
		require.NoError(t, err)
		if payload["type"] != services.EventTypeSimulationEnded {
			continue
		}
		require.Equal(t, services.BaseSimulationID, payload["simulationId"])
		require.Equal(t, float64(3), payload["tick"])

		summaryRaw, ok := payload["summary"]
		require.True(t, ok)
		summary, ok := summaryRaw.(map[string]any)
		require.True(t, ok)
		require.Equal(t, float64(0), summary["completedVisitCount"])
		require.Equal(t, float64(0), summary["totalDurationMs"])
		require.Nil(t, summary["averageDurationMs"])
		_, hasNestedServicing := summary["servicing"]
		require.False(t, hasNestedServicing)
		break
	}
}

func TestSimulationClosedEventOverWebSocket(t *testing.T) {
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

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws/simulations/base/events"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/start", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
	for {
		var payload map[string]any
		err := conn.ReadJSON(&payload)
		require.NoError(t, err)
		if payload["type"] == services.EventTypeSimulationStep {
			break
		}
	}

	req, err = http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/reset", nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
	for {
		var payload map[string]any
		err := conn.ReadJSON(&payload)
		require.NoError(t, err)
		if payload["type"] != services.EventTypeSimulationClosed {
			continue
		}

		require.Equal(t, services.BaseSimulationID, payload["simulationId"])
		require.Equal(t, services.SimulationClosedReasonReset, payload["reason"])

		summaryRaw, ok := payload["summary"]
		require.True(t, ok)
		summary, ok := summaryRaw.(map[string]any)
		require.True(t, ok)
		require.Equal(t, float64(0), summary["completedVisitCount"])
		require.Equal(t, float64(0), summary["totalDurationMs"])
		require.Nil(t, summary["averageDurationMs"])
		_, hasNestedServicing := summary["servicing"]
		require.False(t, hasNestedServicing)
		break
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

func TestCreateBaseSimulationAcceptsUntilTickAndSimulationOptions(t *testing.T) {
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

	body := `{
		"seed": "demo-seed",
		"untilTick": 3,
		"simulationOptions": {
			"constellationOpts": {
				"includeRegions": ["Blekinge"],
				"minPerRegion": 1,
				"maxPerRegion": 1,
				"maxTotal": 1,
				"regionProbability": {"numerator": 1, "denominator": 1}
			},
			"fleetOpts": {
				"aircraftMin": 1,
				"aircraftMax": 1,
				"needsMin": 0,
				"needsMax": 0
			},
			"threatOpts": {
				"spawnChance": {"numerator": 1, "denominator": 1},
				"maxActive": 1
			}
		}
	}`

	resp, err := http.Post(httpServer.URL+"/simulations/base", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	startReq, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/start", nil)
	require.NoError(t, err)
	startResp, err := http.DefaultClient.Do(startReq)
	require.NoError(t, err)
	startResp.Body.Close()
	require.Equal(t, http.StatusAccepted, startResp.StatusCode)

	require.Eventually(t, func() bool {
		info, err := deps.SimulationService.Simulation(services.BaseSimulationID)
		if err != nil {
			return false
		}
		return !info.Running && info.Tick == 3
	}, 2*time.Second, 10*time.Millisecond)

	resp, err = http.Get(httpServer.URL + "/simulations/base")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var infoPayload services.SimulationInfo
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&infoPayload))
	require.Equal(t, int64(3), infoPayload.UntilTick)

	resp, err = http.Get(httpServer.URL + "/simulations/base/aircrafts")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var payload struct {
		Aircrafts []services.Aircraft `json:"aircrafts"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	require.Len(t, payload.Aircrafts, 1)
}

func TestCreateBaseSimulationAcceptsLifecycleOptions(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	body := `{
		"simulationOptions": {
			"constellationOpts": {
				"includeRegions": ["Blekinge"],
				"minPerRegion": 1,
				"maxPerRegion": 1,
				"maxTotal": 1,
				"regionProbability": {"numerator": 1, "denominator": 1}
			},
			"fleetOpts": {
				"aircraftMin": 1,
				"aircraftMax": 1,
				"needsMin": 1,
				"needsMax": 1,
				"needsPool": ["fuel"],
				"severityMin": 2,
				"severityMax": 2
			},
			"threatOpts": {
				"spawnChance": {"numerator": 1, "denominator": 1},
				"maxActive": 1
			},
			"lifecycleOpts": {
				"durations": {
					"outbound": 3600000000000,
					"engaged": 1000000000,
					"inboundDecision": 1000000000,
					"commitApproach": 1000000000,
					"servicing": 1000000000,
					"ready": 0
				},
				"returnThreshold": 1,
				"needRates": {
					"fuel": {
						"outboundMilliPerHour": 9000,
						"engagedMilliPerHour": 9000,
						"servicingMilliPerHour": 9000,
						"variancePermille": 0
					}
				}
			}
		}
	}`

	resp, err := http.Post(httpServer.URL+"/simulations/base", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	base, ok := deps.SimulationService.Base()
	require.True(t, ok)
	require.NotNil(t, base)

	for i := 0; i < 3; i++ {
		base.Step()
	}
	aircrafts := base.Aircrafts()
	require.Len(t, aircrafts, 1)
	require.Equal(t, "Inbound", aircrafts[0].State.Name())
}

func TestCreateBaseSimulationRejectsInvalidLifecycleNeedRateKey(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	body := `{
		"simulationOptions": {
			"lifecycleOpts": {
				"needRates": {
					"bad-key": {
						"outboundMilliPerHour": 1,
						"engagedMilliPerHour": 1,
						"servicingMilliPerHour": 1,
						"variancePermille": 1
					}
				}
			}
		}
	}`

	resp, err := http.Post(httpServer.URL+"/simulations/base", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestOverrideAssignmentEndpointReturnsAircraftAndAssignment(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      2,
			MaxPerRegion:      2,
			MaxTotal:          2,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{AircraftMin: 1, AircraftMax: 1, NeedsMin: 0, NeedsMax: 0},
	}})
	require.NoError(t, err)

	bases, err := deps.SimulationService.Airbases(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, bases, 2)

	aircrafts, err := deps.SimulationService.Aircrafts(services.BaseSimulationID)
	require.NoError(t, err)
	require.NotEmpty(t, aircrafts)
	tailNumber := aircrafts[0].TailNumber
	targetBase := bases[1].ID

	body := bytes.NewBufferString(`{"baseId":"` + targetBase + `"}`)
	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/aircraft/"+tailNumber+"/assignment-override", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var payload struct {
		Aircraft struct {
			TailNumber string  `json:"tailNumber"`
			AssignedTo *string `json:"assignedTo"`
		} `json:"aircraft"`
		Assignment struct {
			Base   string `json:"base"`
			Source string `json:"source"`
		} `json:"assignment"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	require.Equal(t, tailNumber, payload.Aircraft.TailNumber)
	require.NotNil(t, payload.Aircraft.AssignedTo)
	require.Equal(t, targetBase, *payload.Aircraft.AssignedTo)
	require.Equal(t, targetBase, payload.Assignment.Base)
	require.Equal(t, "human", payload.Assignment.Source)
}

func TestOverrideAssignmentEndpointRejectsLateOverride(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      2,
			MaxPerRegion:      2,
			MaxTotal:          2,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{
			AircraftMin: 1,
			AircraftMax: 1,
			NeedsMin:    0,
			NeedsMax:    0,
			StateFactory: func(_ *rand.Rand) simulation.AircraftState {
				return &simulation.InboundState{}
			},
		},
		LifecycleOpts: &simulation.LifecycleModel{
			Durations: simulation.PhaseDurations{
				Outbound:        5 * time.Second,
				Engaged:         5 * time.Second,
				InboundDecision: 3 * time.Second,
				CommitApproach:  4 * time.Second,
				Servicing:       6 * time.Second,
				Ready:           2 * time.Second,
			},
			ReturnThreshold: 80,
			NeedRates: map[simulation.NeedType]simulation.NeedRateModel{
				simulation.NeedFuel:      {OutboundMilliPerHour: 18000000, EngagedMilliPerHour: 18000000, ServicingMilliPerHour: 28800000, VariancePermille: 0},
				simulation.NeedMunitions: {OutboundMilliPerHour: 18000000, EngagedMilliPerHour: 18000000, ServicingMilliPerHour: 18000000, VariancePermille: 0},
			},
		},
	}})
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		require.NoError(t, deps.SimulationService.StepSimulation(services.BaseSimulationID))
	}

	bases, err := deps.SimulationService.Airbases(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, bases, 2)

	aircrafts, err := deps.SimulationService.Aircrafts(services.BaseSimulationID)
	require.NoError(t, err)
	require.NotEmpty(t, aircrafts)
	tailNumber := aircrafts[0].TailNumber
	targetBase := bases[1].ID

	body := bytes.NewBufferString(`{"baseId":"` + targetBase + `"}`)
	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/aircraft/"+tailNumber+"/assignment-override", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestAssignmentOverrideEventOverWebSocket(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      2,
			MaxPerRegion:      2,
			MaxTotal:          2,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{AircraftMin: 1, AircraftMax: 1, NeedsMin: 0, NeedsMax: 0},
	}})
	require.NoError(t, err)

	bases, err := deps.SimulationService.Airbases(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, bases, 2)

	aircrafts, err := deps.SimulationService.Aircrafts(services.BaseSimulationID)
	require.NoError(t, err)
	require.NotEmpty(t, aircrafts)
	tailNumber := aircrafts[0].TailNumber
	targetBase := bases[1].ID

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws/simulations/base/events"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	body := bytes.NewBufferString(`{"baseId":"` + targetBase + `"}`)
	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/aircraft/"+tailNumber+"/assignment-override", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
	for {
		var payload map[string]any
		err := conn.ReadJSON(&payload)
		require.NoError(t, err)
		if payload["type"] != services.EventTypeLandingAssignment {
			continue
		}
		require.Equal(t, services.BaseSimulationID, payload["simulationId"])
		tail, ok := payload["tailNumber"].(string)
		require.True(t, ok)
		if tail != tailNumber {
			continue
		}
		source, ok := payload["source"].(string)
		require.True(t, ok)
		if source != "human" {
			continue
		}
		require.Equal(t, targetBase, payload["baseId"])
		require.Equal(t, "human", source)
		break
	}
}

func TestOverrideAssignmentEndpointRejectsInvalidTailNumber(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      2,
			MaxPerRegion:      2,
			MaxTotal:          2,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{AircraftMin: 1, AircraftMax: 1, NeedsMin: 0, NeedsMax: 0},
	}})
	require.NoError(t, err)

	bases, err := deps.SimulationService.Airbases(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, bases, 2)

	body := bytes.NewBufferString(`{"baseId":"` + bases[1].ID + `"}`)
	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/aircraft/not-hex/assignment-override", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestOverrideAssignmentEndpointRejectsInvalidBaseID(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      2,
			MaxPerRegion:      2,
			MaxTotal:          2,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{AircraftMin: 1, AircraftMax: 1, NeedsMin: 0, NeedsMax: 0},
	}})
	require.NoError(t, err)

	aircrafts, err := deps.SimulationService.Aircrafts(services.BaseSimulationID)
	require.NoError(t, err)
	require.NotEmpty(t, aircrafts)

	body := bytes.NewBufferString(`{"baseId":"not-hex"}`)
	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/aircraft/"+aircrafts[0].TailNumber+"/assignment-override", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestOverrideAssignmentEndpointRejectsUnknownAircraft(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      2,
			MaxPerRegion:      2,
			MaxTotal:          2,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{AircraftMin: 1, AircraftMax: 1, NeedsMin: 0, NeedsMax: 0},
	}})
	require.NoError(t, err)

	bases, err := deps.SimulationService.Airbases(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, bases, 2)

	body := bytes.NewBufferString(`{"baseId":"` + bases[1].ID + `"}`)
	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/aircraft/ffffffffffffffff/assignment-override", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestOverrideAssignmentEndpointRejectsUnknownBase(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      2,
			MaxPerRegion:      2,
			MaxTotal:          2,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{AircraftMin: 1, AircraftMax: 1, NeedsMin: 0, NeedsMax: 0},
	}})
	require.NoError(t, err)

	aircrafts, err := deps.SimulationService.Aircrafts(services.BaseSimulationID)
	require.NoError(t, err)
	require.NotEmpty(t, aircrafts)

	body := bytes.NewBufferString(`{"baseId":"ffffffffffffffff"}`)
	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/aircraft/"+aircrafts[0].TailNumber+"/assignment-override", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestOverrideAssignmentEndpointRejectsUnknownSimulation(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	body := bytes.NewBufferString(`{"baseId":"ffffffffffffffff"}`)
	req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/missing/aircraft/ffffffffffffffff/assignment-override", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAssignmentOverrideEventOverWebSocketEmitsHumanEventForEachRepeatedOverride(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	config := viper.New()
	deps := initDeps(config)
	server := newServer(logger, config, deps)
	httpServer := httptest.NewServer(server.Handler)
	defer httpServer.Close()

	_, err := deps.SimulationService.CreateBaseSimulation(services.BaseSimulationConfig{Options: &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      2,
			MaxPerRegion:      2,
			MaxTotal:          2,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{AircraftMin: 1, AircraftMax: 1, NeedsMin: 0, NeedsMax: 0},
	}})
	require.NoError(t, err)

	bases, err := deps.SimulationService.Airbases(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, bases, 2)

	aircrafts, err := deps.SimulationService.Aircrafts(services.BaseSimulationID)
	require.NoError(t, err)
	require.NotEmpty(t, aircrafts)
	tailNumber := aircrafts[0].TailNumber

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws/simulations/base/events"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	postOverride := func(baseID string) {
		body := bytes.NewBufferString(`{"baseId":"` + baseID + `"}`)
		req, err := http.NewRequest(http.MethodPost, httpServer.URL+"/simulations/base/aircraft/"+tailNumber+"/assignment-override", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}

	postOverride(bases[1].ID)
	postOverride(bases[0].ID)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
	humanEvents := make([]string, 0, 2)
	for len(humanEvents) < 2 {
		var payload map[string]any
		err := conn.ReadJSON(&payload)
		require.NoError(t, err)
		if payload["type"] != services.EventTypeLandingAssignment {
			continue
		}
		tail, ok := payload["tailNumber"].(string)
		require.True(t, ok)
		if tail != tailNumber {
			continue
		}
		source, ok := payload["source"].(string)
		require.True(t, ok)
		if source != "human" {
			continue
		}
		baseID, ok := payload["baseId"].(string)
		require.True(t, ok)
		humanEvents = append(humanEvents, baseID)
	}

	require.Equal(t, []string{bases[1].ID, bases[0].ID}, humanEvents)
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
