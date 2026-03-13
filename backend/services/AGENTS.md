# SERVICES PACKAGE KNOWLEDGE BASE

## OVERVIEW
Simulation orchestration layer: owns base/branch lifecycle, runner coordination, event broadcasting, and API-facing read models.

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Main orchestration | `simulation.go` | `SimulationService`, branch/start/pause/resume/reset |
| Event payloads | `events.go` | Service-emitted event shapes |
| Subscriber behavior | `broadcaster.go` | Subscribe/unsubscribe, slow-client handling |
| Read models | `simulation_models.go` | API-facing data projections |
| Coverage | `simulation_test.go`, `simulation_e2e_test.go`, `events_test.go` | Branching + integration expectations |

## CONVENTIONS
- Base simulation ID is `base`; current branching support starts from base only.
- Service layer owns simulation IDs and injects them into outbound events.
- Branch creation clones simulation state; branch behavior must stay deterministic.
- Slow subscribers are dropped rather than allowed to stall simulation progress.
- Contract-facing changes should be mirrored in `backend/docs/context/api.md` and `backend/docs/context/frontend-live-events.md`.

## ANTI-PATTERNS
- Letting API handlers mutate simulation state directly.
- Blocking service progress on subscriber throughput.
- Changing event shapes or lifecycle semantics without test updates.
- Treating branch IDs or event payloads as incidental implementation details.

## COMMANDS
```bash
go test ./services
go test ./services -run TestSimulationService
```

## NOTES
- This package is where orchestration concerns belong; keep raw engine rules in `backend/simulation`.
