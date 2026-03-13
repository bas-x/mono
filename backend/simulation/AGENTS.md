# SIMULATION PACKAGE KNOWLEDGE BASE

## OVERVIEW
Deterministic event-driven simulation engine. Highest-risk package for replay, branching, lifecycle, and domain invariants.

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Engine construction | `simulation.go` | Simulator creation + top-level flow |
| Advancement / timing | `runner.go`, `time.go` | Runner semantics, tick bounds, pacing |
| Aircraft lifecycle | `state.go`, `lifecycle.go`, `ready.go`, `servicing.go`, `inbound.go`, `outbound.go`, `engaged.go`, `committed.go` | State transitions |
| Dispatch / assignments | `dispatcher.go`, `hooks.go` | Assignment hooks and emitted records |
| Threats / environment | `threat.go`, `environment.go`, `constellation.go` | World inputs |
| Safety net | `*_test.go` | Determinism, branch, runner, lifecycle expectations |

## CONVENTIONS
- Event-driven, not fixed-tick truth. UI ticks are derived views only.
- Time advances monotonically to scheduled event boundaries.
- Same-time conflicts require explicit stable ordering.
- Replay reconstruction must be deterministic and side-effect free.
- Tests are not optional decoration here; they encode engine invariants.

## ANTI-PATTERNS
- Wall-clock reads inside simulation decisions.
- Unseeded randomness or implicit tie-breaking.
- Hidden mutation that cannot be reconstructed from event history.
- Using replay/scrub state as writable source of truth.
- Performance shortcuts that break branch equivalence or determinism.

## COMMANDS
```bash
go test ./simulation
go test ./simulation -run TestSimulation
```

## NOTES
- Domain terminology should stay aligned with the root glossary.
- If behavior changes, update `docs/context/simulation-engine/*` and related replay docs.
