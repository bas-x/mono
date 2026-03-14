package services_test

import (
	"encoding/hex"
	"math/rand/v2"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/services"
	"github.com/bas-x/basex/simulation"
)

func TestSimulationServiceEndToEnd_BaseReadModels(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(2, 2))
	require.NoError(t, err)

	airbases, err := svc.Airbases(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, airbases, 2)
	for _, airbase := range airbases {
		require.NotEmpty(t, airbase.RegionID)
		require.NotEmpty(t, airbase.Region)
		require.Len(t, airbase.ID, 16)
		_, decodeErr := hex.DecodeString(airbase.ID)
		require.NoError(t, decodeErr)
	}

	aircrafts, err := svc.Aircrafts(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, aircrafts, 2)
	for _, aircraft := range aircrafts {
		require.Len(t, aircraft.TailNumber, 16)
		_, decodeErr := hex.DecodeString(aircraft.TailNumber)
		require.NoError(t, decodeErr)
		require.NotEmpty(t, aircraft.State)
		require.NotNil(t, aircraft.Needs)
		require.NotEmpty(t, aircraft.Needs)
	}
}

func TestSimulationServiceEndToEnd_LifecycleAndSimulationIDHandling(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()

	_, err := svc.Airbases(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)

	_, err = svc.Aircrafts("branch-a")
	require.ErrorIs(t, err, services.ErrSimulationNotFound)

	_, err = svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	_, err = svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.ErrorIs(t, err, services.ErrBaseAlreadyExists)

	_, err = svc.Airbases("branch-a")
	require.ErrorIs(t, err, services.ErrSimulationNotFound)

	svc.Reset()

	_, err = svc.Aircrafts(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)
}

func TestSimulationServiceEndToEnd_EmitsEvents(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{})
	events := subscribeToServiceEvents(t, svc)
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	step := requireNextSimulationStepEvent(t, events, time.Second, services.BaseSimulationID)
	require.Equal(t, services.BaseSimulationID, step.SimulationID)
}

func TestSimulationServiceEndToEnd_LiveBroadcasterDropsSlowSubscriberWhileFastSubscriberContinues(t *testing.T) {
	t.Parallel()

	const (
		slowBufferCapacity = 16
		dropTriggerTick    = slowBufferCapacity + 1
	)

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	slowSubscriberID, slowEvents := svc.Broadcaster().Subscribe()
	t.Cleanup(func() {
		svc.Broadcaster().Unsubscribe(slowSubscriberID)
	})

	fastEvents := subscribeToServiceEvents(t, svc)

	for expectedTick := uint64(1); expectedTick <= uint64(dropTriggerTick); expectedTick++ {
		require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

		fastStep := requireNextSimulationStepEvent(t, fastEvents, time.Second, services.BaseSimulationID)
		require.Equal(t, services.BaseSimulationID, fastStep.SimulationID)
		require.Equal(t, expectedTick, fastStep.Tick)
	}

	sawSlowScopedStep := false
	slowClosed := false
	closeDeadline := time.Now().Add(time.Second)
	for !slowClosed {
		remaining := time.Until(closeDeadline)
		if remaining <= 0 {
			t.Fatal("timed out waiting for slow subscriber to be dropped")
		}

		select {
		case rawEvent, ok := <-slowEvents:
			if !ok {
				slowClosed = true
				continue
			}

			slowStep, typed := rawEvent.(services.SimulationStepEvent)
			if typed && slowStep.SimulationID == services.BaseSimulationID {
				sawSlowScopedStep = true
			}
		case <-time.After(remaining):
			t.Fatal("timed out waiting for slow subscriber to close")
		}
	}
	require.True(t, sawSlowScopedStep, "slow subscriber should receive scoped simulation.step events before being dropped")

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	fastAfterDrop := requireNextSimulationStepEvent(t, fastEvents, time.Second, services.BaseSimulationID)
	require.Equal(t, services.BaseSimulationID, fastAfterDrop.SimulationID)
	require.Equal(t, uint64(dropTriggerTick+1), fastAfterDrop.Tick)
}

func TestSimulationServiceEndToEnd_StartSimulationAndStatus(t *testing.T) {
	t.Parallel()

	svc := newControlledRunnerService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	list := svc.Simulations()
	listedByID := simulationInfoIndexByID(t, list)
	require.Len(t, listedByID, 1)
	listedBase, ok := listedByID[services.BaseSimulationID]
	require.True(t, ok)
	require.False(t, listedBase.Running)

	events := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))
	require.ErrorIs(t, svc.StartSimulation(services.BaseSimulationID), services.ErrSimulationRunning)

	step := requireNextSimulationStepEvent(t, events, 2*time.Second, services.BaseSimulationID)
	require.Equal(t, services.BaseSimulationID, step.SimulationID)
	require.Greater(t, step.Tick, uint64(0))
	require.True(t, requireSimulationInfo(t, svc, services.BaseSimulationID).Running)
}

func TestSimulationServiceEndToEnd_GlobalLifecycleMatrixUsesAnyValidSimulationID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		startTarget  func(string) string
		startAgain   func(string) string
		pauseTarget  func(string) string
		pauseAgain   func(string) string
		resumeTarget func(string) string
		resumeAgain  func(string) string
	}{
		{
			name:         "start via base pause via branch resume via base",
			startTarget:  func(string) string { return services.BaseSimulationID },
			startAgain:   func(branchID string) string { return branchID },
			pauseTarget:  func(branchID string) string { return branchID },
			pauseAgain:   func(string) string { return services.BaseSimulationID },
			resumeTarget: func(string) string { return services.BaseSimulationID },
			resumeAgain:  func(branchID string) string { return branchID },
		},
		{
			name:         "start via branch pause via base resume via branch",
			startTarget:  func(branchID string) string { return branchID },
			startAgain:   func(string) string { return services.BaseSimulationID },
			pauseTarget:  func(string) string { return services.BaseSimulationID },
			pauseAgain:   func(branchID string) string { return branchID },
			resumeTarget: func(branchID string) string { return branchID },
			resumeAgain:  func(string) string { return services.BaseSimulationID },
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc := newControlledRunnerService()
			_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
			require.NoError(t, err)

			branchID, err := svc.BranchSimulation(services.BaseSimulationID)
			require.NoError(t, err)

			events := subscribeToServiceEvents(t, svc)
			requireLifecycleMatrix(t, svc, branchID,
				lifecycleExpectation{running: false, paused: false},
				lifecycleExpectation{running: false, paused: false},
			)

			require.NoError(t, svc.StartSimulation(tc.startTarget(branchID)))
			requireLifecycleMatrix(t, svc, branchID,
				lifecycleExpectation{running: true, paused: false},
				lifecycleExpectation{running: true, paused: false},
			)

			baseStarted := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, services.BaseSimulationID, 0)
			branchStarted := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, branchID, 0)

			require.ErrorIs(t, svc.StartSimulation(tc.startAgain(branchID)), services.ErrSimulationRunning)
			requireLifecycleMatrix(t, svc, branchID,
				lifecycleExpectation{running: true, paused: false},
				lifecycleExpectation{running: true, paused: false},
			)

			require.NoError(t, svc.PauseSimulation(tc.pauseTarget(branchID)))
			requireLifecycleMatrix(t, svc, branchID,
				lifecycleExpectation{running: true, paused: true},
				lifecycleExpectation{running: true, paused: true},
			)

			require.ErrorIs(t, svc.PauseSimulation(tc.pauseAgain(branchID)), services.ErrSimulationPaused)
			requireLifecycleMatrix(t, svc, branchID,
				lifecycleExpectation{running: true, paused: true},
				lifecycleExpectation{running: true, paused: true},
			)

			basePaused := requireSimulationInfo(t, svc, services.BaseSimulationID)
			branchPaused := requireSimulationInfo(t, svc, branchID)

			require.NoError(t, svc.ResumeSimulation(tc.resumeTarget(branchID)))
			requireLifecycleMatrix(t, svc, branchID,
				lifecycleExpectation{running: true, paused: false},
				lifecycleExpectation{running: true, paused: false},
			)

			baseResumed := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, services.BaseSimulationID, basePaused.Tick)
			branchResumed := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, branchID, branchPaused.Tick)
			require.Greater(t, baseResumed.Tick, baseStarted.Tick)
			require.Greater(t, branchResumed.Tick, branchStarted.Tick)

			require.ErrorIs(t, svc.ResumeSimulation(tc.resumeAgain(branchID)), services.ErrSimulationNotPaused)
			requireLifecycleMatrix(t, svc, branchID,
				lifecycleExpectation{running: true, paused: false},
				lifecycleExpectation{running: true, paused: false},
			)
		})
	}
}

func TestSimulationServiceEndToEnd_LifecycleRequiresValidSimulationIDBeforeGlobalAction(t *testing.T) {
	t.Parallel()

	t.Run("missing base and unknown branch return ID errors", func(t *testing.T) {
		svc := newControlledRunnerService()

		require.ErrorIs(t, svc.StartSimulation(services.BaseSimulationID), services.ErrBaseNotFound)
		require.ErrorIs(t, svc.PauseSimulation(services.BaseSimulationID), services.ErrBaseNotFound)
		require.ErrorIs(t, svc.ResumeSimulation(services.BaseSimulationID), services.ErrBaseNotFound)

		require.ErrorIs(t, svc.StartSimulation("branch-missing"), services.ErrSimulationNotFound)
		require.ErrorIs(t, svc.PauseSimulation("branch-missing"), services.ErrSimulationNotFound)
		require.ErrorIs(t, svc.ResumeSimulation("branch-missing"), services.ErrSimulationNotFound)
	})

	t.Run("unknown branch ID does not mutate idle running or paused matrix", func(t *testing.T) {
		svc := newControlledRunnerService()
		_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
		require.NoError(t, err)

		branchID, err := svc.BranchSimulation(services.BaseSimulationID)
		require.NoError(t, err)

		events := subscribeToServiceEvents(t, svc)
		requireLifecycleMatrix(t, svc, branchID,
			lifecycleExpectation{running: false, paused: false},
			lifecycleExpectation{running: false, paused: false},
		)

		require.ErrorIs(t, svc.StartSimulation("branch-missing"), services.ErrSimulationNotFound)
		requireLifecycleMatrix(t, svc, branchID,
			lifecycleExpectation{running: false, paused: false},
			lifecycleExpectation{running: false, paused: false},
		)
		requireNoSimulationStepEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
		requireNoSimulationStepEvent(t, events, 150*time.Millisecond, branchID)

		require.NoError(t, svc.StartSimulation(branchID))
		requireLifecycleMatrix(t, svc, branchID,
			lifecycleExpectation{running: true, paused: false},
			lifecycleExpectation{running: true, paused: false},
		)

		baseRunning := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, services.BaseSimulationID, 0)
		branchRunning := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, branchID, 0)

		require.ErrorIs(t, svc.PauseSimulation("branch-missing"), services.ErrSimulationNotFound)
		requireLifecycleMatrix(t, svc, branchID,
			lifecycleExpectation{running: true, paused: false},
			lifecycleExpectation{running: true, paused: false},
		)

		baseAfterInvalidPause := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, services.BaseSimulationID, baseRunning.Tick)
		branchAfterInvalidPause := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, branchID, branchRunning.Tick)

		require.NoError(t, svc.PauseSimulation(services.BaseSimulationID))
		requireLifecycleMatrix(t, svc, branchID,
			lifecycleExpectation{running: true, paused: true},
			lifecycleExpectation{running: true, paused: true},
		)

		require.ErrorIs(t, svc.ResumeSimulation("branch-missing"), services.ErrSimulationNotFound)
		requireLifecycleMatrix(t, svc, branchID,
			lifecycleExpectation{running: true, paused: true},
			lifecycleExpectation{running: true, paused: true},
		)

		basePaused := requireSimulationInfo(t, svc, services.BaseSimulationID)
		branchPaused := requireSimulationInfo(t, svc, branchID)

		require.NoError(t, svc.ResumeSimulation(branchID))
		requireLifecycleMatrix(t, svc, branchID,
			lifecycleExpectation{running: true, paused: false},
			lifecycleExpectation{running: true, paused: false},
		)

		baseAfterValidResume := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, services.BaseSimulationID, basePaused.Tick)
		branchAfterValidResume := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, branchID, branchPaused.Tick)
		require.Greater(t, baseAfterValidResume.Tick, baseAfterInvalidPause.Tick)
		require.Greater(t, branchAfterValidResume.Tick, branchAfterInvalidPause.Tick)
	})
}

func TestSimulationServiceEndToEnd_GlobalLifecycleHandlesMixedPausedStatesAcrossBranches(t *testing.T) {
	t.Parallel()

	svc := newControlledRunnerService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	branchOneID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	events := subscribeToServiceEvents(t, svc)

	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))
	requireLifecycleMatrix(t, svc, branchOneID,
		lifecycleExpectation{running: true, paused: false},
		lifecycleExpectation{running: true, paused: false},
	)

	baseStarted := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, services.BaseSimulationID, 0)
	branchOneStarted := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, branchOneID, 0)

	branchTwoID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	requireSimulationLifecycleState(t, svc, services.BaseSimulationID, lifecycleExpectation{running: true, paused: true})
	requireSimulationLifecycleState(t, svc, branchOneID, lifecycleExpectation{running: true, paused: false})
	requireSimulationLifecycleState(t, svc, branchTwoID, lifecycleExpectation{running: true, paused: true})

	require.NoError(t, svc.ResumeSimulation(branchTwoID))
	requireSimulationLifecycleState(t, svc, services.BaseSimulationID, lifecycleExpectation{running: true, paused: false})
	requireSimulationLifecycleState(t, svc, branchOneID, lifecycleExpectation{running: true, paused: false})
	requireSimulationLifecycleState(t, svc, branchTwoID, lifecycleExpectation{running: true, paused: false})

	baseAfterMixedResume := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, services.BaseSimulationID, baseStarted.Tick)
	branchOneAfterMixedResume := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, branchOneID, branchOneStarted.Tick)
	branchTwoAfterMixedResume := requireNextSimulationStepEventAfterTick(t, events, 2*time.Second, branchTwoID, 0)
	require.Greater(t, baseAfterMixedResume.Tick, uint64(0))
	require.Greater(t, branchOneAfterMixedResume.Tick, uint64(0))
	require.Greater(t, branchTwoAfterMixedResume.Tick, uint64(0))

	require.ErrorIs(t, svc.ResumeSimulation(services.BaseSimulationID), services.ErrSimulationNotPaused)
	requireSimulationLifecycleState(t, svc, services.BaseSimulationID, lifecycleExpectation{running: true, paused: false})
	requireSimulationLifecycleState(t, svc, branchOneID, lifecycleExpectation{running: true, paused: false})
	requireSimulationLifecycleState(t, svc, branchTwoID, lifecycleExpectation{running: true, paused: false})

	branchThreeID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	requireSimulationLifecycleState(t, svc, services.BaseSimulationID, lifecycleExpectation{running: true, paused: true})
	requireSimulationLifecycleState(t, svc, branchOneID, lifecycleExpectation{running: true, paused: false})
	requireSimulationLifecycleState(t, svc, branchTwoID, lifecycleExpectation{running: true, paused: false})
	requireSimulationLifecycleState(t, svc, branchThreeID, lifecycleExpectation{running: true, paused: true})

	require.NoError(t, svc.PauseSimulation(branchOneID))
	requireSimulationLifecycleState(t, svc, services.BaseSimulationID, lifecycleExpectation{running: true, paused: true})
	requireSimulationLifecycleState(t, svc, branchOneID, lifecycleExpectation{running: true, paused: true})
	requireSimulationLifecycleState(t, svc, branchTwoID, lifecycleExpectation{running: true, paused: true})
	requireSimulationLifecycleState(t, svc, branchThreeID, lifecycleExpectation{running: true, paused: true})

	require.ErrorIs(t, svc.PauseSimulation(branchThreeID), services.ErrSimulationPaused)
	requireSimulationLifecycleState(t, svc, services.BaseSimulationID, lifecycleExpectation{running: true, paused: true})
	requireSimulationLifecycleState(t, svc, branchOneID, lifecycleExpectation{running: true, paused: true})
	requireSimulationLifecycleState(t, svc, branchTwoID, lifecycleExpectation{running: true, paused: true})
	requireSimulationLifecycleState(t, svc, branchThreeID, lifecycleExpectation{running: true, paused: true})
}

func TestSimulationServiceEndToEnd_EmitsSimulationEndedEvent(t *testing.T) {
	t.Parallel()

	svc := newControlledRunnerService()
	_, err := svc.CreateBaseSimulation(deterministicBaseUntilTickConfig(1, 1, 3))
	require.NoError(t, err)

	events := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))

	ended := requireNextSimulationEndedEvent(t, events, 2*time.Second, services.BaseSimulationID)
	require.Equal(t, services.BaseSimulationID, ended.SimulationID)
	require.Equal(t, uint64(3), ended.Tick)
}

func TestSimulationServiceEndToEnd_SummaryNaturalEnd_ZeroCaseContract(t *testing.T) {
	t.Parallel()

	svc := newControlledRunnerService()
	_, err := svc.CreateBaseSimulation(deterministicBaseUntilTickConfig(1, 1, 3))
	require.NoError(t, err)

	events := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))

	ended := requireNextSimulationEndedEvent(t, events, 2*time.Second, services.BaseSimulationID)
	require.Equal(t, services.EventTypeSimulationEnded, ended.Type)
	require.Equal(t, services.BaseSimulationID, ended.SimulationID)
	require.Equal(t, uint64(3), ended.Tick)
	requireSimulationEndedSummaryContract(t, ended, servicingSummaryExpectation{
		completedVisitCount: 0,
		totalDurationMs:     0,
		averageDurationMs:   nil,
	})
}

func TestSimulationServiceEndToEnd_SummaryNaturalEnd_ContainsExactServicingSummaryFields(t *testing.T) {
	t.Parallel()

	svc := newControlledRunnerService()
	_, err := svc.CreateBaseSimulation(deterministicBaseUntilTickConfig(1, 1, 3))
	require.NoError(t, err)

	events := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))

	ended := requireNextSimulationEndedEvent(t, events, 2*time.Second, services.BaseSimulationID)
	require.Equal(t, services.BaseSimulationID, ended.SimulationID)
	requireSimulationEndedSummaryContract(t, ended, servicingSummaryExpectation{
		completedVisitCount: 0,
		totalDurationMs:     0,
		averageDurationMs:   nil,
	})
}

func TestSimulationServiceEndToEnd_BranchSummary_InheritsCompletedVisitsVisibleAtFork(t *testing.T) {
	t.Parallel()

	svc := newBranchSummaryService()
	_, err := svc.CreateBaseSimulation(branchSummaryBaseConfig())
	require.NoError(t, err)

	events := subscribeToServiceEvents(t, svc)
	baseEventsAtFork := advanceBranchSummaryScenarioToFork(t, svc, events)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	baseInfoAtFork := requireSimulationInfo(t, svc, services.BaseSimulationID)
	branchInfo := requireSimulationInfo(t, svc, branchID)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, baseInfoAtFork.Tick, *branchInfo.SplitTick)
	require.Equal(t, baseInfoAtFork.Timestamp, *branchInfo.SplitTimestamp)

	baseSummaryAtFork := summarizeCompletedServicingVisitsFromEvents(baseEventsAtFork, branchSummaryResolution)
	branchSummaryAtFork := summarizeBranchCompletedServicingVisitsAtSplit(baseEventsAtFork, *branchInfo.SplitTimestamp, nil, branchSummaryResolution)

	requireServicingSummaryExpectation(t, baseSummaryAtFork, servicingSummaryExpectation{
		completedVisitCount: 1,
		totalDurationMs:     5000,
		averageDurationMs:   ptr(int64(5000)),
	})
	requireServicingSummaryExpectation(t, branchSummaryAtFork, servicingSummaryExpectation{
		completedVisitCount: 1,
		totalDurationMs:     5000,
		averageDurationMs:   ptr(int64(5000)),
	})
}

func TestSimulationServiceEndToEnd_BranchSummary_DivergenceKeepsBranchOnlyCompletedVisitsScoped(t *testing.T) {
	t.Parallel()

	svc := newBranchSummaryService()
	_, err := svc.CreateBaseSimulation(branchSummaryBaseConfig())
	require.NoError(t, err)

	events := subscribeToServiceEvents(t, svc)
	baseEventsAtFork := advanceBranchSummaryScenarioToFork(t, svc, events)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	baseInfoAtFork := requireSimulationInfo(t, svc, services.BaseSimulationID)
	branchInfoAtFork := requireSimulationInfo(t, svc, branchID)
	require.NotNil(t, branchInfoAtFork.SplitTimestamp)

	branchEventsAfterFork := advanceBranchSummaryScenarioToBranchOnlyCompletion(t, svc, events, branchID)
	require.Len(t, branchEventsAfterFork, 3)
	branchExitEvent := branchEventsAfterFork[len(branchEventsAfterFork)-1]
	require.Equal(t, "Servicing", branchExitEvent.OldState)
	require.Equal(t, "Ready", branchExitEvent.NewState)
	requireNoAircraftStateChangeEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)

	baseInfoAfterBranchOnlyVisit := requireSimulationInfo(t, svc, services.BaseSimulationID)
	require.Equal(t, baseInfoAtFork.Tick, baseInfoAfterBranchOnlyVisit.Tick)
	require.Equal(t, baseInfoAtFork.Timestamp, baseInfoAfterBranchOnlyVisit.Timestamp)

	baseSummaryAfterBranchDivergence := summarizeCompletedServicingVisitsFromEvents(baseEventsAtFork, branchSummaryResolution)
	branchSummaryAfterBranchDivergence := summarizeBranchCompletedServicingVisitsAtSplit(
		baseEventsAtFork,
		*branchInfoAtFork.SplitTimestamp,
		branchEventsAfterFork,
		branchSummaryResolution,
	)

	requireServicingSummaryExpectation(t, baseSummaryAfterBranchDivergence, servicingSummaryExpectation{
		completedVisitCount: 1,
		totalDurationMs:     5000,
		averageDurationMs:   ptr(int64(5000)),
	})
	requireServicingSummaryExpectation(t, branchSummaryAfterBranchDivergence, servicingSummaryExpectation{
		completedVisitCount: 2,
		totalDurationMs:     10000,
		averageDurationMs:   ptr(int64(5000)),
	})
}

func TestSimulationServiceEndToEnd_BranchInheritsUntilTickAsObservedAbsoluteLimit(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(deterministicBaseUntilTickConfig(1, 1, 5))
	require.NoError(t, err)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	baseInfo := requireSimulationInfo(t, svc, services.BaseSimulationID)
	branchInfo := requireSimulationInfo(t, svc, branchID)
	listedByID := simulationInfoIndexByID(t, svc.Simulations())

	require.Equal(t, uint64(2), baseInfo.Tick)
	require.Equal(t, int64(5), baseInfo.UntilTick)
	require.Equal(t, int64(5), branchInfo.UntilTick)
	require.Equal(t, baseInfo.UntilTick, listedByID[services.BaseSimulationID].UntilTick)
	require.Equal(t, branchInfo.UntilTick, listedByID[branchID].UntilTick)
	require.NotNil(t, branchInfo.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchInfo.ParentID)
	require.NotNil(t, branchInfo.SplitTick)
	require.Equal(t, baseInfo.Tick, *branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, baseInfo.Timestamp, *branchInfo.SplitTimestamp)
}

func TestSimulationServiceEndToEnd_BaseNaturalEndRemainsSeparateFromBranchAfterBranching(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{
		RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 8},
	})
	_, err := svc.CreateBaseSimulation(deterministicBaseUntilTickConfig(1, 1, 5))
	require.NoError(t, err)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

	baseBeforeStart := requireSimulationInfo(t, svc, services.BaseSimulationID)
	branchBeforeStart := requireSimulationInfo(t, svc, branchID)
	require.Equal(t, uint64(4), baseBeforeStart.Tick)
	require.Equal(t, uint64(2), branchBeforeStart.Tick)
	require.Equal(t, int64(5), baseBeforeStart.UntilTick)
	require.Equal(t, int64(5), branchBeforeStart.UntilTick)

	events := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))

	baseEnded := requireNextSimulationEndedEvent(t, events, 2*time.Second, services.BaseSimulationID)
	require.Equal(t, uint64(5), baseEnded.Tick)

	baseAfterEnd := requireSimulationInfo(t, svc, services.BaseSimulationID)
	branchAfterBaseEnd := requireSimulationInfo(t, svc, branchID)
	require.False(t, baseAfterEnd.Running)
	require.False(t, baseAfterEnd.Paused)
	require.Equal(t, uint64(5), baseAfterEnd.Tick)
	require.True(t, branchAfterBaseEnd.Running)
	require.False(t, branchAfterBaseEnd.Paused)
	require.Less(t, branchAfterBaseEnd.Tick, uint64(5))

	requireNoSimulationEndedEvent(t, events, 150*time.Millisecond, branchID)

	branchEnded := requireNextSimulationEndedEvent(t, events, 2*time.Second, branchID)
	require.Equal(t, uint64(5), branchEnded.Tick)

	baseAfterBranchEnd := requireSimulationInfo(t, svc, services.BaseSimulationID)
	branchAfterEnd := requireSimulationInfo(t, svc, branchID)
	require.False(t, baseAfterBranchEnd.Running)
	require.False(t, baseAfterBranchEnd.Paused)
	require.Equal(t, uint64(5), baseAfterBranchEnd.Tick)
	require.False(t, branchAfterEnd.Running)
	require.False(t, branchAfterEnd.Paused)
	require.Equal(t, uint64(5), branchAfterEnd.Tick)
}

func TestSimulationServiceEndToEnd_BranchNaturalEndRemainsSeparateFromBaseAfterBranching(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{
		RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 8},
	})
	_, err := svc.CreateBaseSimulation(deterministicBaseUntilTickConfig(1, 1, 5))
	require.NoError(t, err)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	require.NoError(t, svc.StepSimulation(branchID))
	require.NoError(t, svc.StepSimulation(branchID))

	baseBeforeStart := requireSimulationInfo(t, svc, services.BaseSimulationID)
	branchBeforeStart := requireSimulationInfo(t, svc, branchID)
	require.Equal(t, uint64(2), baseBeforeStart.Tick)
	require.Equal(t, uint64(4), branchBeforeStart.Tick)
	require.Equal(t, int64(5), baseBeforeStart.UntilTick)
	require.Equal(t, int64(5), branchBeforeStart.UntilTick)

	events := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StartSimulation(branchID))

	branchEnded := requireNextSimulationEndedEvent(t, events, 2*time.Second, branchID)
	require.Equal(t, uint64(5), branchEnded.Tick)

	branchAfterEnd := requireSimulationInfo(t, svc, branchID)
	baseAfterBranchEnd := requireSimulationInfo(t, svc, services.BaseSimulationID)
	require.False(t, branchAfterEnd.Running)
	require.False(t, branchAfterEnd.Paused)
	require.Equal(t, uint64(5), branchAfterEnd.Tick)
	require.True(t, baseAfterBranchEnd.Running)
	require.False(t, baseAfterBranchEnd.Paused)
	require.Less(t, baseAfterBranchEnd.Tick, uint64(5))

	requireNoSimulationEndedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)

	baseEnded := requireNextSimulationEndedEvent(t, events, 2*time.Second, services.BaseSimulationID)
	require.Equal(t, uint64(5), baseEnded.Tick)

	baseAfterEnd := requireSimulationInfo(t, svc, services.BaseSimulationID)
	branchAfterBaseEnd := requireSimulationInfo(t, svc, branchID)
	require.False(t, baseAfterEnd.Running)
	require.False(t, baseAfterEnd.Paused)
	require.Equal(t, uint64(5), baseAfterEnd.Tick)
	require.False(t, branchAfterBaseEnd.Running)
	require.False(t, branchAfterBaseEnd.Paused)
	require.Equal(t, uint64(5), branchAfterBaseEnd.Tick)
}

func TestSimulationServiceEndToEnd_StepSimulationMissingAndResetErrorsEmitNoStepEvents(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	events := subscribeToServiceEvents(t, svc)

	err := svc.StepSimulation(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)
	requireNoSimulationStepEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)

	_, err = svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	missingBranchID := "branch-missing"
	err = svc.StepSimulation(missingBranchID)
	require.ErrorIs(t, err, services.ErrSimulationNotFound)
	requireNoSimulationStepEvent(t, events, 150*time.Millisecond, missingBranchID)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)
	require.NoError(t, svc.ResetSimulation(branchID))

	err = svc.StepSimulation(branchID)
	require.ErrorIs(t, err, services.ErrSimulationNotFound)
	requireNoSimulationStepEvent(t, events, 150*time.Millisecond, branchID)

	require.NoError(t, svc.ResetSimulation(services.BaseSimulationID))
	err = svc.StepSimulation(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)
	requireNoSimulationStepEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
}

func TestSimulationServiceEndToEnd_StepSimulationIdleBaseAdvancesReadModelAndEmitsExactStepEvent(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	before, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)

	events := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

	step := requireNextSimulationStepEvent(t, events, time.Second, services.BaseSimulationID)
	after, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)

	require.False(t, after.Running)
	require.False(t, after.Paused)
	require.Equal(t, before.Tick+1, after.Tick)
	require.Equal(t, before.Timestamp.Add(10*time.Minute), after.Timestamp)
	require.Equal(t, services.EventTypeSimulationStep, step.Type)
	require.Equal(t, services.BaseSimulationID, step.SimulationID)
	require.Equal(t, after.Tick, step.Tick)
	require.Equal(t, after.Timestamp, step.Timestamp)
	requireNoSimulationStepEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
}

func TestSimulationServiceEndToEnd_StepSimulationIdleBranchAdvancesOnlyBranchAndPreservesLineage(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	beforeBase, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	beforeBranch, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, beforeBranch.ParentID)
	require.Equal(t, services.BaseSimulationID, *beforeBranch.ParentID)
	require.NotNil(t, beforeBranch.SplitTick)
	require.NotNil(t, beforeBranch.SplitTimestamp)

	events := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StepSimulation(branchID))

	step := requireNextSimulationStepEvent(t, events, time.Second, branchID)
	afterBase, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	afterBranch, err := svc.Simulation(branchID)
	require.NoError(t, err)

	require.Equal(t, beforeBase.Tick, afterBase.Tick)
	require.Equal(t, beforeBase.Timestamp, afterBase.Timestamp)
	require.False(t, afterBranch.Running)
	require.False(t, afterBranch.Paused)
	require.Equal(t, beforeBranch.Tick+1, afterBranch.Tick)
	require.Equal(t, beforeBranch.Timestamp.Add(10*time.Minute), afterBranch.Timestamp)
	require.NotNil(t, afterBranch.ParentID)
	require.Equal(t, services.BaseSimulationID, *afterBranch.ParentID)
	require.NotNil(t, afterBranch.SplitTick)
	require.NotNil(t, afterBranch.SplitTimestamp)
	require.Equal(t, *beforeBranch.SplitTick, *afterBranch.SplitTick)
	require.Equal(t, *beforeBranch.SplitTimestamp, *afterBranch.SplitTimestamp)
	require.Equal(t, services.EventTypeSimulationStep, step.Type)
	require.Equal(t, branchID, step.SimulationID)
	require.Equal(t, afterBranch.Tick, step.Tick)
	require.Equal(t, afterBranch.Timestamp, step.Timestamp)
	requireNoSimulationStepEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
	requireNoSimulationStepEvent(t, events, 150*time.Millisecond, branchID)
}

func TestSimulationServiceEndToEnd_StepSimulationPausedBaseAndBranchAdvanceDeterministically(t *testing.T) {
	t.Parallel()

	svc := newControlledRunnerService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	runnerEvents := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))
	requireNextSimulationStepEvent(t, runnerEvents, 2*time.Second, services.BaseSimulationID)
	requireNextSimulationStepEvent(t, runnerEvents, 2*time.Second, branchID)

	require.NoError(t, svc.PauseSimulation(services.BaseSimulationID))

	require.Eventually(t, func() bool {
		baseInfo, baseErr := svc.Simulation(services.BaseSimulationID)
		if baseErr != nil || !baseInfo.Running || !baseInfo.Paused {
			return false
		}
		branchInfo, branchErr := svc.Simulation(branchID)
		if branchErr != nil || !branchInfo.Running || !branchInfo.Paused {
			return false
		}
		return true
	}, time.Second, 10*time.Millisecond)

	beforeBase, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	beforeBranch, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, beforeBranch.ParentID)
	require.Equal(t, services.BaseSimulationID, *beforeBranch.ParentID)
	require.NotNil(t, beforeBranch.SplitTick)
	require.NotNil(t, beforeBranch.SplitTimestamp)

	events := subscribeToServiceEvents(t, svc)
	requireNoSimulationStepEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
	requireNoSimulationStepEvent(t, events, 150*time.Millisecond, branchID)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	require.NoError(t, svc.StepSimulation(branchID))

	baseStep := requireNextSimulationStepEvent(t, events, time.Second, services.BaseSimulationID)
	branchStep := requireNextSimulationStepEvent(t, events, time.Second, branchID)
	afterBase, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	afterBranch, err := svc.Simulation(branchID)
	require.NoError(t, err)

	require.True(t, afterBase.Running)
	require.True(t, afterBase.Paused)
	require.Equal(t, beforeBase.Tick+1, afterBase.Tick)
	require.Equal(t, beforeBase.Timestamp.Add(5*time.Second), afterBase.Timestamp)
	require.Equal(t, services.EventTypeSimulationStep, baseStep.Type)
	require.Equal(t, services.BaseSimulationID, baseStep.SimulationID)
	require.Equal(t, afterBase.Tick, baseStep.Tick)
	require.Equal(t, afterBase.Timestamp, baseStep.Timestamp)

	require.True(t, afterBranch.Running)
	require.True(t, afterBranch.Paused)
	require.Equal(t, beforeBranch.Tick+1, afterBranch.Tick)
	require.Equal(t, beforeBranch.Timestamp.Add(5*time.Second), afterBranch.Timestamp)
	require.NotNil(t, afterBranch.ParentID)
	require.Equal(t, services.BaseSimulationID, *afterBranch.ParentID)
	require.NotNil(t, afterBranch.SplitTick)
	require.NotNil(t, afterBranch.SplitTimestamp)
	require.Equal(t, *beforeBranch.SplitTick, *afterBranch.SplitTick)
	require.Equal(t, *beforeBranch.SplitTimestamp, *afterBranch.SplitTimestamp)
	require.Equal(t, services.EventTypeSimulationStep, branchStep.Type)
	require.Equal(t, branchID, branchStep.SimulationID)
	require.Equal(t, afterBranch.Tick, branchStep.Tick)
	require.Equal(t, afterBranch.Timestamp, branchStep.Timestamp)

	requireNoSimulationStepEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
	requireNoSimulationStepEvent(t, events, 150*time.Millisecond, branchID)
}

func TestSimulationServiceEndToEnd_ResetRunningBaseEmitsResetCloseSummaryAndDoesNotEmitSimulationEndedEvent(t *testing.T) {
	t.Parallel()

	svc := newControlledRunnerService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	events := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))
	requireNextSimulationStepEvent(t, events, 2*time.Second, services.BaseSimulationID)

	require.NoError(t, svc.ResetSimulation(services.BaseSimulationID))
	closed := requireNextSimulationClosedEventPayload(t, events, 2*time.Second, services.BaseSimulationID)
	require.Equal(t, "simulation_closed", closed["type"])
	require.Equal(t, services.BaseSimulationID, closed["simulationId"])
	require.Equal(t, "reset", closed["reason"])

	summaryRaw, ok := closed["summary"]
	require.True(t, ok)
	summary, ok := summaryRaw.(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(0), summary["completedVisitCount"])
	require.Equal(t, float64(0), summary["totalDurationMs"])
	require.Nil(t, summary["averageDurationMs"])
	_, hasNestedServicing := summary["servicing"]
	require.False(t, hasNestedServicing)

	requireNoSimulationEndedEvent(t, events, 200*time.Millisecond, services.BaseSimulationID)
	requireNoSimulationClosedEventPayload(t, events, 200*time.Millisecond, services.BaseSimulationID)

	_, err = svc.Simulation(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)
	require.Empty(t, svc.Simulations())
}

func TestSimulationServiceEndToEnd_ResetBranchEmitsCancelCloseSummaryDoesNotEmitSimulationEndedEventAndPreservesOtherSimulations(t *testing.T) {
	t.Parallel()

	svc := newControlledRunnerService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	branchToReset, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)
	unaffectedBranch, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	events := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))

	requireNextSimulationStepEvent(t, events, 2*time.Second, services.BaseSimulationID)
	requireNextSimulationStepEvent(t, events, 2*time.Second, branchToReset)
	requireNextSimulationStepEvent(t, events, 2*time.Second, unaffectedBranch)

	baseBeforeReset, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	unaffectedBeforeReset, err := svc.Simulation(unaffectedBranch)
	require.NoError(t, err)

	require.NoError(t, svc.ResetSimulation(branchToReset))
	closed := requireNextSimulationClosedEventPayload(t, events, 2*time.Second, branchToReset)
	require.Equal(t, "simulation_closed", closed["type"])
	require.Equal(t, branchToReset, closed["simulationId"])
	require.Equal(t, "cancel", closed["reason"])

	summaryRaw, ok := closed["summary"]
	require.True(t, ok)
	summary, ok := summaryRaw.(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(0), summary["completedVisitCount"])
	require.Equal(t, float64(0), summary["totalDurationMs"])
	require.Nil(t, summary["averageDurationMs"])
	_, hasNestedServicing := summary["servicing"]
	require.False(t, hasNestedServicing)

	requireNoSimulationEndedEvent(t, events, 200*time.Millisecond, branchToReset)
	requireNoSimulationClosedEventPayload(t, events, 200*time.Millisecond, branchToReset)

	_, err = svc.Simulation(branchToReset)
	require.ErrorIs(t, err, services.ErrSimulationNotFound)

	_, err = svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	_, err = svc.Simulation(unaffectedBranch)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		baseInfo, baseErr := svc.Simulation(services.BaseSimulationID)
		if baseErr != nil {
			return false
		}
		unaffectedInfo, branchErr := svc.Simulation(unaffectedBranch)
		if branchErr != nil {
			return false
		}
		return baseInfo.Tick > baseBeforeReset.Tick && unaffectedInfo.Tick > unaffectedBeforeReset.Tick
	}, 2*time.Second, 10*time.Millisecond)

	listed := svc.Simulations()
	require.Len(t, listed, 2)
	listedByID := simulationInfoIndexByID(t, listed)
	require.Len(t, listedByID, 2)
	_, ok = listedByID[services.BaseSimulationID]
	require.True(t, ok)
	_, ok = listedByID[unaffectedBranch]
	require.True(t, ok)
}

func TestSimulationServiceEndToEnd_BranchIsolationResetKeepsSiblingAndBaseScopedAndMetadataStable(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

	baseAtFirstFork, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)

	branchToReset, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)
	branchToResetAtFork, err := svc.Simulation(branchToReset)
	require.NoError(t, err)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	baseAtSecondFork, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)

	siblingBranch, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)
	siblingAtFork, err := svc.Simulation(siblingBranch)
	require.NoError(t, err)

	require.NotNil(t, branchToResetAtFork.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchToResetAtFork.ParentID)
	require.NotNil(t, branchToResetAtFork.SplitTick)
	require.NotNil(t, branchToResetAtFork.SplitTimestamp)
	require.Equal(t, baseAtFirstFork.Tick, *branchToResetAtFork.SplitTick)
	require.Equal(t, baseAtFirstFork.Timestamp, *branchToResetAtFork.SplitTimestamp)

	require.NotNil(t, siblingAtFork.ParentID)
	require.Equal(t, services.BaseSimulationID, *siblingAtFork.ParentID)
	require.NotNil(t, siblingAtFork.SplitTick)
	require.NotNil(t, siblingAtFork.SplitTimestamp)
	require.Equal(t, baseAtSecondFork.Tick, *siblingAtFork.SplitTick)
	require.Equal(t, baseAtSecondFork.Timestamp, *siblingAtFork.SplitTimestamp)
	require.NotEqual(t, *branchToResetAtFork.SplitTick, *siblingAtFork.SplitTick)

	baseBeforeIsolatedMutation, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	siblingBeforeIsolatedMutation, err := svc.Simulation(siblingBranch)
	require.NoError(t, err)

	require.NoError(t, svc.StepSimulation(branchToReset))
	require.NoError(t, svc.StepSimulation(branchToReset))

	branchAfterIsolatedMutation, err := svc.Simulation(branchToReset)
	require.NoError(t, err)
	baseAfterIsolatedMutation, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	siblingAfterIsolatedMutation, err := svc.Simulation(siblingBranch)
	require.NoError(t, err)

	require.Equal(t, baseBeforeIsolatedMutation.Tick, baseAfterIsolatedMutation.Tick)
	require.Equal(t, baseBeforeIsolatedMutation.Timestamp, baseAfterIsolatedMutation.Timestamp)
	require.Equal(t, siblingBeforeIsolatedMutation.Tick, siblingAfterIsolatedMutation.Tick)
	require.Equal(t, siblingBeforeIsolatedMutation.Timestamp, siblingAfterIsolatedMutation.Timestamp)
	require.Equal(t, branchToResetAtFork.Tick+2, branchAfterIsolatedMutation.Tick)
	require.Equal(t, branchToResetAtFork.Timestamp.Add(20*time.Minute), branchAfterIsolatedMutation.Timestamp)

	require.NotNil(t, branchAfterIsolatedMutation.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchAfterIsolatedMutation.ParentID)
	require.NotNil(t, branchAfterIsolatedMutation.SplitTick)
	require.Equal(t, *branchToResetAtFork.SplitTick, *branchAfterIsolatedMutation.SplitTick)
	require.NotNil(t, branchAfterIsolatedMutation.SplitTimestamp)
	require.Equal(t, *branchToResetAtFork.SplitTimestamp, *branchAfterIsolatedMutation.SplitTimestamp)

	require.NoError(t, svc.ResetSimulation(branchToReset))

	_, err = svc.Simulation(branchToReset)
	require.ErrorIs(t, err, services.ErrSimulationNotFound)
	_, err = svc.Airbases(branchToReset)
	require.ErrorIs(t, err, services.ErrSimulationNotFound)
	_, err = svc.Aircrafts(branchToReset)
	require.ErrorIs(t, err, services.ErrSimulationNotFound)
	_, err = svc.Threats(branchToReset)
	require.ErrorIs(t, err, services.ErrSimulationNotFound)

	baseAfterReset, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	siblingAfterReset, err := svc.Simulation(siblingBranch)
	require.NoError(t, err)

	require.Equal(t, baseAfterIsolatedMutation.Tick, baseAfterReset.Tick)
	require.Equal(t, baseAfterIsolatedMutation.Timestamp, baseAfterReset.Timestamp)
	require.Equal(t, siblingAfterIsolatedMutation.Tick, siblingAfterReset.Tick)
	require.Equal(t, siblingAfterIsolatedMutation.Timestamp, siblingAfterReset.Timestamp)

	events := subscribeToServiceEvents(t, svc)

	err = svc.StepSimulation(branchToReset)
	require.ErrorIs(t, err, services.ErrSimulationNotFound)
	requireNoSimulationStepEvent(t, events, 150*time.Millisecond, branchToReset)

	baseBeforeStepEvent, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	siblingBeforeStepEvent, err := svc.Simulation(siblingBranch)
	require.NoError(t, err)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	baseStep := requireNextSimulationStepEvent(t, events, time.Second, services.BaseSimulationID)
	require.Equal(t, services.BaseSimulationID, baseStep.SimulationID)
	require.Equal(t, baseBeforeStepEvent.Tick+1, baseStep.Tick)

	require.NoError(t, svc.StepSimulation(siblingBranch))
	siblingStep := requireNextSimulationStepEvent(t, events, time.Second, siblingBranch)
	require.Equal(t, siblingBranch, siblingStep.SimulationID)
	require.Equal(t, siblingBeforeStepEvent.Tick+1, siblingStep.Tick)

	requireNoSimulationStepEvent(t, events, 150*time.Millisecond, branchToReset)

	baseAfterScopedSteps, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	siblingAfterScopedSteps, err := svc.Simulation(siblingBranch)
	require.NoError(t, err)

	require.Equal(t, baseBeforeStepEvent.Tick+1, baseAfterScopedSteps.Tick)
	require.Equal(t, siblingBeforeStepEvent.Tick+1, siblingAfterScopedSteps.Tick)

	require.NotNil(t, siblingAfterScopedSteps.ParentID)
	require.Equal(t, services.BaseSimulationID, *siblingAfterScopedSteps.ParentID)
	require.NotNil(t, siblingAfterScopedSteps.SplitTick)
	require.Equal(t, *siblingAtFork.SplitTick, *siblingAfterScopedSteps.SplitTick)
	require.NotNil(t, siblingAfterScopedSteps.SplitTimestamp)
	require.Equal(t, *siblingAtFork.SplitTimestamp, *siblingAfterScopedSteps.SplitTimestamp)
}

func TestSimulationServiceEndToEnd_ResetBaseWithBranchesClearsObservableState(t *testing.T) {
	t.Parallel()

	svc := newControlledRunnerService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	branchA, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)
	branchB, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))
	require.Eventually(t, func() bool {
		baseInfo, baseErr := svc.Simulation(services.BaseSimulationID)
		if baseErr != nil || !baseInfo.Running {
			return false
		}
		branchAInfo, branchAErr := svc.Simulation(branchA)
		if branchAErr != nil || !branchAInfo.Running {
			return false
		}
		branchBInfo, branchBErr := svc.Simulation(branchB)
		if branchBErr != nil || !branchBInfo.Running {
			return false
		}
		return true
	}, 2*time.Second, 10*time.Millisecond)

	require.NoError(t, svc.ResetSimulation(services.BaseSimulationID))

	require.Empty(t, svc.Simulations())
	_, hasBase := svc.Base()
	require.False(t, hasBase)

	_, err = svc.Simulation(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)
	_, err = svc.Simulation(branchA)
	require.ErrorIs(t, err, services.ErrSimulationNotFound)
	_, err = svc.Simulation(branchB)
	require.ErrorIs(t, err, services.ErrSimulationNotFound)

	_, err = svc.Airbases(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)
	_, err = svc.Airbases(branchA)
	require.ErrorIs(t, err, services.ErrSimulationNotFound)
	_, err = svc.Aircrafts(branchB)
	require.ErrorIs(t, err, services.ErrSimulationNotFound)
}

func TestSimulationServiceBranch_BaseReturnsRandomNonBaseID(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)
	require.NotEmpty(t, branchID)
	require.NotEqual(t, services.BaseSimulationID, branchID)

	baseInfo, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)

	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfo.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchInfo.ParentID)
	require.NotNil(t, branchInfo.SplitTick)
	require.Equal(t, uint64(0), *branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, baseInfo.Tick, *branchInfo.SplitTick)
	require.Equal(t, baseInfo.Timestamp, *branchInfo.SplitTimestamp)

	list := svc.Simulations()
	listedByID := simulationInfoIndexByID(t, list)
	require.Len(t, listedByID, 2)
	_, ok := listedByID[services.BaseSimulationID]
	require.True(t, ok)
	_, ok = listedByID[branchID]
	require.True(t, ok)
}

func TestSimulationServiceBranch_PausesSourceAndBranch(t *testing.T) {
	t.Parallel()

	svc := newControlledRunnerService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	baseInfo, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)

	require.True(t, baseInfo.Paused)
	require.True(t, branchInfo.Paused)
	require.True(t, baseInfo.Running)
	require.True(t, branchInfo.Running)
}

func TestSimulationServiceBranch_IdleBaseRemainsStartableAfterBranch(t *testing.T) {
	t.Parallel()

	svc := newControlledRunnerService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)
	require.NotEmpty(t, branchID)

	baseInfo, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfo.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchInfo.ParentID)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, baseInfo.Tick, *branchInfo.SplitTick)
	require.Equal(t, baseInfo.Timestamp, *branchInfo.SplitTimestamp)
	require.False(t, baseInfo.Running)
	require.False(t, baseInfo.Paused)
	require.False(t, branchInfo.Running)
	require.False(t, branchInfo.Paused)

	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))
}

func TestSimulationServiceBranchCreatedEvent_IdleBaseEmitsExactLineagePayload(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	events := subscribeToServiceEvents(t, svc)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	event := requireNextBranchCreatedEvent(t, events, time.Second, services.BaseSimulationID)
	require.Equal(t, services.EventTypeBranchCreated, event.Type)
	require.Equal(t, services.BaseSimulationID, event.SimulationID)
	require.Equal(t, branchID, event.BranchID)
	require.Equal(t, services.BaseSimulationID, event.ParentID)

	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, *branchInfo.SplitTick, event.SplitTick)
	require.Equal(t, *branchInfo.SplitTimestamp, event.SplitTimestamp)

	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
}

func TestSimulationServiceBranchCreatedEvent_RunningBaseEmitsExactLineagePayload(t *testing.T) {
	t.Parallel()

	svc := newControlledRunnerService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))

	events := subscribeToServiceEvents(t, svc)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	event := requireNextBranchCreatedEvent(t, events, time.Second, services.BaseSimulationID)
	require.Equal(t, services.EventTypeBranchCreated, event.Type)
	require.Equal(t, services.BaseSimulationID, event.SimulationID)
	require.Equal(t, branchID, event.BranchID)
	require.Equal(t, services.BaseSimulationID, event.ParentID)

	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, *branchInfo.SplitTick, event.SplitTick)
	require.Equal(t, *branchInfo.SplitTimestamp, event.SplitTimestamp)

	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
}

func TestSimulationServiceBranchCreatedEvent_MultipleSequentialBaseBranchesEmitSeparateEvents(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	events := subscribeToServiceEvents(t, svc)

	firstBranchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	firstEvent := requireNextBranchCreatedEvent(t, events, time.Second, services.BaseSimulationID)
	require.Equal(t, services.EventTypeBranchCreated, firstEvent.Type)
	require.Equal(t, services.BaseSimulationID, firstEvent.SimulationID)
	require.Equal(t, firstBranchID, firstEvent.BranchID)
	require.Equal(t, services.BaseSimulationID, firstEvent.ParentID)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

	secondBranchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	secondEvent := requireNextBranchCreatedEvent(t, events, time.Second, services.BaseSimulationID)
	require.Equal(t, services.EventTypeBranchCreated, secondEvent.Type)
	require.Equal(t, services.BaseSimulationID, secondEvent.SimulationID)
	require.Equal(t, secondBranchID, secondEvent.BranchID)
	require.Equal(t, services.BaseSimulationID, secondEvent.ParentID)

	firstInfo, err := svc.Simulation(firstBranchID)
	require.NoError(t, err)
	require.NotNil(t, firstInfo.SplitTick)
	require.NotNil(t, firstInfo.SplitTimestamp)
	require.Equal(t, *firstInfo.SplitTick, firstEvent.SplitTick)
	require.Equal(t, *firstInfo.SplitTimestamp, firstEvent.SplitTimestamp)

	secondInfo, err := svc.Simulation(secondBranchID)
	require.NoError(t, err)
	require.NotNil(t, secondInfo.SplitTick)
	require.NotNil(t, secondInfo.SplitTimestamp)
	require.Equal(t, *secondInfo.SplitTick, secondEvent.SplitTick)
	require.Equal(t, *secondInfo.SplitTimestamp, secondEvent.SplitTimestamp)

	require.NotEqual(t, firstEvent.BranchID, secondEvent.BranchID)
	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
}

func TestSimulationServiceBranchCreatedEvent_OmitsSourceEventForLegacyBranchCreation(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	events := subscribeToServiceEvents(t, svc)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	event := requireNextBranchCreatedEvent(t, events, time.Second, services.BaseSimulationID)
	require.Equal(t, services.EventTypeBranchCreated, event.Type)
	require.Equal(t, services.BaseSimulationID, event.SimulationID)
	require.Equal(t, branchID, event.BranchID)
	require.Equal(t, services.BaseSimulationID, event.ParentID)
	require.Nil(t, event.SourceEvent)

	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, *branchInfo.SplitTick, event.SplitTick)
	require.Equal(t, *branchInfo.SplitTimestamp, event.SplitTimestamp)
	require.Nil(t, branchInfo.SourceEvent)

	listedByID := simulationInfoIndexByID(t, svc.Simulations())
	listedBranchInfo, ok := listedByID[branchID]
	require.True(t, ok)
	require.Nil(t, listedBranchInfo.SourceEvent)

	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
}

func TestSimulationServiceBranchNoEvent_MissingBaseBranchAttempt(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	events := subscribeToServiceEvents(t, svc)

	_, err := svc.BranchSimulation(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)
	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
}

func TestSimulationServiceBranchNoEvent_NonBaseBranchAttempt(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	events := subscribeToServiceEvents(t, svc)

	_, err = svc.BranchSimulation("branch-a")
	require.ErrorIs(t, err, services.ErrSimulationNotFound)
	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
}

func TestSimulationServiceBranch_MissingBaseFails(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()

	_, err := svc.BranchSimulation(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)
}

func TestSimulationServiceBranch_NonBaseSimulationRejectedInV1(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(1, 1))
	require.NoError(t, err)

	_, err = svc.BranchSimulation("branch-a")
	require.ErrorIs(t, err, services.ErrSimulationNotFound)
}

func TestSimulationServiceBranch_ReadModelsAccessibleByBranchID(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(2, 2))
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	airbases, err := svc.Airbases(branchID)
	require.NoError(t, err)
	require.Len(t, airbases, 2)

	aircrafts, err := svc.Aircrafts(branchID)
	require.NoError(t, err)
	require.Len(t, aircrafts, 2)

	threats, err := svc.Threats(branchID)
	require.NoError(t, err)
	require.NotNil(t, threats)
}

func TestSimulationServiceBranch_EmitsBranchScopedEventsOnly(t *testing.T) {
	t.Parallel()

	svc := newPositionTrackingService()
	_, err := svc.CreateBaseSimulation(positionTrackingBaseConfig([32]byte{9, 9, 9}, 2))
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.False(t, branchInfo.Paused)
	require.False(t, branchInfo.Running)

	rawEvents := subscribeToServiceEvents(t, svc)

	require.NoError(t, svc.StepSimulation(branchID))
	step := requireNextSimulationStepEvent(t, rawEvents, 2*time.Second, branchID)
	require.Equal(t, branchID, step.SimulationID)
	positions := requireNextAllAircraftPositionsEvent(t, rawEvents, 2*time.Second, branchID)
	require.Equal(t, branchID, positions.SimulationID)
	requireNoSimulationStepEvent(t, rawEvents, 150*time.Millisecond, services.BaseSimulationID)
	requireNoAllAircraftPositionsEvent(t, rawEvents, 150*time.Millisecond, services.BaseSimulationID)

}

func TestSimulationServiceBranch_DeterministicParityAfterEquivalentAdvancement(t *testing.T) {
	t.Parallel()

	svc := newPositionTrackingService()
	_, err := svc.CreateBaseSimulation(positionTrackingBaseConfig([32]byte{4, 5, 6}, 3))
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	for range 5 {
		require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
		require.NoError(t, svc.StepSimulation(branchID))
	}

	baseInfo, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.Equal(t, baseInfo.Tick, branchInfo.Tick)
	require.Equal(t, baseInfo.Timestamp, branchInfo.Timestamp)
	require.NotNil(t, branchInfo.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchInfo.ParentID)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, uint64(0), *branchInfo.SplitTick)
	baseSplitTick := *branchInfo.SplitTick
	baseSplitTimestamp := *branchInfo.SplitTimestamp

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	require.NoError(t, svc.StepSimulation(branchID))
	require.NoError(t, svc.StepSimulation(branchID))

	branchInfoAfter, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfoAfter.SplitTick)
	require.NotNil(t, branchInfoAfter.SplitTimestamp)
	require.Equal(t, baseSplitTick, *branchInfoAfter.SplitTick)
	require.Equal(t, baseSplitTimestamp, *branchInfoAfter.SplitTimestamp)

	baseAircraft, err := svc.Aircrafts(services.BaseSimulationID)
	require.NoError(t, err)
	branchAircraft, err := svc.Aircrafts(branchID)
	require.NoError(t, err)
	require.Equal(t, baseAircraft, branchAircraft)

	baseThreats, err := svc.Threats(services.BaseSimulationID)
	require.NoError(t, err)
	branchThreats, err := svc.Threats(branchID)
	require.NoError(t, err)
	require.Equal(t, baseThreats, branchThreats)
}

func TestSimulationServiceEndToEnd_ReadModelParity_BranchCreationClonesBaseCollectionsAndListDetail(t *testing.T) {
	t.Parallel()

	svc := newPositionTrackingService()
	_, err := svc.CreateBaseSimulation(positionTrackingBaseConfig([32]byte{2, 4, 6}, 4))
	require.NoError(t, err)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

	beforeBranch := captureReadModelSnapshots(t, svc, services.BaseSimulationID)[services.BaseSimulationID]
	require.Len(t, beforeBranch.airbases, 1)
	require.Len(t, beforeBranch.aircrafts, 4)
	require.Len(t, beforeBranch.threats, 2)

	events := subscribeToServiceEvents(t, svc)
	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)
	branchCreated := requireNextBranchCreatedEvent(t, events, time.Second, services.BaseSimulationID)

	snapshots := captureReadModelSnapshots(t, svc, services.BaseSimulationID, branchID)
	baseSnapshot := snapshots[services.BaseSimulationID]
	branchSnapshot := snapshots[branchID]

	require.Equal(t, beforeBranch.info, baseSnapshot.info)
	require.Equal(t, baseSnapshot.airbases, branchSnapshot.airbases)
	require.Equal(t, baseSnapshot.aircrafts, branchSnapshot.aircrafts)
	require.Equal(t, baseSnapshot.threats, branchSnapshot.threats)

	require.NotNil(t, branchSnapshot.info.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchSnapshot.info.ParentID)
	require.NotNil(t, branchSnapshot.info.SplitTick)
	require.Equal(t, baseSnapshot.info.Tick, *branchSnapshot.info.SplitTick)
	require.NotNil(t, branchSnapshot.info.SplitTimestamp)
	require.Equal(t, baseSnapshot.info.Timestamp, *branchSnapshot.info.SplitTimestamp)
	require.Nil(t, branchSnapshot.info.SourceEvent)

	require.Equal(t, services.EventTypeBranchCreated, branchCreated.Type)
	require.Equal(t, services.BaseSimulationID, branchCreated.SimulationID)
	require.Equal(t, branchID, branchCreated.BranchID)
	require.Equal(t, services.BaseSimulationID, branchCreated.ParentID)
	require.Equal(t, baseSnapshot.info.Tick, branchCreated.SplitTick)
	require.Equal(t, baseSnapshot.info.Timestamp, branchCreated.SplitTimestamp)
	require.Nil(t, branchCreated.SourceEvent)
}

func TestSimulationServiceEndToEnd_ReadModelParity_PausedLifecycleStepsMatchEventsAndListDetail(t *testing.T) {
	t.Parallel()

	svc := newControlledRunnerService()
	_, err := svc.CreateBaseSimulation(deterministicBaseConfig(2, 2))
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	runnerEvents := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))
	requireNextSimulationStepEventAfterTick(t, runnerEvents, 2*time.Second, services.BaseSimulationID, 0)
	requireNextSimulationStepEventAfterTick(t, runnerEvents, 2*time.Second, branchID, 0)

	require.NoError(t, svc.PauseSimulation(branchID))
	require.Eventually(t, func() bool {
		baseInfo, baseErr := svc.Simulation(services.BaseSimulationID)
		if baseErr != nil || !baseInfo.Running || !baseInfo.Paused {
			return false
		}

		branchInfo, branchErr := svc.Simulation(branchID)
		if branchErr != nil || !branchInfo.Running || !branchInfo.Paused {
			return false
		}

		return true
	}, time.Second, 10*time.Millisecond)

	beforeStep := captureReadModelSnapshots(t, svc, services.BaseSimulationID, branchID)
	manualEvents := subscribeToServiceEvents(t, svc)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	require.NoError(t, svc.StepSimulation(branchID))

	baseStep := requireNextSimulationStepEvent(t, manualEvents, time.Second, services.BaseSimulationID)
	branchStep := requireNextSimulationStepEvent(t, manualEvents, time.Second, branchID)
	afterStep := captureReadModelSnapshots(t, svc, services.BaseSimulationID, branchID)

	baseBefore := beforeStep[services.BaseSimulationID]
	branchBefore := beforeStep[branchID]
	baseAfter := afterStep[services.BaseSimulationID]
	branchAfter := afterStep[branchID]

	require.True(t, baseAfter.info.Running)
	require.True(t, baseAfter.info.Paused)
	require.Equal(t, baseBefore.info.Tick+1, baseAfter.info.Tick)
	require.Equal(t, baseBefore.info.Timestamp.Add(5*time.Second), baseAfter.info.Timestamp)
	require.Equal(t, baseAfter.info.Tick, baseStep.Tick)
	require.Equal(t, baseAfter.info.Timestamp, baseStep.Timestamp)

	require.True(t, branchAfter.info.Running)
	require.True(t, branchAfter.info.Paused)
	require.Equal(t, branchBefore.info.Tick+1, branchAfter.info.Tick)
	require.Equal(t, branchBefore.info.Timestamp.Add(5*time.Second), branchAfter.info.Timestamp)
	require.Equal(t, branchAfter.info.Tick, branchStep.Tick)
	require.Equal(t, branchAfter.info.Timestamp, branchStep.Timestamp)
	require.NotNil(t, branchAfter.info.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchAfter.info.ParentID)
	require.NotNil(t, branchAfter.info.SplitTick)
	require.Equal(t, *branchBefore.info.SplitTick, *branchAfter.info.SplitTick)
	require.NotNil(t, branchAfter.info.SplitTimestamp)
	require.Equal(t, *branchBefore.info.SplitTimestamp, *branchAfter.info.SplitTimestamp)

	require.Equal(t, baseAfter.airbases, branchAfter.airbases)
	require.Equal(t, baseAfter.aircrafts, branchAfter.aircrafts)
	require.Equal(t, baseAfter.threats, branchAfter.threats)
}

func TestSimulationServiceEndToEnd_ReadModelParity_DivergedBaseAndBranchSnapshotsStayOrderSafe(t *testing.T) {
	t.Parallel()

	svc := newPositionTrackingService()
	_, err := svc.CreateBaseSimulation(positionTrackingBaseConfig([32]byte{4, 5, 6}, 6))
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	for range 3 {
		require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
		require.NoError(t, svc.StepSimulation(branchID))
	}

	equalSnapshots := captureReadModelSnapshots(t, svc, services.BaseSimulationID, branchID)
	require.Equal(t, equalSnapshots[services.BaseSimulationID].airbases, equalSnapshots[branchID].airbases)
	require.Equal(t, equalSnapshots[services.BaseSimulationID].aircrafts, equalSnapshots[branchID].aircrafts)
	require.Equal(t, equalSnapshots[services.BaseSimulationID].threats, equalSnapshots[branchID].threats)

	for range 2 {
		require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	}

	divergedSnapshots := captureReadModelSnapshots(t, svc, services.BaseSimulationID, branchID)
	baseSnapshot := divergedSnapshots[services.BaseSimulationID]
	branchSnapshot := divergedSnapshots[branchID]

	require.Equal(t, branchSnapshot.info.Tick+2, baseSnapshot.info.Tick)
	require.Equal(t, branchSnapshot.info.Timestamp.Add(2*time.Second), baseSnapshot.info.Timestamp)
	require.NotNil(t, branchSnapshot.info.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchSnapshot.info.ParentID)
	require.NotNil(t, branchSnapshot.info.SplitTick)
	require.Equal(t, *equalSnapshots[branchID].info.SplitTick, *branchSnapshot.info.SplitTick)
	require.NotNil(t, branchSnapshot.info.SplitTimestamp)
	require.Equal(t, *equalSnapshots[branchID].info.SplitTimestamp, *branchSnapshot.info.SplitTimestamp)

	require.Equal(t, baseSnapshot.airbases, branchSnapshot.airbases)

	differingAircraftPositions := 0
	for tailNumber, branchAircraft := range branchSnapshot.aircrafts {
		baseAircraft, ok := baseSnapshot.aircrafts[tailNumber]
		require.Truef(t, ok, "missing aircraft %q in base snapshot", tailNumber)
		require.Equal(t, branchAircraft.TailNumber, baseAircraft.TailNumber)
		require.Equal(t, branchAircraft.Needs, baseAircraft.Needs)
		require.Equal(t, branchAircraft.State, baseAircraft.State)
		require.Equal(t, branchAircraft.AssignedTo, baseAircraft.AssignedTo)
		if branchAircraft.Position != baseAircraft.Position {
			differingAircraftPositions++
		}
	}
	require.Greater(t, differingAircraftPositions, 0)

	extraThreats := extraThreats(baseSnapshot.threats, branchSnapshot.threats)
	require.Len(t, extraThreats, 2)
	for threatID, branchThreat := range branchSnapshot.threats {
		baseThreat, ok := baseSnapshot.threats[threatID]
		require.Truef(t, ok, "missing threat %q in base snapshot", threatID)
		require.Equal(t, branchThreat, baseThreat)
	}

	sort.Slice(extraThreats, func(i, j int) bool {
		return extraThreats[i].CreatedTick < extraThreats[j].CreatedTick
	})
	require.Equal(t, []uint64{branchSnapshot.info.Tick + 1, branchSnapshot.info.Tick + 2}, []uint64{extraThreats[0].CreatedTick, extraThreats[1].CreatedTick})
	require.Equal(t, []time.Time{branchSnapshot.info.Timestamp.Add(time.Second), branchSnapshot.info.Timestamp.Add(2 * time.Second)}, []time.Time{extraThreats[0].CreatedAt, extraThreats[1].CreatedAt})
}

func TestSimulationServiceEndToEnd_ThreatSpawnedAndTargetedEvents_BaseStepMatchThreatSnapshot(t *testing.T) {
	t.Parallel()

	svc := newPositionTrackingService()
	_, err := svc.CreateBaseSimulation(deterministicThreatLifecycleBaseConfig([32]byte{7, 7, 7}))
	require.NoError(t, err)

	before := captureReadModelSnapshots(t, svc, services.BaseSimulationID)[services.BaseSimulationID]
	require.Empty(t, before.threats)
	tailNumber := requireOnlyAircraftTailNumber(t, before.aircrafts)

	events := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

	after := captureReadModelSnapshots(t, svc, services.BaseSimulationID)[services.BaseSimulationID]
	spawned := requireNextThreatSpawnedEvent(t, events, time.Second, services.BaseSimulationID)
	targeted := requireNextThreatTargetedEvent(t, events, time.Second, services.BaseSimulationID)
	expectedThreat := requireOnlyThreatSnapshot(t, after.threats)

	requireThreatSpawnedEventMatchesSnapshot(t, spawned, services.BaseSimulationID, expectedThreat, after.info.Timestamp)
	requireThreatTargetedEventMatchesSnapshot(t, targeted, services.BaseSimulationID, expectedThreat, tailNumber, after.info.Timestamp)
	require.Equal(t, spawned.Threat, targeted.Threat)
	requireNoThreatDespawnedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
}

func TestSimulationServiceEndToEnd_ThreatLifecycleEvents_BranchStepsStayScopedAndMatchThreatSnapshots(t *testing.T) {
	t.Parallel()

	svc := newPositionTrackingService()
	_, err := svc.CreateBaseSimulation(deterministicThreatLifecycleBaseConfig([32]byte{8, 8, 8}))
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	before := captureReadModelSnapshots(t, svc, services.BaseSimulationID, branchID)
	baseBefore := before[services.BaseSimulationID]
	branchBefore := before[branchID]
	require.Empty(t, baseBefore.threats)
	require.Empty(t, branchBefore.threats)
	tailNumber := requireOnlyAircraftTailNumber(t, branchBefore.aircrafts)

	events := subscribeToServiceEvents(t, svc)
	require.NoError(t, svc.StepSimulation(branchID))

	afterSpawnAndTarget := captureReadModelSnapshots(t, svc, services.BaseSimulationID, branchID)
	baseAfterSpawnAndTarget := afterSpawnAndTarget[services.BaseSimulationID]
	branchAfterSpawnAndTarget := afterSpawnAndTarget[branchID]
	spawned := requireNextThreatSpawnedEvent(t, events, time.Second, branchID)
	targeted := requireNextThreatTargetedEvent(t, events, time.Second, branchID)
	expectedThreat := requireOnlyThreatSnapshot(t, branchAfterSpawnAndTarget.threats)

	require.Equal(t, baseBefore.info, baseAfterSpawnAndTarget.info)
	require.Equal(t, baseBefore.airbases, baseAfterSpawnAndTarget.airbases)
	require.Equal(t, baseBefore.aircrafts, baseAfterSpawnAndTarget.aircrafts)
	require.Equal(t, baseBefore.threats, baseAfterSpawnAndTarget.threats)
	requireThreatSpawnedEventMatchesSnapshot(t, spawned, branchID, expectedThreat, branchAfterSpawnAndTarget.info.Timestamp)
	requireThreatTargetedEventMatchesSnapshot(t, targeted, branchID, expectedThreat, tailNumber, branchAfterSpawnAndTarget.info.Timestamp)
	require.Equal(t, spawned.Threat, targeted.Threat)
	requireNoThreatSpawnedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
	requireNoThreatTargetedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
	requireNoThreatDespawnedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
	requireNoThreatDespawnedEvent(t, events, 150*time.Millisecond, branchID)

	require.NoError(t, svc.StepSimulation(branchID))

	afterDespawn := captureReadModelSnapshots(t, svc, services.BaseSimulationID, branchID)
	baseAfterDespawn := afterDespawn[services.BaseSimulationID]
	branchAfterDespawn := afterDespawn[branchID]
	despawned := requireNextThreatDespawnedEvent(t, events, time.Second, branchID)

	require.Equal(t, baseBefore.info, baseAfterDespawn.info)
	require.Equal(t, baseBefore.airbases, baseAfterDespawn.airbases)
	require.Equal(t, baseBefore.aircrafts, baseAfterDespawn.aircrafts)
	require.Equal(t, baseBefore.threats, baseAfterDespawn.threats)
	requireThreatDespawnedEventMatchesSnapshot(t, despawned, branchID, expectedThreat, branchAfterDespawn.info.Timestamp)
	require.Empty(t, branchAfterDespawn.threats)
	requireNoThreatSpawnedEvent(t, events, 150*time.Millisecond, branchID)
	requireNoThreatTargetedEvent(t, events, 150*time.Millisecond, branchID)
	requireNoThreatSpawnedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
	requireNoThreatTargetedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
	requireNoThreatDespawnedEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
}

func TestSimulationServiceEndToEnd_AircraftStateChangeEvent_BaseCommitMatchesAircraftSnapshot(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(inboundOverrideBaseConfig())
	require.NoError(t, err)

	before := captureReadModelSnapshots(t, svc, services.BaseSimulationID)
	baseBefore := before[services.BaseSimulationID]
	require.Len(t, baseBefore.aircrafts, 1)

	airbases, err := svc.Airbases(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, airbases, 2)

	var tailNumber string
	for tail := range baseBefore.aircrafts {
		tailNumber = tail
	}
	require.NotEmpty(t, tailNumber)

	expectedAssignedBaseID := airbases[0].ID
	events := subscribeToServiceEvents(t, svc)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	afterAssignment := captureReadModelSnapshots(t, svc, services.BaseSimulationID)[services.BaseSimulationID]
	landingEvent := requireNextLandingAssignmentEvent(t, events, time.Second, services.BaseSimulationID)

	require.Equal(t, services.EventTypeLandingAssignment, landingEvent.Type)
	require.Equal(t, services.BaseSimulationID, landingEvent.SimulationID)
	require.Equal(t, tailNumber, landingEvent.TailNumber)
	require.Equal(t, expectedAssignedBaseID, landingEvent.BaseID)
	require.Equal(t, services.AssignmentSourceAlgorithm, landingEvent.Source)
	require.Equal(t, afterAssignment.aircrafts[tailNumber].Needs, landingEvent.Needs)
	require.Equal(t, afterAssignment.airbases[expectedAssignedBaseID].Capabilities, landingEvent.Capabilities)
	require.Equal(t, afterAssignment.info.Timestamp, landingEvent.Timestamp)

	aircraftAfterAssignment := afterAssignment.aircrafts[tailNumber]
	require.Equal(t, "Inbound", aircraftAfterAssignment.State)
	require.Nil(t, aircraftAfterAssignment.AssignedTo)
	require.Equal(t, baseBefore.airbases, afterAssignment.airbases)
	require.Equal(t, baseBefore.threats, afterAssignment.threats)
	requireNoAircraftStateChangeEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	afterCommit := captureReadModelSnapshots(t, svc, services.BaseSimulationID)[services.BaseSimulationID]
	stateChange := requireNextAircraftStateChangeEvent(t, events, time.Second, services.BaseSimulationID)
	committedAircraft := afterCommit.aircrafts[tailNumber]

	require.Equal(t, afterAssignment.info.Tick+1, afterCommit.info.Tick)
	require.Equal(t, afterAssignment.info.Timestamp.Add(10*time.Minute), afterCommit.info.Timestamp)
	requireAircraftStateChangeEventMatchesSnapshot(
		t,
		stateChange,
		services.BaseSimulationID,
		"Inbound",
		"Committed",
		committedAircraft,
		afterCommit.info.Timestamp,
	)
	require.Equal(t, "Committed", committedAircraft.State)
	require.NotNil(t, committedAircraft.AssignedTo)
	require.Equal(t, landingEvent.BaseID, *committedAircraft.AssignedTo)
	requireNoAircraftStateChangeEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
	requireNoLandingAssignmentEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
}

func TestSimulationServiceEndToEnd_AircraftStateChangeEvent_BranchCommitMatchesAircraftSnapshot(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(inboundOverrideBaseConfig())
	require.NoError(t, err)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	before := captureReadModelSnapshots(t, svc, services.BaseSimulationID, branchID)
	baseBefore := before[services.BaseSimulationID]
	branchBefore := before[branchID]
	require.Len(t, baseBefore.aircrafts, 1)
	require.Len(t, branchBefore.aircrafts, 1)
	branchAirbases, err := svc.Airbases(branchID)
	require.NoError(t, err)
	require.Len(t, branchAirbases, 2)

	var tailNumber string
	for tail := range branchBefore.aircrafts {
		tailNumber = tail
	}
	require.NotEmpty(t, tailNumber)

	expectedAlgorithmBaseID := branchAirbases[1].ID
	targetBaseID := branchAirbases[0].ID
	require.NotEqual(t, expectedAlgorithmBaseID, targetBaseID)

	events := subscribeToServiceEvents(t, svc)
	updatedAircraft, updatedAssignment, err := svc.OverrideAssignment(branchID, tailNumber, targetBaseID)
	require.NoError(t, err)
	require.Equal(t, tailNumber, updatedAircraft.TailNumber)
	require.Equal(t, "Inbound", updatedAircraft.State)
	require.NotNil(t, updatedAircraft.AssignedTo)
	require.Equal(t, targetBaseID, *updatedAircraft.AssignedTo)
	require.Equal(t, targetBaseID, updatedAssignment.Base)
	require.Equal(t, services.AssignmentSourceHuman, updatedAssignment.Source)

	landingEvents := []services.LandingAssignmentEvent{
		requireNextLandingAssignmentEvent(t, events, time.Second, branchID),
		requireNextLandingAssignmentEvent(t, events, time.Second, branchID),
	}
	landingEventsBySource := make(map[services.LandingAssignmentSource]services.LandingAssignmentEvent, len(landingEvents))
	for _, event := range landingEvents {
		_, exists := landingEventsBySource[event.Source]
		require.Falsef(t, exists, "duplicate %q landing assignment event for simulation %q", event.Source, branchID)
		landingEventsBySource[event.Source] = event

		require.Equal(t, services.EventTypeLandingAssignment, event.Type)
		require.Equal(t, branchID, event.SimulationID)
		require.Equal(t, tailNumber, event.TailNumber)
		require.Equal(t, branchBefore.info.Timestamp, event.Timestamp)
	}

	algorithmEvent, ok := landingEventsBySource[services.AssignmentSourceAlgorithm]
	require.True(t, ok)
	require.Equal(t, expectedAlgorithmBaseID, algorithmEvent.BaseID)

	humanEvent, ok := landingEventsBySource[services.AssignmentSourceHuman]
	require.True(t, ok)
	require.Equal(t, targetBaseID, humanEvent.BaseID)

	afterOverride := captureReadModelSnapshots(t, svc, services.BaseSimulationID, branchID)
	baseAfterOverride := afterOverride[services.BaseSimulationID]
	branchAfterOverride := afterOverride[branchID]
	require.Equal(t, branchAfterOverride.aircrafts[tailNumber].Needs, algorithmEvent.Needs)
	require.Equal(t, branchAfterOverride.aircrafts[tailNumber].Needs, humanEvent.Needs)
	require.Equal(t, branchAfterOverride.airbases[expectedAlgorithmBaseID].Capabilities, algorithmEvent.Capabilities)
	require.Equal(t, branchAfterOverride.airbases[targetBaseID].Capabilities, humanEvent.Capabilities)
	require.Equal(t, baseBefore.info, baseAfterOverride.info)
	require.Equal(t, baseBefore.airbases, baseAfterOverride.airbases)
	require.Equal(t, baseBefore.aircrafts, baseAfterOverride.aircrafts)
	require.Equal(t, baseBefore.threats, baseAfterOverride.threats)
	require.Equal(t, branchBefore.info, branchAfterOverride.info)
	require.Equal(t, branchBefore.airbases, branchAfterOverride.airbases)
	require.Equal(t, branchBefore.aircrafts, branchAfterOverride.aircrafts)
	require.Equal(t, branchBefore.threats, branchAfterOverride.threats)
	requireNoLandingAssignmentEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)

	require.NoError(t, svc.StepSimulation(branchID))

	after := captureReadModelSnapshots(t, svc, services.BaseSimulationID, branchID)
	baseAfter := after[services.BaseSimulationID]
	branchAfter := after[branchID]
	stateChange := requireNextAircraftStateChangeEvent(t, events, time.Second, branchID)

	require.Equal(t, baseBefore.info, baseAfter.info)
	require.Equal(t, baseBefore.airbases, baseAfter.airbases)
	require.Equal(t, baseBefore.aircrafts, baseAfter.aircrafts)
	require.Equal(t, baseBefore.threats, baseAfter.threats)

	baseAircraft := baseAfter.aircrafts[tailNumber]
	branchAircraft := branchAfter.aircrafts[tailNumber]
	require.Equal(t, branchBefore.info.Tick+1, branchAfter.info.Tick)
	require.Equal(t, branchBefore.info.Timestamp.Add(10*time.Minute), branchAfter.info.Timestamp)
	require.Equal(t, branchBefore.airbases, branchAfter.airbases)
	require.Equal(t, branchBefore.threats, branchAfter.threats)
	requireAircraftStateChangeEventMatchesSnapshot(
		t,
		stateChange,
		branchID,
		"Inbound",
		"Committed",
		branchAircraft,
		branchAfter.info.Timestamp,
	)
	require.Equal(t, "Committed", branchAircraft.State)
	require.Nil(t, baseBefore.aircrafts[tailNumber].AssignedTo)
	require.Equal(t, "Inbound", baseAircraft.State)
	require.Nil(t, baseAircraft.AssignedTo)
	require.NotNil(t, branchAircraft.AssignedTo)
	require.Equal(t, humanEvent.BaseID, *branchAircraft.AssignedTo)
	require.NotEqual(t, baseAircraft.AssignedTo, branchAircraft.AssignedTo)
	requireNoAircraftStateChangeEvent(t, events, 150*time.Millisecond, branchID)
	requireNoAircraftStateChangeEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)
}

func TestSimulationServiceEndToEnd_OverrideAssignmentOnBranchLateOverrideRejectedWithoutMutatingBase(t *testing.T) {
	t.Parallel()

	svc := newDeterministicService()
	_, err := svc.CreateBaseSimulation(inboundOverrideBaseConfig())
	require.NoError(t, err)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	branchBeforeStep, err := svc.Simulation(branchID)
	require.NoError(t, err)

	before := captureReadModelSnapshots(t, svc, services.BaseSimulationID, branchID)
	baseBefore := before[services.BaseSimulationID]
	branchBefore := before[branchID]
	require.Len(t, branchBefore.aircrafts, 1)

	var tailNumber string
	for tail := range branchBefore.aircrafts {
		tailNumber = tail
	}
	require.NotEmpty(t, tailNumber)

	baseIDs := make([]string, 0, len(branchBefore.airbases))
	for baseID := range branchBefore.airbases {
		baseIDs = append(baseIDs, baseID)
	}
	sort.Strings(baseIDs)
	require.Len(t, baseIDs, 2)

	require.NoError(t, svc.StepSimulation(branchID))
	branchAfterStep, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.Equal(t, branchBeforeStep.Tick+1, branchAfterStep.Tick)
	require.Equal(t, branchBeforeStep.Timestamp.Add(10*time.Minute), branchAfterStep.Timestamp)

	branchAircraftAfterStep, err := svc.Aircrafts(branchID)
	require.NoError(t, err)
	require.Len(t, branchAircraftAfterStep, 1)
	require.Equal(t, "Committed", branchAircraftAfterStep[0].State)

	events := subscribeToServiceEvents(t, svc)
	_, _, err = svc.OverrideAssignment(branchID, tailNumber, baseIDs[1])
	require.ErrorIs(t, err, services.ErrAssignmentTooLate)
	requireNoLandingAssignmentEvent(t, events, 150*time.Millisecond, branchID)
	requireNoLandingAssignmentEvent(t, events, 150*time.Millisecond, services.BaseSimulationID)

	afterFailure := captureReadModelSnapshots(t, svc, services.BaseSimulationID, branchID)
	baseAfterFailure := afterFailure[services.BaseSimulationID]
	branchAfterFailure := afterFailure[branchID]

	require.Equal(t, baseBefore.info, baseAfterFailure.info)
	require.Equal(t, baseBefore.airbases, baseAfterFailure.airbases)
	require.Equal(t, baseBefore.aircrafts, baseAfterFailure.aircrafts)
	require.Equal(t, baseBefore.threats, baseAfterFailure.threats)

	require.Equal(t, branchAfterStep, branchAfterFailure.info)
	require.Equal(t, branchBefore.airbases, branchAfterFailure.airbases)
	require.Equal(t, aircraftsByTailNumber(t, branchAircraftAfterStep), branchAfterFailure.aircrafts)
	require.Equal(t, branchBefore.threats, branchAfterFailure.threats)
}

func TestSimulationServiceBranch_DeterministicEventIDsDoNotDuplicateBaseID(t *testing.T) {
	t.Parallel()

	svc := newPositionTrackingService()
	_, err := svc.CreateBaseSimulation(positionTrackingBaseConfig([32]byte{1, 3, 5}, 2))
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	rawEvents := subscribeToServiceEvents(t, svc)

	require.NoError(t, svc.StepSimulation(branchID))

	step := requireNextSimulationStepEvent(t, rawEvents, time.Second, branchID)
	require.NotEqual(t, services.BaseSimulationID, step.SimulationID)
	require.Equal(t, branchID, step.SimulationID)
	requireNoSimulationStepEvent(t, rawEvents, 25*time.Millisecond, branchID)
}

func asLandingAssignmentEvent(event services.Event) (services.LandingAssignmentEvent, bool) {
	typed, ok := event.(services.LandingAssignmentEvent)
	return typed, ok
}

func asAircraftStateChangeEvent(event services.Event) (services.AircraftStateChangeEvent, bool) {
	typed, ok := event.(services.AircraftStateChangeEvent)
	return typed, ok
}

func asThreatSpawnedEvent(event services.Event) (services.ThreatSpawnedEvent, bool) {
	typed, ok := event.(services.ThreatSpawnedEvent)
	return typed, ok
}

func asThreatTargetedEvent(event services.Event) (services.ThreatTargetedEvent, bool) {
	typed, ok := event.(services.ThreatTargetedEvent)
	return typed, ok
}

func asThreatDespawnedEvent(event services.Event) (services.ThreatDespawnedEvent, bool) {
	typed, ok := event.(services.ThreatDespawnedEvent)
	return typed, ok
}

func requireNextLandingAssignmentEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) services.LandingAssignmentEvent {
	t.Helper()
	return requireNextScopedEvent(t, watcher, timeout, services.EventTypeLandingAssignment, simulationID, asLandingAssignmentEvent)
}

func requireNoLandingAssignmentEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) {
	t.Helper()
	requireNoScopedEvent(t, watcher, timeout, services.EventTypeLandingAssignment, simulationID, asLandingAssignmentEvent)
}

func requireNextThreatSpawnedEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) services.ThreatSpawnedEvent {
	t.Helper()
	return requireNextScopedEvent(t, watcher, timeout, services.EventTypeThreatSpawned, simulationID, asThreatSpawnedEvent)
}

func requireNoThreatSpawnedEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) {
	t.Helper()
	requireNoScopedEvent(t, watcher, timeout, services.EventTypeThreatSpawned, simulationID, asThreatSpawnedEvent)
}

func requireNextThreatTargetedEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) services.ThreatTargetedEvent {
	t.Helper()
	return requireNextScopedEvent(t, watcher, timeout, services.EventTypeThreatTargeted, simulationID, asThreatTargetedEvent)
}

func requireNoThreatTargetedEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) {
	t.Helper()
	requireNoScopedEvent(t, watcher, timeout, services.EventTypeThreatTargeted, simulationID, asThreatTargetedEvent)
}

func requireNextThreatDespawnedEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) services.ThreatDespawnedEvent {
	t.Helper()
	return requireNextScopedEvent(t, watcher, timeout, services.EventTypeThreatDespawned, simulationID, asThreatDespawnedEvent)
}

func requireNoThreatDespawnedEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) {
	t.Helper()
	requireNoScopedEvent(t, watcher, timeout, services.EventTypeThreatDespawned, simulationID, asThreatDespawnedEvent)
}

func requireThreatSpawnedEventMatchesSnapshot(
	t *testing.T,
	event services.ThreatSpawnedEvent,
	simulationID string,
	expectedThreat services.Threat,
	expectedTimestamp time.Time,
) {
	t.Helper()

	require.Equal(t, services.EventTypeThreatSpawned, event.Type)
	require.Equal(t, simulationID, event.SimulationID)
	require.Equal(t, expectedThreat, event.Threat)
	require.Equal(t, expectedTimestamp, event.Timestamp)
}

func requireThreatTargetedEventMatchesSnapshot(
	t *testing.T,
	event services.ThreatTargetedEvent,
	simulationID string,
	expectedThreat services.Threat,
	expectedTailNumber string,
	expectedTimestamp time.Time,
) {
	t.Helper()

	require.Equal(t, services.EventTypeThreatTargeted, event.Type)
	require.Equal(t, simulationID, event.SimulationID)
	require.Equal(t, expectedThreat, event.Threat)
	require.Equal(t, expectedTailNumber, event.TailNumber)
	require.Equal(t, expectedTimestamp, event.Timestamp)
}

func requireThreatDespawnedEventMatchesSnapshot(
	t *testing.T,
	event services.ThreatDespawnedEvent,
	simulationID string,
	expectedThreat services.Threat,
	expectedTimestamp time.Time,
) {
	t.Helper()

	require.Equal(t, services.EventTypeThreatDespawned, event.Type)
	require.Equal(t, simulationID, event.SimulationID)
	require.Equal(t, expectedThreat, event.Threat)
	require.Equal(t, expectedTimestamp, event.Timestamp)
}

func requireAircraftStateChangeEventMatchesSnapshot(
	t *testing.T,
	event services.AircraftStateChangeEvent,
	simulationID string,
	oldState string,
	newState string,
	expectedAircraft services.Aircraft,
	expectedTimestamp time.Time,
) {
	t.Helper()

	require.Equal(t, services.EventTypeAircraftStateChange, event.Type)
	require.Equal(t, simulationID, event.SimulationID)
	require.Equal(t, expectedAircraft.TailNumber, event.TailNumber)
	require.Equal(t, oldState, event.OldState)
	require.Equal(t, newState, event.NewState)
	require.Equal(t, expectedAircraft, event.Aircraft)
	require.Equal(t, expectedTimestamp, event.Timestamp)
}

func requireNextAircraftStateChangeEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) services.AircraftStateChangeEvent {
	t.Helper()
	return requireNextScopedEvent(t, watcher, timeout, services.EventTypeAircraftStateChange, simulationID, asAircraftStateChangeEvent)
}

func requireNoAircraftStateChangeEvent(t *testing.T, watcher *serviceEventWatcher, timeout time.Duration, simulationID string) {
	t.Helper()
	requireNoScopedEvent(t, watcher, timeout, services.EventTypeAircraftStateChange, simulationID, asAircraftStateChangeEvent)
}

func TestAllAircraftPositionsEventEmitted(t *testing.T) {
	t.Parallel()

	const (
		fleetSize = 3
		steps     = 10
	)

	svc := newPositionTrackingService()
	_, err := svc.CreateBaseSimulation(positionTrackingBaseConfig([32]byte{6, 7, 8}, uint(fleetSize)))
	require.NoError(t, err)

	initialAircraft, err := svc.Aircrafts(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, initialAircraft, fleetSize)

	initialPositions := make(map[string]services.Point, fleetSize)
	for _, aircraft := range initialAircraft {
		initialPositions[aircraft.TailNumber] = aircraft.Position
	}

	rawEvents := subscribeToServiceEvents(t, svc)

	sawMovement := false
	broadcastCount := 0
	for i := range steps {
		require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

		event := requireNextAllAircraftPositionsEvent(t, rawEvents, time.Second, services.BaseSimulationID)
		require.Equal(t, services.EventTypeAllAircraftPositions, event.Type)
		require.Equal(t, services.BaseSimulationID, event.SimulationID)
		require.Equal(t, uint64(i+1), event.Tick)
		require.Len(t, event.Positions, fleetSize)

		aircrafts, err := svc.Aircrafts(services.BaseSimulationID)
		require.NoError(t, err)
		aircraftsByTail := aircraftsByTailNumber(t, aircrafts)

		for _, snapshot := range event.Positions {
			initial, ok := initialPositions[snapshot.TailNumber]
			require.True(t, ok)
			currentAircraft, ok := aircraftsByTail[snapshot.TailNumber]
			require.True(t, ok)
			require.Equal(t, currentAircraft.Position, snapshot.Position)
			require.Equal(t, currentAircraft.State, snapshot.State)
			require.Equal(t, currentAircraft.Needs, snapshot.Needs)
			if snapshot.Position != initial {
				sawMovement = true
			}
		}

		broadcastCount++
	}
	require.Equal(t, steps, broadcastCount)
	require.True(t, sawMovement)

	requireNoAllAircraftPositionsEvent(t, rawEvents, 25*time.Millisecond, services.BaseSimulationID)
}

type lifecycleExpectation struct {
	running bool
	paused  bool
}

func requireLifecycleMatrix(
	t *testing.T,
	svc *services.SimulationService,
	branchID string,
	base lifecycleExpectation,
	branch lifecycleExpectation,
) {
	t.Helper()

	requireSimulationLifecycleState(t, svc, services.BaseSimulationID, base)
	requireSimulationLifecycleState(t, svc, branchID, branch)
}

func requireSimulationLifecycleState(
	t *testing.T,
	svc *services.SimulationService,
	simulationID string,
	expected lifecycleExpectation,
) {
	t.Helper()

	info := requireSimulationInfo(t, svc, simulationID)
	require.Equalf(t, expected.running, info.Running, "unexpected running state for simulation %q", simulationID)
	require.Equalf(t, expected.paused, info.Paused, "unexpected paused state for simulation %q", simulationID)
}

func requireSimulationInfo(t *testing.T, svc *services.SimulationService, simulationID string) services.SimulationInfo {
	t.Helper()

	info, err := svc.Simulation(simulationID)
	require.NoError(t, err)
	return info
}

func requireNextSimulationStepEventAfterTick(
	t *testing.T,
	watcher *serviceEventWatcher,
	timeout time.Duration,
	simulationID string,
	minTick uint64,
) services.SimulationStepEvent {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			t.Fatalf("timed out waiting for simulation %q step event after tick %d", simulationID, minTick)
		}

		event := requireNextSimulationStepEvent(t, watcher, remaining, simulationID)
		if event.Tick > minTick {
			return event
		}
	}
}

type simulationReadModelSnapshot struct {
	info      services.SimulationInfo
	airbases  map[string]services.Airbase
	aircrafts map[string]services.Aircraft
	threats   map[string]services.Threat
}

func captureReadModelSnapshots(
	t *testing.T,
	svc *services.SimulationService,
	simulationIDs ...string,
) map[string]simulationReadModelSnapshot {
	t.Helper()

	listedByID := simulationInfoIndexByID(t, svc.Simulations())
	require.Len(t, listedByID, len(simulationIDs))

	snapshots := make(map[string]simulationReadModelSnapshot, len(simulationIDs))
	for _, simulationID := range simulationIDs {
		info, err := svc.Simulation(simulationID)
		require.NoError(t, err)

		listed, ok := listedByID[simulationID]
		require.Truef(t, ok, "missing simulation %q in Simulations()", simulationID)
		require.Equal(t, info, listed)

		airbases, err := svc.Airbases(simulationID)
		require.NoError(t, err)

		aircrafts, err := svc.Aircrafts(simulationID)
		require.NoError(t, err)

		threats, err := svc.Threats(simulationID)
		require.NoError(t, err)

		snapshots[simulationID] = simulationReadModelSnapshot{
			info:      info,
			airbases:  airbasesByID(t, airbases),
			aircrafts: aircraftsByTailNumber(t, aircrafts),
			threats:   threatsByID(t, threats),
		}
	}

	return snapshots
}

func simulationInfoIndexByID(t *testing.T, infos []services.SimulationInfo) map[string]services.SimulationInfo {
	t.Helper()

	indexed := make(map[string]services.SimulationInfo, len(infos))
	for _, info := range infos {
		_, exists := indexed[info.ID]
		require.Falsef(t, exists, "duplicate simulation %q in Simulations()", info.ID)
		indexed[info.ID] = info
	}

	return indexed
}

func airbasesByID(t *testing.T, airbases []services.Airbase) map[string]services.Airbase {
	t.Helper()

	indexed := make(map[string]services.Airbase, len(airbases))
	for _, airbase := range airbases {
		_, exists := indexed[airbase.ID]
		require.Falsef(t, exists, "duplicate airbase %q in read model", airbase.ID)
		indexed[airbase.ID] = airbase
	}

	return indexed
}

func aircraftsByTailNumber(t *testing.T, aircrafts []services.Aircraft) map[string]services.Aircraft {
	t.Helper()

	indexed := make(map[string]services.Aircraft, len(aircrafts))
	for _, aircraft := range aircrafts {
		_, exists := indexed[aircraft.TailNumber]
		require.Falsef(t, exists, "duplicate aircraft %q in read model", aircraft.TailNumber)
		indexed[aircraft.TailNumber] = aircraft
	}

	return indexed
}

func threatsByID(t *testing.T, threats []services.Threat) map[string]services.Threat {
	t.Helper()

	indexed := make(map[string]services.Threat, len(threats))
	for _, threat := range threats {
		_, exists := indexed[threat.ID]
		require.Falsef(t, exists, "duplicate threat %q in read model", threat.ID)
		indexed[threat.ID] = threat
	}

	return indexed
}

func deterministicThreatLifecycleBaseConfig(seed [32]byte) services.BaseSimulationConfig {
	cfg := positionTrackingBaseConfig(seed, 1)
	cfg.Options.ThreatOpts.MaxActive = 1
	cfg.Options.ThreatOpts.MaxActiveTicks = 1
	return cfg
}

func requireOnlyAircraftTailNumber(t *testing.T, aircrafts map[string]services.Aircraft) string {
	t.Helper()

	require.Len(t, aircrafts, 1)
	for tailNumber := range aircrafts {
		return tailNumber
	}

	t.Fatal("expected exactly one aircraft tail number")
	return ""
}

func requireOnlyThreatSnapshot(t *testing.T, threats map[string]services.Threat) services.Threat {
	t.Helper()

	require.Len(t, threats, 1)
	for _, threat := range threats {
		return threat
	}

	t.Fatal("expected exactly one threat snapshot")
	return services.Threat{}
}

func extraThreats(baseThreats, branchThreats map[string]services.Threat) []services.Threat {
	extra := make([]services.Threat, 0, len(baseThreats)-len(branchThreats))
	for threatID, baseThreat := range baseThreats {
		if _, ok := branchThreats[threatID]; ok {
			continue
		}
		extra = append(extra, baseThreat)
	}

	return extra
}

const branchSummaryResolution = time.Second

type servicingSummaryExpectation struct {
	completedVisitCount int64
	totalDurationMs     int64
	averageDurationMs   *int64
}

func newBranchSummaryService() *services.SimulationService {
	return services.NewSimulationService(services.SimulationServiceConfig{Resolution: branchSummaryResolution})
}

func branchSummaryBaseConfig() services.BaseSimulationConfig {
	options := inboundOverrideSimulationOptions()
	options.ConstellationOpts.MinPerRegion = 1
	options.ConstellationOpts.MaxPerRegion = 1
	options.ConstellationOpts.MaxTotal = 1
	options.FleetOpts.AircraftMin = 2
	options.FleetOpts.AircraftMax = 2
	options.LifecycleOpts.Durations.InboundDecision = 12 * time.Second
	options.LifecycleOpts.Durations.Servicing = 5 * time.Second

	stateIndex := 0
	options.FleetOpts.StateFactory = func(_ *rand.Rand) simulation.AircraftState {
		stateIndex++
		if stateIndex == 1 {
			return &simulation.CommittedState{}
		}
		return &simulation.InboundState{}
	}

	return services.BaseSimulationConfig{Options: options}
}

func advanceBranchSummaryScenarioToFork(
	t *testing.T,
	svc *services.SimulationService,
	events *serviceEventWatcher,
) []services.AircraftStateChangeEvent {
	t.Helper()

	baseEvents := make([]services.AircraftStateChangeEvent, 0, 4)
	for step := 0; step < 20; step++ {
		require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
		baseEvents = append(baseEvents, collectAircraftStateChangeEvents(t, events, 10*time.Millisecond, services.BaseSimulationID)...)

		if branchSummaryForkReached(t, svc, baseEvents) {
			return baseEvents
		}
	}

	t.Fatalf("timed out reaching branch summary fork state; events=%#v", baseEvents)
	return nil
}

func advanceBranchSummaryScenarioToBranchOnlyCompletion(
	t *testing.T,
	svc *services.SimulationService,
	events *serviceEventWatcher,
	branchID string,
) []services.AircraftStateChangeEvent {
	t.Helper()

	branchEvents := make([]services.AircraftStateChangeEvent, 0, 4)
	for step := 0; step < 20; step++ {
		require.NoError(t, svc.StepSimulation(branchID))

		stepEvents := collectAircraftStateChangeEvents(t, events, 10*time.Millisecond, branchID)
		branchEvents = append(branchEvents, stepEvents...)
		for _, event := range stepEvents {
			if event.OldState == "Servicing" && event.NewState == "Ready" {
				return branchEvents
			}
		}
	}

	t.Fatalf("timed out reaching branch-only servicing completion for %q", branchID)
	return nil
}

func branchSummaryForkReached(
	t *testing.T,
	svc *services.SimulationService,
	baseEvents []services.AircraftStateChangeEvent,
) bool {
	t.Helper()

	summary := summarizeCompletedServicingVisitsFromEvents(baseEvents, branchSummaryResolution)
	if summary.completedVisitCount != 1 {
		return false
	}

	servicingExits := 0
	for _, event := range baseEvents {
		if event.OldState == "Servicing" {
			servicingExits++
		}
	}
	if servicingExits != 1 {
		return false
	}

	aircrafts, err := svc.Aircrafts(services.BaseSimulationID)
	require.NoError(t, err)

	readyCount := 0
	inboundCount := 0
	for _, aircraft := range aircrafts {
		switch aircraft.State {
		case "Ready":
			readyCount++
		case "Inbound":
			inboundCount++
		}
	}

	return readyCount == 1 && inboundCount == 1
}

func collectAircraftStateChangeEvents(
	t *testing.T,
	watcher *serviceEventWatcher,
	timeout time.Duration,
	simulationID string,
) []services.AircraftStateChangeEvent {
	t.Helper()

	collected := make([]services.AircraftStateChangeEvent, 0)
	for {
		event, ok := takePendingMatching(watcher, services.EventTypeAircraftStateChange, simulationID, asAircraftStateChangeEvent)
		if !ok {
			break
		}
		collected = append(collected, event)
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case raw, ok := <-watcher.events:
			if !ok {
				t.Fatalf("event stream closed while collecting aircraft_state_change events for simulation %q", simulationID)
			}

			event, typed := asAircraftStateChangeEvent(raw)
			if typed && event.EventType() == services.EventTypeAircraftStateChange && event.EventSimulationID() == simulationID {
				collected = append(collected, event)
				continue
			}

			watcher.pending = append(watcher.pending, raw)
		case <-timer.C:
			return collected
		}
	}
}

func summarizeBranchCompletedServicingVisitsAtSplit(
	baseEvents []services.AircraftStateChangeEvent,
	splitTimestamp time.Time,
	branchEvents []services.AircraftStateChangeEvent,
	resolution time.Duration,
) servicingSummaryExpectation {
	combined := make([]services.AircraftStateChangeEvent, 0, len(baseEvents)+len(branchEvents))
	for _, event := range baseEvents {
		if event.Timestamp.After(splitTimestamp) {
			continue
		}
		combined = append(combined, event)
	}
	combined = append(combined, branchEvents...)
	return summarizeCompletedServicingVisitsFromEvents(combined, resolution)
}

func summarizeCompletedServicingVisitsFromEvents(
	events []services.AircraftStateChangeEvent,
	resolution time.Duration,
) servicingSummaryExpectation {
	sortedEvents := append([]services.AircraftStateChangeEvent(nil), events...)
	sort.SliceStable(sortedEvents, func(i, j int) bool {
		if sortedEvents[i].Timestamp.Equal(sortedEvents[j].Timestamp) {
			if sortedEvents[i].TailNumber == sortedEvents[j].TailNumber {
				if sortedEvents[i].OldState == sortedEvents[j].OldState {
					return sortedEvents[i].NewState < sortedEvents[j].NewState
				}
				return sortedEvents[i].OldState < sortedEvents[j].OldState
			}
			return sortedEvents[i].TailNumber < sortedEvents[j].TailNumber
		}
		return sortedEvents[i].Timestamp.Before(sortedEvents[j].Timestamp)
	})

	activeServicingVisits := make(map[string]time.Time)
	var totalDurationMs int64
	var completedVisitCount int64

	for _, event := range sortedEvents {
		switch {
		case event.NewState == "Servicing":
			activeServicingVisits[event.TailNumber] = event.Timestamp.Add(resolution)
		case event.OldState == "Servicing":
			enteredAt, ok := activeServicingVisits[event.TailNumber]
			if !ok || !event.Timestamp.After(enteredAt) {
				delete(activeServicingVisits, event.TailNumber)
				continue
			}
			completedVisitCount++
			totalDurationMs += event.Timestamp.Sub(enteredAt).Milliseconds()
			delete(activeServicingVisits, event.TailNumber)
		}
	}

	summary := servicingSummaryExpectation{
		completedVisitCount: completedVisitCount,
		totalDurationMs:     totalDurationMs,
	}
	if completedVisitCount == 0 {
		return summary
	}

	averageDurationMs := totalDurationMs / completedVisitCount
	summary.averageDurationMs = &averageDurationMs
	return summary
}

func requireServicingSummaryExpectation(t *testing.T, actual servicingSummaryExpectation, expected servicingSummaryExpectation) {
	t.Helper()

	require.Equal(t, expected.completedVisitCount, actual.completedVisitCount)
	require.Equal(t, expected.totalDurationMs, actual.totalDurationMs)
	if expected.averageDurationMs == nil {
		require.Nil(t, actual.averageDurationMs)
		return
	}
	require.NotNil(t, actual.averageDurationMs)
	require.Equal(t, *expected.averageDurationMs, *actual.averageDurationMs)
}

func requireSimulationEndedSummaryContract(t *testing.T, ended services.SimulationEndedEvent, expected servicingSummaryExpectation) {
	t.Helper()

	summaryValue := reflect.ValueOf(ended).FieldByName("Summary")
	require.True(t, summaryValue.IsValid(), "SimulationEndedEvent should expose Summary")
	summaryValue = dereferenceValue(t, summaryValue, "Summary")
	require.Equal(t, expected.completedVisitCount, reflectedInt64Field(t, summaryValue, "CompletedVisitCount"))
	require.Equal(t, expected.totalDurationMs, reflectedInt64Field(t, summaryValue, "TotalDurationMs"))

	averageDurationMs := reflectedOptionalInt64Field(t, summaryValue, "AverageDurationMs")
	if expected.averageDurationMs == nil {
		require.Nil(t, averageDurationMs)
		return
	}
	require.NotNil(t, averageDurationMs)
	require.Equal(t, *expected.averageDurationMs, *averageDurationMs)
}

func dereferenceValue(t *testing.T, value reflect.Value, fieldName string) reflect.Value {
	t.Helper()

	for value.Kind() == reflect.Pointer {
		require.Falsef(t, value.IsNil(), "%s should not be nil", fieldName)
		value = value.Elem()
	}
	return value
}

func reflectedInt64Field(t *testing.T, value reflect.Value, fieldName string) int64 {
	t.Helper()

	field := value.FieldByName(fieldName)
	require.Truef(t, field.IsValid(), "summary should expose %s", fieldName)
	require.Truef(t, field.CanInt(), "summary field %s should be integer", fieldName)
	return field.Int()
}

func reflectedOptionalInt64Field(t *testing.T, value reflect.Value, fieldName string) *int64 {
	t.Helper()

	field := value.FieldByName(fieldName)
	require.Truef(t, field.IsValid(), "summary should expose %s", fieldName)
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			return nil
		}
		field = field.Elem()
	}
	require.Truef(t, field.CanInt(), "summary field %s should be optional integer", fieldName)
	fieldValue := field.Int()
	return &fieldValue
}

func ptr[T any](value T) *T {
	return &value
}
