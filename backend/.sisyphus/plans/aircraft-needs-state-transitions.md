# Aircraft Needs Progression and Needs-Driven State Transitions

## TL;DR

> **Quick Summary**: Add deterministic need decay during outbound/engaged phases and make the outbound→inbound transition driven by time + needs together. Servicing fully resets needs.
> 
> **Deliverables**:
> - Need progression helpers in `simulation/need.go`
> - Updated state logic in `simulation/state.go` (OutboundState, EngagedState, ServicingState)
> - Aircraft need evolution in `simulation/aircraft.go`
> - TDD test coverage for all behavior changes
> 
> **Estimated Effort**: Medium
> **Parallel Execution**: YES - 2 waves
> **Critical Path**: Task 1 → Task 2 → Task 3 → Task 4 → Task 5 → Task 6

---

## Context

### Original Request
Continue from the completed simulation hooks, WebSocket events, and lifecycle work. Enhance aircraft state transitions and needs mechanics so that needs degrade over time and drive state transitions.

### Interview Summary
**Key Discussions**:
- All needs should change over time (not just fuel)
- Decay happens during outbound and engaged phases only
- Outbound → inbound uses time + needs together
- All needs contribute equally to the return decision
- Servicing fully resets all needs
- TDD approach

**Research Findings**:
- `simulation/state.go` uses fixed durations for Outbound/Engaged before unconditional transitions
- `simulation/need.go` has Need types, invariants, Clone but no progression helpers
- `simulation/aircraft.go` steps state, emits hooks, but does not evolve needs
- Existing test coverage in `simulation/simulation_test.go` and `simulation/simulation_e2e_test.go`

### Self-Review Gap Analysis
**Identified Gaps** (addressed):
- Decay rate not specified → Default: configurable per-step severity increase (default 5 per step)
- Threshold for return not specified → Default: configurable severity threshold (default 80)
- How time + needs combine → Default: EITHER time expires OR needs threshold triggers inbound
- Aircraft with no needs → Edge case: aircraft with empty needs never triggers needs-based return, relies on time only

---

## Work Objectives

### Core Objective
Make aircraft needs evolve deterministically during flight and use needs as a factor in the outbound→inbound state transition decision.

### Concrete Deliverables
- `simulation/need.go`: `Degrade()` method and `NeedsThresholdReached()` helper
- `simulation/state.go`: Updated `OutboundState.Step()` and `EngagedState.Step()` to degrade needs and check threshold
- `simulation/state.go`: Updated `ServicingState.Step()` to reset needs
- `simulation/aircraft.go`: Pass aircraft needs to state step (already has pointer access)
- TDD tests for each behavior

### Definition of Done
- [x] `go test ./simulation/...` passes with new tests covering need decay and threshold transitions
- [x] `go build ./...` succeeds with no errors
- [x] Existing `TestAircraftStateTransitions` still passes
- [x] New test demonstrates needs-based early return from outbound/engaged
- [x] New test demonstrates servicing resets needs

### Must Have
- Determinism preserved (no randomness without seed)
- Need decay happens per-step during outbound and engaged
- Return decision uses time OR needs threshold
- Servicing resets all needs to severity 0
- All existing tests pass

### Must NOT Have (Guardrails)
- No changes to WebSocket/event/service layer unless strictly required
- No changes to dispatcher or landing assignment logic
- No new dependencies
- No wall-clock reads (`time.Now()`) in simulation logic
- No breaking changes to existing public API
- No over-abstraction (keep decay logic inline in state Step methods)

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES
- **Automated tests**: TDD
- **Framework**: Go standard `testing` with `testify/require`
- **If TDD**: Each task follows RED (failing test) → GREEN (minimal impl) → REFACTOR

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Simulation logic**: Use Bash (`go test -v -run TestName`) — Run specific tests, capture output
- **Build verification**: Use Bash (`go build ./...`) — Verify compilation

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation — can run in parallel):
├── Task 1: TDD Need.Degrade() method [quick]
├── Task 2: TDD NeedsThresholdReached() helper [quick]
└── Task 3: TDD ServicingState resets needs [quick]

Wave 2 (Integration — depends on Wave 1):
├── Task 4: TDD OutboundState needs decay + threshold check [deep]
├── Task 5: TDD EngagedState needs decay + threshold check [deep]
└── Task 6: E2E test needs-driven full cycle [deep]

Wave FINAL (Verification — after all tasks):
├── Task F1: Run full test suite and verify all pass [quick]
└── Task F2: Scope fidelity check [quick]

Critical Path: Task 1 → Task 4 → Task 6 → F1
Parallel Speedup: ~40% faster than sequential
Max Concurrent: 3 (Wave 1)
```

### Dependency Matrix

| Task | Depends On | Blocks |
|------|------------|--------|
| 1 | — | 4, 5 |
| 2 | — | 4, 5 |
| 3 | — | 6 |
| 4 | 1, 2 | 6 |
| 5 | 1, 2 | 6 |
| 6 | 3, 4, 5 | F1, F2 |
| F1 | 6 | — |
| F2 | 6 | — |

### Agent Dispatch Summary

- **Wave 1**: 3 tasks — T1 → `quick`, T2 → `quick`, T3 → `quick`
- **Wave 2**: 3 tasks — T4 → `deep`, T5 → `deep`, T6 → `deep`
- **Wave FINAL**: 2 tasks — F1 → `quick`, F2 → `quick`

---

## TODOs

- [x] 1. TDD Need.Degrade() Method

  **What to do**:
  - Write failing test `TestNeed_Degrade` that verifies:
    - Calling `Degrade(amount)` increases Severity by amount
    - Severity is capped at 100 (never exceeds)
    - Degrade with 0 is a no-op
  - Implement `func (n *Need) Degrade(amount int)` in `simulation/need.go`
  - Run test to verify GREEN

  **Must NOT do**:
  - Do not add randomness
  - Do not modify other Need methods

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3)
  - **Blocks**: Tasks 4, 5
  - **Blocked By**: None

  **References**:
  - `simulation/need.go:39-44` — Need struct definition with Severity field
  - `simulation/need.go:82-89` — AssertInvariants shows Severity must be 0-100
  - `simulation/need.go:92-99` — Clone pattern to follow

  **Acceptance Criteria**:
  - [ ] Test file: `simulation/need_test.go` exists with `TestNeed_Degrade`
  - [ ] `go test ./simulation/... -run TestNeed_Degrade` → PASS

  **QA Scenarios**:
  ```
  Scenario: Need degradation increases severity
    Tool: Bash (go test)
    Preconditions: simulation/need_test.go contains TestNeed_Degrade
    Steps:
      1. Run: go test ./simulation/... -v -run TestNeed_Degrade
      2. Assert: output contains "PASS"
      3. Assert: output contains "TestNeed_Degrade"
    Expected Result: Test passes, severity increases correctly
    Failure Indicators: "FAIL" in output, test not found
    Evidence: .sisyphus/evidence/task-1-need-degrade-test.txt

  Scenario: Severity cap at 100
    Tool: Bash (go test)
    Preconditions: Test includes case where Degrade would exceed 100
    Steps:
      1. Run: go test ./simulation/... -v -run TestNeed_Degrade
      2. Assert: test covers severity capping
    Expected Result: Severity never exceeds 100
    Evidence: .sisyphus/evidence/task-1-need-degrade-cap.txt
  ```

  **Commit**: YES
  - Message: `feat(simulation): add Need.Degrade method`
  - Files: `simulation/need.go`, `simulation/need_test.go`
  - Pre-commit: `go test ./simulation/... -run TestNeed_Degrade`

- [x] 2. TDD NeedsThresholdReached() Helper

  **What to do**:
  - Write failing test `TestNeedsThresholdReached` that verifies:
    - Returns true if ANY need has Severity >= threshold
    - Returns false if all needs below threshold
    - Returns false for empty needs slice
  - Implement `func NeedsThresholdReached(needs []Need, threshold int) bool` in `simulation/need.go`
  - Run test to verify GREEN

  **Must NOT do**:
  - Do not weight needs differently (equal weighting confirmed)
  - Do not add aircraft dependency

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3)
  - **Blocks**: Tasks 4, 5
  - **Blocked By**: None

  **References**:
  - `simulation/need.go:39-44` — Need struct with Severity field
  - `simulation/need.go:24-37` — AllNeedTypes for reference

  **Acceptance Criteria**:
  - [ ] Test in `simulation/need_test.go` with `TestNeedsThresholdReached`
  - [ ] `go test ./simulation/... -run TestNeedsThresholdReached` → PASS

  **QA Scenarios**:
  ```
  Scenario: Threshold detection works
    Tool: Bash (go test)
    Preconditions: simulation/need_test.go contains TestNeedsThresholdReached
    Steps:
      1. Run: go test ./simulation/... -v -run TestNeedsThresholdReached
      2. Assert: output contains "PASS"
    Expected Result: Function correctly detects when any need exceeds threshold
    Failure Indicators: "FAIL" in output
    Evidence: .sisyphus/evidence/task-2-threshold-test.txt

  Scenario: Empty needs returns false
    Tool: Bash (go test)
    Preconditions: Test includes empty slice case
    Steps:
      1. Run: go test ./simulation/... -v -run TestNeedsThresholdReached
      2. Assert: empty needs case covered and returns false
    Expected Result: Empty needs slice returns false
    Evidence: .sisyphus/evidence/task-2-threshold-empty.txt
  ```

  **Commit**: YES
  - Message: `feat(simulation): add NeedsThresholdReached helper`
  - Files: `simulation/need.go`, `simulation/need_test.go`
  - Pre-commit: `go test ./simulation/... -run TestNeedsThresholdReached`

- [x] 3. TDD ServicingState Resets Needs

  **What to do**:
  - Write failing test `TestServicingState_ResetsNeeds` that verifies:
    - When aircraft enters ServicingState with degraded needs
    - After servicing duration completes
    - All needs have Severity reset to 0
  - Update `ServicingState.Step()` in `simulation/state.go` to reset needs when transitioning to Ready
  - Add `func (a *Aircraft) ResetNeeds()` helper if needed
  - Run test to verify GREEN

  **Must NOT do**:
  - Do not change servicing duration
  - Do not add new state types

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2)
  - **Blocks**: Task 6
  - **Blocked By**: None

  **References**:
  - `simulation/state.go:174-200` — ServicingState struct and Step method
  - `simulation/state.go:22` — servicingDuration constant (6s)
  - `simulation/aircraft.go:7-13` — Aircraft struct with Needs field
  - `simulation/simulation_test.go:173-228` — TestAircraftStateTransitions pattern

  **Acceptance Criteria**:
  - [ ] Test in `simulation/state_test.go` with `TestServicingState_ResetsNeeds`
  - [ ] `go test ./simulation/... -run TestServicingState_ResetsNeeds` → PASS

  **QA Scenarios**:
  ```
  Scenario: Servicing resets all needs to zero
    Tool: Bash (go test)
    Preconditions: simulation/state_test.go contains TestServicingState_ResetsNeeds
    Steps:
      1. Run: go test ./simulation/... -v -run TestServicingState_ResetsNeeds
      2. Assert: output contains "PASS"
    Expected Result: After servicing completes, all need severities are 0
    Failure Indicators: "FAIL" in output
    Evidence: .sisyphus/evidence/task-3-servicing-reset.txt

  Scenario: Needs with various severities all reset
    Tool: Bash (go test)
    Preconditions: Test uses aircraft with multiple needs at different severities
    Steps:
      1. Run: go test ./simulation/... -v -run TestServicingState_ResetsNeeds
      2. Assert: test verifies all needs reset, not just one
    Expected Result: Every need on aircraft has Severity 0 after servicing
    Evidence: .sisyphus/evidence/task-3-servicing-all-needs.txt
  ```

  **Commit**: YES
  - Message: `feat(simulation): reset needs in ServicingState`
  - Files: `simulation/state.go`, `simulation/state_test.go`, possibly `simulation/aircraft.go`
  - Pre-commit: `go test ./simulation/... -run TestServicingState_ResetsNeeds`

- [x] 4. TDD OutboundState Needs Decay + Threshold Check

  **What to do**:
  - Write failing test `TestOutboundState_DegradeNeeds` that verifies:
    - Each step in OutboundState calls Degrade on all aircraft needs
    - Default decay amount per step (use constant, e.g., `needDecayPerStep = 5`)
  - Write failing test `TestOutboundState_EarlyReturnOnThreshold` that verifies:
    - If NeedsThresholdReached returns true, transition to InboundState immediately
    - Default threshold (use constant, e.g., `needsReturnThreshold = 80`)
  - Update `OutboundState.Step()` to:
    1. Degrade all aircraft needs each step
    2. Check threshold and return InboundState early if reached
    3. Otherwise continue existing time-based logic
  - Run tests to verify GREEN

  **Must NOT do**:
  - Do not remove time-based transition (time OR needs triggers return)
  - Do not change outboundDuration constant
  - Do not modify hook emission logic

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES (within Wave 2)
  - **Parallel Group**: Wave 2 (with Task 5)
  - **Blocks**: Task 6
  - **Blocked By**: Tasks 1, 2

  **References**:
  - `simulation/state.go:48-74` — OutboundState struct and Step method
  - `simulation/state.go:17-23` — Duration constants
  - `simulation/state.go:5-9` — FlightContext with Clock
  - `simulation/aircraft.go:7-13` — Aircraft struct with Needs
  - `simulation/simulation_test.go:206-210` — How to step simulation and check state

  **Acceptance Criteria**:
  - [ ] Test `TestOutboundState_DegradeNeeds` in `simulation/state_test.go`
  - [ ] Test `TestOutboundState_EarlyReturnOnThreshold` in `simulation/state_test.go`
  - [ ] `go test ./simulation/... -run TestOutboundState` → PASS
  - [ ] Existing `TestAircraftStateTransitions` still passes

  **QA Scenarios**:
  ```
  Scenario: Needs degrade each outbound step
    Tool: Bash (go test)
    Preconditions: simulation/state_test.go contains TestOutboundState_DegradeNeeds
    Steps:
      1. Run: go test ./simulation/... -v -run TestOutboundState_DegradeNeeds
      2. Assert: output contains "PASS"
    Expected Result: After N steps, need severity increased by N * decayAmount
    Failure Indicators: "FAIL" in output
    Evidence: .sisyphus/evidence/task-4-outbound-decay.txt

  Scenario: Early return when threshold reached
    Tool: Bash (go test)
    Preconditions: simulation/state_test.go contains TestOutboundState_EarlyReturnOnThreshold
    Steps:
      1. Run: go test ./simulation/... -v -run TestOutboundState_EarlyReturnOnThreshold
      2. Assert: output contains "PASS"
    Expected Result: Aircraft transitions to Inbound before time expires when needs critical
    Failure Indicators: "FAIL" in output
    Evidence: .sisyphus/evidence/task-4-outbound-threshold.txt

  Scenario: Existing state transitions still work
    Tool: Bash (go test)
    Preconditions: No regressions introduced
    Steps:
      1. Run: go test ./simulation/... -v -run TestAircraftStateTransitions
      2. Assert: output contains "PASS"
    Expected Result: Time-based transitions still function
    Evidence: .sisyphus/evidence/task-4-regression.txt
  ```

  **Commit**: YES
  - Message: `feat(simulation): degrade needs and check threshold in OutboundState`
  - Files: `simulation/state.go`, `simulation/state_test.go`
  - Pre-commit: `go test ./simulation/... -run "TestOutboundState|TestAircraftStateTransitions"`

- [x] 5. TDD EngagedState Needs Decay + Threshold Check

  **What to do**:
  - Write failing test `TestEngagedState_DegradeNeeds` that verifies:
    - Each step in EngagedState calls Degrade on all aircraft needs
    - Uses same decay rate as OutboundState
  - Write failing test `TestEngagedState_EarlyReturnOnThreshold` that verifies:
    - If NeedsThresholdReached returns true, transition to InboundState immediately
    - Uses same threshold as OutboundState
  - Update `EngagedState.Step()` to:
    1. Degrade all aircraft needs each step
    2. Check threshold and return InboundState early if reached
    3. Otherwise continue existing time-based logic
  - Run tests to verify GREEN

  **Must NOT do**:
  - Do not remove time-based transition
  - Do not change engagedDuration constant
  - Do not duplicate constants (reuse from OutboundState task)

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES (within Wave 2)
  - **Parallel Group**: Wave 2 (with Task 4)
  - **Blocks**: Task 6
  - **Blocked By**: Tasks 1, 2

  **References**:
  - `simulation/state.go:76-102` — EngagedState struct and Step method
  - `simulation/state.go:17-23` — Duration constants
  - Task 4 implementation — Reuse decay/threshold constants

  **Acceptance Criteria**:
  - [ ] Test `TestEngagedState_DegradeNeeds` in `simulation/state_test.go`
  - [ ] Test `TestEngagedState_EarlyReturnOnThreshold` in `simulation/state_test.go`
  - [ ] `go test ./simulation/... -run TestEngagedState` → PASS

  **QA Scenarios**:
  ```
  Scenario: Needs degrade each engaged step
    Tool: Bash (go test)
    Preconditions: simulation/state_test.go contains TestEngagedState_DegradeNeeds
    Steps:
      1. Run: go test ./simulation/... -v -run TestEngagedState_DegradeNeeds
      2. Assert: output contains "PASS"
    Expected Result: Need severity increases during engaged phase
    Failure Indicators: "FAIL" in output
    Evidence: .sisyphus/evidence/task-5-engaged-decay.txt

  Scenario: Early return when threshold reached in engaged
    Tool: Bash (go test)
    Preconditions: simulation/state_test.go contains TestEngagedState_EarlyReturnOnThreshold
    Steps:
      1. Run: go test ./simulation/... -v -run TestEngagedState_EarlyReturnOnThreshold
      2. Assert: output contains "PASS"
    Expected Result: Aircraft transitions to Inbound when needs become critical during engagement
    Failure Indicators: "FAIL" in output
    Evidence: .sisyphus/evidence/task-5-engaged-threshold.txt
  ```

  **Commit**: YES
  - Message: `feat(simulation): degrade needs and check threshold in EngagedState`
  - Files: `simulation/state.go`, `simulation/state_test.go`
  - Pre-commit: `go test ./simulation/... -run TestEngagedState`

- [x] 6. E2E Test Needs-Driven Full Cycle

  **What to do**:
  - Write E2E test `TestSimulation_NeedsDrivenStateTransitions` that verifies:
    - Aircraft starts Outbound with moderate need severity (e.g., 60)
    - Needs degrade during outbound and engaged
    - Aircraft returns to base early when needs reach threshold
    - After servicing, needs are reset to 0
    - Aircraft can sortie again with fresh needs
  - This test should use the full simulation (like TestAircraftStateTransitions)
  - Verify determinism: same seed produces same behavior

  **Must NOT do**:
  - Do not modify existing E2E tests
  - Do not test WebSocket/service layer

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential (after Wave 2)
  - **Blocks**: F1, F2
  - **Blocked By**: Tasks 3, 4, 5

  **References**:
  - `simulation/simulation_test.go:173-228` — TestAircraftStateTransitions pattern
  - `simulation/simulation_e2e_test.go:12-72` — E2E test structure
  - `simulation/hooks_test.go:81-91` — newHookTestSimulation helper

  **Acceptance Criteria**:
  - [ ] Test `TestSimulation_NeedsDrivenStateTransitions` in `simulation/simulation_test.go`
  - [ ] `go test ./simulation/... -run TestSimulation_NeedsDrivenStateTransitions` → PASS
  - [ ] Test demonstrates needs-based early return
  - [ ] Test demonstrates servicing reset

  **QA Scenarios**:
  ```
  Scenario: Full needs-driven cycle works end-to-end
    Tool: Bash (go test)
    Preconditions: simulation/simulation_test.go contains TestSimulation_NeedsDrivenStateTransitions
    Steps:
      1. Run: go test ./simulation/... -v -run TestSimulation_NeedsDrivenStateTransitions
      2. Assert: output contains "PASS"
    Expected Result: Aircraft completes full cycle with needs-driven transitions
    Failure Indicators: "FAIL" in output
    Evidence: .sisyphus/evidence/task-6-e2e-needs.txt

  Scenario: Determinism preserved
    Tool: Bash (go test)
    Preconditions: Test runs twice with same seed
    Steps:
      1. Run: go test ./simulation/... -v -run TestSimulation_NeedsDrivenStateTransitions -count=2
      2. Assert: both runs produce identical results
    Expected Result: Same seed = same outcome
    Evidence: .sisyphus/evidence/task-6-determinism.txt

  Scenario: All simulation tests still pass
    Tool: Bash (go test)
    Preconditions: No regressions from changes
    Steps:
      1. Run: go test ./simulation/... -v
      2. Assert: output shows all tests PASS
    Expected Result: Zero test failures
    Failure Indicators: Any "FAIL" in output
    Evidence: .sisyphus/evidence/task-6-full-suite.txt
  ```

  **Commit**: YES
  - Message: `test(simulation): add E2E test for needs-driven state cycle`
  - Files: `simulation/simulation_test.go`
  - Pre-commit: `go test ./simulation/...`

---

## Final Verification Wave

- [x] F1. **Run Full Test Suite** — `quick`
  Run `go test ./...` and `go build ./...`. Verify all tests pass and build succeeds.
  Output: `Tests [N pass/N fail] | Build [PASS/FAIL] | VERDICT`

- [x] F2. **Scope Fidelity Check** — `quick`
  Verify all tasks completed as specified. Check no changes leaked into services/api/WebSocket layers. Verify Must NOT Have guardrails respected.
  Output: `Tasks [N/N compliant] | Guardrails [CLEAN/N issues] | VERDICT`

---

## Commit Strategy

| Task | Commit |
|------|--------|
| 1 | `feat(simulation): add Need.Degrade method` |
| 2 | `feat(simulation): add NeedsThresholdReached helper` |
| 3 | `feat(simulation): reset needs in ServicingState` |
| 4 | `feat(simulation): degrade needs and check threshold in OutboundState` |
| 5 | `feat(simulation): degrade needs and check threshold in EngagedState` |
| 6 | `test(simulation): add E2E test for needs-driven state cycle` |

---

## Success Criteria

### Verification Commands
```bash
go test ./simulation/... -v  # Expected: all tests pass including new TDD tests
go build ./...               # Expected: no errors
```

### Final Checklist
- [x] All "Must Have" present
- [x] All "Must NOT Have" absent
- [x] All tests pass
- [x] Determinism preserved
